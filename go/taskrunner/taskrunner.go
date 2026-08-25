// Package taskrunner 提供"用账号队列并发处理内容"的通用任务引擎：多线程并发处理
// 账号池，单个账号可循环处理多次，连续失败达到阈值换号，整个账号池可循环多轮，
// 支持暂停/恢复与取消，通过回调把进度事件传给上层。具体处理协议（Requester）与
// 处理对象类型（Item，比如文章、视频、评论文本）由各消费方注入，本包不关心。
package taskrunner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/japansms40-web/dfclientkit/go/account"
)

// Requester 执行一次"用某账号处理某个 Item"的请求，返回结果描述或错误。
type Requester[Item any] interface {
	Publish(ctx context.Context, acc account.Account, item Item) (string, error)
}

// RepoCreator 在处理某个账号前视需要建一个"仓库/空间"。
type RepoCreator interface {
	CreateSpace(ctx context.Context, acc account.Account) error
}

// EventKind 标识一次处理过程中的事件类型。
type EventKind int

const (
	EventAttemptStart   EventKind = iota // 某账号的一次处理尝试开始
	EventAttemptSuccess                  // 该次尝试成功
	EventAttemptFailure                  // 该次尝试失败
	EventAccountSwitch                   // 放弃当前账号，换下一个（达到每号处理上限/连续失败换号/建仓库失败）
	EventRoundStart                      // 新的一轮开始
	EventRoundProgress                   // 本轮进度更新
	EventRoundDone                       // 本轮结束
)

// Event 是回传给上层的进度事件。
type Event struct {
	Kind         EventKind
	AccountIndex int // 账号在原始队列中的下标；EventRoundStart/RoundProgress/RoundDone 不适用
	CK           string
	ItemLabel    string // 本次处理对象的展示文案（比如文章标题），由调用方通过 itemLabel 提供
	Result       string // 成功时 Requester 返回的结果描述
	Err          error  // 失败/换号时的原因
	Round        int
	RoundTotal   int
	RoundDone    int
}

// IndexedAccount 携带账号在原始队列中的下标，用于事件回传时定位前端要更新的行。
type IndexedAccount struct {
	Index   int
	Account account.Account
}

// RunConfig 是一次批量任务的运行参数。json tag 显式给出是因为消费方通常会把
// 它内嵌进自己的 Config 结构体再整体序列化（比如通过 Wails 传给前端）——匿名嵌入
// 字段没有 tag 时，encoding/json 会退化成用 Go 字段名（首字母大写）当 JSON key，
// 和前端约定的 camelCase 字段名对不上。
type RunConfig struct {
	Threads          int  `json:"threads"`          // 并发线程数
	IntervalSec      int  `json:"intervalSec"`      // 同一账号相邻两次处理尝试之间的等待秒数
	PerAccountCount  int  `json:"perAccountCount"`  // 单个账号最多处理多少次
	FailSwitchCount  int  `json:"failSwitchCount"`  // 账号连续失败达到此次数就换号
	CycleRounds      int  `json:"cycleRounds"`      // 账号池整体循环轮数
	RoundIntervalSec int  `json:"roundIntervalSec"` // 相邻两轮之间的等待秒数
	CreateRepo       bool `json:"createRepo"`       // 处理账号前是否先建仓库/空间
}

// Normalize 纠正非法数值，供消费方在保存/发起任务前调用，也在 Run 内部兜底调用。
func (c *RunConfig) Normalize() {
	if c.Threads < 1 {
		c.Threads = 1
	}
	if c.IntervalSec < 0 {
		c.IntervalSec = 0
	}
	if c.PerAccountCount < 1 {
		c.PerAccountCount = 1
	}
	if c.FailSwitchCount < 1 {
		c.FailSwitchCount = 1
	}
	if c.CycleRounds < 1 {
		c.CycleRounds = 1
	}
	if c.RoundIntervalSec < 0 {
		c.RoundIntervalSec = 0
	}
}

// PauseGate 是可在运行中途暂停/恢复批量任务的开关，多个 worker 共用同一个实例。
type PauseGate struct {
	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
}

// NewPauseGate 创建一个初始为"未暂停"的开关。
func NewPauseGate() *PauseGate {
	g := &PauseGate{}
	g.cond = sync.NewCond(&g.mu)
	return g
}

// Pause 暂停：worker 会在完成当前尝试后阻塞，直到 Resume 或 ctx 取消。
func (g *PauseGate) Pause() {
	g.mu.Lock()
	g.paused = true
	g.mu.Unlock()
}

// Resume 恢复运行。
func (g *PauseGate) Resume() {
	g.mu.Lock()
	g.paused = false
	g.mu.Unlock()
	g.cond.Broadcast()
}

// IsPaused 返回当前是否处于暂停状态。
func (g *PauseGate) IsPaused() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.paused
}

// wakeAll 唤醒所有等待者（不改变暂停状态），用于 ctx 取消时让 Wait 尽快返回。
func (g *PauseGate) wakeAll() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cond.Broadcast()
}

// wait 在暂停期间阻塞；ctx 取消时立即返回 ctx.Err()。
func (g *PauseGate) wait(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for g.paused {
		if err := ctx.Err(); err != nil {
			return err
		}
		g.cond.Wait()
	}
	return ctx.Err()
}

// Runner 并发执行账号处理任务，Item 是被处理对象的类型（文章、视频……由消费方决定）。
type Runner[Item any] struct {
	client Requester[Item]
	repo   RepoCreator
}

// New 创建 Runner；repo 为 nil 时忽略"创建仓库"选项。
func New[Item any](client Requester[Item], repo RepoCreator) *Runner[Item] {
	return &Runner[Item]{client: client, repo: repo}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// Run 执行账号池的批量处理任务。gate 为 nil 时视为从不暂停。items 为空时直接返回。
// itemLabel 从一个 Item 里提取用于事件展示的文案（比如文章标题）。
func (r *Runner[Item]) Run(ctx context.Context, cfg RunConfig, gate *PauseGate, pool []IndexedAccount, items []Item, itemLabel func(Item) string, onEvent func(Event)) error {
	if len(items) == 0 || len(pool) == 0 {
		return nil
	}
	cfg.Normalize()
	if gate == nil {
		gate = NewPauseGate()
	}

	stopWake := make(chan struct{})
	defer close(stopWake)
	go func() {
		select {
		case <-ctx.Done():
			gate.wakeAll()
		case <-stopWake:
		}
	}()

	for round := 1; round <= cfg.CycleRounds; round++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		onEvent(Event{Kind: EventRoundStart, Round: round, RoundTotal: len(pool)})

		work := make(chan IndexedAccount, len(pool))
		for _, ia := range pool {
			work <- ia
		}
		close(work)

		var wg sync.WaitGroup
		var doneCount int32
		for t := 0; t < cfg.Threads; t++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for ia := range work {
					if ctx.Err() != nil {
						return
					}
					r.runAccount(ctx, cfg, gate, ia, items, itemLabel, onEvent)
					n := atomic.AddInt32(&doneCount, 1)
					onEvent(Event{Kind: EventRoundProgress, Round: round, RoundDone: int(n), RoundTotal: len(pool)})
				}
			}()
		}
		wg.Wait()

		if err := ctx.Err(); err != nil {
			return err
		}
		onEvent(Event{Kind: EventRoundDone, Round: round, RoundTotal: len(pool)})

		if round < cfg.CycleRounds {
			if err := sleepCtx(ctx, time.Duration(cfg.RoundIntervalSec)*time.Second); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Runner[Item]) runAccount(ctx context.Context, cfg RunConfig, gate *PauseGate, ia IndexedAccount, items []Item, itemLabel func(Item) string, onEvent func(Event)) {
	if cfg.CreateRepo && r.repo != nil {
		if err := r.repo.CreateSpace(ctx, ia.Account); err != nil {
			onEvent(Event{Kind: EventAccountSwitch, AccountIndex: ia.Index, CK: ia.Account.CK, Err: err})
			return
		}
	}

	consecFail := 0
	for i := 0; i < cfg.PerAccountCount; i++ {
		if err := gate.wait(ctx); err != nil {
			return
		}
		if err := ctx.Err(); err != nil {
			return
		}

		item := items[i%len(items)]
		label := itemLabel(item)
		onEvent(Event{Kind: EventAttemptStart, AccountIndex: ia.Index, CK: ia.Account.CK, ItemLabel: label})

		result, err := r.client.Publish(ctx, ia.Account, item)
		if err != nil {
			consecFail++
			onEvent(Event{Kind: EventAttemptFailure, AccountIndex: ia.Index, CK: ia.Account.CK, ItemLabel: label, Err: err})
			if consecFail >= cfg.FailSwitchCount {
				onEvent(Event{Kind: EventAccountSwitch, AccountIndex: ia.Index, CK: ia.Account.CK, Err: errors.New("连续失败达到换号阈值")})
				return
			}
		} else {
			consecFail = 0
			onEvent(Event{Kind: EventAttemptSuccess, AccountIndex: ia.Index, CK: ia.Account.CK, ItemLabel: label, Result: result})
		}

		if i < cfg.PerAccountCount-1 && cfg.IntervalSec > 0 {
			if err := sleepCtx(ctx, time.Duration(cfg.IntervalSec)*time.Second); err != nil {
				return
			}
		}
	}
}
