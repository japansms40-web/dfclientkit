package taskrunner

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/japansms40-web/dfclientkit/go/account"
)

// 回归测试：RunConfig 经常被消费方匿名嵌入进自己的 Config 结构体再整体序列化
// （比如通过 Wails 传给前端）。如果字段丢了 json tag，encoding/json 会退化成用
// Go 字段名（首字母大写）当 key，前端按 camelCase 读会全部拿到 undefined——这个
// bug 在 githubbaidu 迁移到 dfclientkit 时真实发生过一次，这里锁死不能再犯。
func TestRunConfigJSONFieldNamesAreCamelCase(t *testing.T) {
	cfg := RunConfig{
		Threads: 1, IntervalSec: 2, PerAccountCount: 3,
		FailSwitchCount: 4, CycleRounds: 5, RoundIntervalSec: 6, CreateRepo: true,
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	for _, key := range []string{
		"threads", "intervalSec", "perAccountCount",
		"failSwitchCount", "cycleRounds", "roundIntervalSec", "createRepo",
	} {
		if _, ok := m[key]; !ok {
			t.Errorf("序列化结果缺少 camelCase 字段 %q，实际是 %s", key, b)
		}
	}
}

type testItem struct {
	Title string
}

func itemTitle(i testItem) string { return i.Title }

// fakeRequester 实现 Requester[testItem] 接口，用于测试。alwaysFail 为 true 时永远失败。
type fakeRequester struct {
	mu         sync.Mutex
	alwaysFail bool
	calls      map[string]int
}

func (f *fakeRequester) Publish(ctx context.Context, acc account.Account, item testItem) (string, error) {
	f.mu.Lock()
	if f.calls == nil {
		f.calls = map[string]int{}
	}
	f.calls[acc.CK]++
	f.mu.Unlock()
	if f.alwaysFail {
		return "", errors.New("boom")
	}
	return "ok:" + acc.CK, nil
}

func (f *fakeRequester) callCount(ck string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[ck]
}

func testTargets(cks ...string) []IndexedAccount {
	out := make([]IndexedAccount, len(cks))
	for i, ck := range cks {
		out[i] = IndexedAccount{Index: i, Account: account.Account{CK: ck}}
	}
	return out
}

func collectEvents(t *testing.T, r *Runner[testItem], cfg RunConfig, gate *PauseGate, pool []IndexedAccount, items []testItem) []Event {
	t.Helper()
	var mu sync.Mutex
	var events []Event
	err := r.Run(context.Background(), cfg, gate, pool, items, itemTitle, func(e Event) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Run err: %v", err)
	}
	return events
}

func TestRun_SingleAttemptAllSuccess(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1", "ck2")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	var success int
	for _, e := range events {
		if e.Kind == EventAttemptSuccess {
			success++
		}
	}
	if success != 2 {
		t.Errorf("成功次数 = %d, want 2", success)
	}
}

func TestRun_PerAccountLoopsMultipleTimes(t *testing.T) {
	items := []testItem{{Title: "a"}, {Title: "b"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 3, FailSwitchCount: 10, CycleRounds: 1}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	var titles []string
	for _, e := range events {
		if e.Kind == EventAttemptStart {
			titles = append(titles, e.ItemLabel)
		}
	}
	want := []string{"a", "b", "a"} // 循环复用内容
	if len(titles) != len(want) {
		t.Fatalf("尝试次数 = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Errorf("titles[%d] = %q, want %q", i, titles[i], want[i])
		}
	}
}

func TestRun_FailSwitchStopsAccountEarly(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 10, FailSwitchCount: 2, CycleRounds: 1}
	fr := &fakeRequester{alwaysFail: true}
	events := collectEvents(t, New[testItem](fr, nil), cfg, nil, targets, items)

	if got := fr.callCount("ck1"); got != 2 {
		t.Errorf("应只尝试 2 次就换号, 实际调用 %d 次", got)
	}
	var switched int
	for _, e := range events {
		if e.Kind == EventAccountSwitch {
			switched++
		}
	}
	if switched != 1 {
		t.Errorf("应有一次换号事件, got %d", switched)
	}
}

func TestRun_FailThenSuccessResetsCounter(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 4, FailSwitchCount: 2, CycleRounds: 1}

	calls := 0
	var mu sync.Mutex
	requester := requesterFunc(func(ctx context.Context, acc account.Account, item testItem) (string, error) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 2 { // 第 2 次失败，其余成功；失败不连续所以不会触发换号
			return "", errors.New("boom")
		}
		return "ok", nil
	})
	events := collectEvents(t, New[testItem](requester, nil), cfg, nil, targets, items)

	var switched int
	for _, e := range events {
		if e.Kind == EventAccountSwitch {
			switched++
		}
	}
	if switched != 0 {
		t.Errorf("非连续失败不应触发换号, got %d 次换号", switched)
	}
	if calls != 4 {
		t.Errorf("应完整跑满 4 次, got %d", calls)
	}
}

func TestRun_MultipleRounds(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 3, RoundIntervalSec: 0}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	var starts, dones []int
	for _, e := range events {
		if e.Kind == EventRoundStart {
			starts = append(starts, e.Round)
		}
		if e.Kind == EventRoundDone {
			dones = append(dones, e.Round)
		}
	}
	if len(starts) != 3 || len(dones) != 3 {
		t.Fatalf("应有 3 轮 start/done, got starts=%v dones=%v", starts, dones)
	}
	for i, want := range []int{1, 2, 3} {
		if starts[i] != want || dones[i] != want {
			t.Errorf("第 %d 轮编号不对: starts=%v dones=%v", i, starts, dones)
		}
	}
}

func TestRun_ConcurrencyProcessesAllAccounts(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1", "ck2", "ck3", "ck4", "ck5", "ck6")
	cfg := RunConfig{Threads: 4, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	events := collectEvents(t, New[testItem](&fakeRequester{}, nil), cfg, nil, targets, items)

	seen := map[int]bool{}
	for _, e := range events {
		if e.Kind == EventAttemptSuccess {
			seen[e.AccountIndex] = true
		}
	}
	if len(seen) != 6 {
		t.Errorf("应处理全部 6 个账号, got %d: %v", len(seen), seen)
	}
}

func TestRun_CreateRepoFailureSwitchesWithoutAttempt(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 5, FailSwitchCount: 5, CycleRounds: 1, CreateRepo: true}
	fr := &fakeRequester{}
	events := collectEvents(t, New[testItem](fr, repoCreatorFunc(func(ctx context.Context, acc account.Account) error {
		return errors.New("no space")
	})), cfg, nil, targets, items)

	if fr.callCount("ck1") != 0 {
		t.Errorf("建仓库失败不应尝试处理, 实际调用 %d 次", fr.callCount("ck1"))
	}
	var switched, attempts int
	for _, e := range events {
		if e.Kind == EventAccountSwitch {
			switched++
		}
		if e.Kind == EventAttemptStart || e.Kind == EventAttemptSuccess || e.Kind == EventAttemptFailure {
			attempts++
		}
	}
	if switched != 1 {
		t.Errorf("应只有一次换号事件, got %d: %+v", switched, events)
	}
	if attempts != 0 {
		t.Errorf("建仓库失败不应产生任何处理尝试事件, got %d", attempts)
	}
}

func TestRun_PauseBlocksUntilResumed(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	gate := NewPauseGate()
	gate.Pause()

	events := make(chan Event, 8)
	done := make(chan error, 1)
	go func() {
		done <- New[testItem](&fakeRequester{}, nil).Run(context.Background(), cfg, gate, targets, items, itemTitle, func(e Event) {
			events <- e
		})
	}()

	// 轮次开始事件不受暂停影响（在派发 worker 之前就会发出），但暂停期间不应有任何处理尝试事件。
	select {
	case e := <-events:
		if e.Kind != EventRoundStart {
			t.Fatalf("暂停期间只应看到 EventRoundStart, got %+v", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("迟迟没有收到 EventRoundStart")
	}
	select {
	case e := <-events:
		t.Fatalf("暂停期间不应产生处理尝试事件, got %+v", e)
	case <-time.After(150 * time.Millisecond):
	}

	gate.Resume()

	select {
	case e := <-events:
		if e.Kind != EventAttemptStart {
			t.Errorf("恢复后第一个事件应为 EventAttemptStart, got %v", e.Kind)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("恢复后长时间没有事件")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run err: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run 迟迟未结束")
	}
}

func TestRun_CancelDuringWorkStopsProcessing(t *testing.T) {
	items := []testItem{{Title: "a"}}
	targets := testTargets("ck1", "ck2", "ck3")
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	ctx, cancel := context.WithCancel(context.Background())

	var mu sync.Mutex
	var started int
	err := New[testItem](&fakeRequester{}, nil).Run(ctx, cfg, nil, targets, items, itemTitle, func(e Event) {
		if e.Kind == EventAttemptStart {
			mu.Lock()
			started++
			mu.Unlock()
			cancel()
		}
	})
	if err == nil {
		t.Errorf("取消后 Run 应返回 context 错误")
	}
	if started != 1 {
		t.Errorf("取消后不应再开始新账号, started=%d", started)
	}
}

func TestRun_NoContentOrAccountsIsNoop(t *testing.T) {
	cfg := RunConfig{Threads: 1, PerAccountCount: 1, FailSwitchCount: 1, CycleRounds: 1}
	r := New[testItem](&fakeRequester{}, nil)

	if err := r.Run(context.Background(), cfg, nil, testTargets("ck1"), nil, itemTitle, func(Event) {
		t.Fatal("没有内容时不应产生事件")
	}); err != nil {
		t.Fatalf("Run err: %v", err)
	}
	if err := r.Run(context.Background(), cfg, nil, nil, []testItem{{Title: "a"}}, itemTitle, func(Event) {
		t.Fatal("没有账号时不应产生事件")
	}); err != nil {
		t.Fatalf("Run err: %v", err)
	}
}

type requesterFunc func(ctx context.Context, acc account.Account, item testItem) (string, error)

func (f requesterFunc) Publish(ctx context.Context, acc account.Account, item testItem) (string, error) {
	return f(ctx, acc, item)
}

type repoCreatorFunc func(ctx context.Context, acc account.Account) error

func (f repoCreatorFunc) CreateSpace(ctx context.Context, acc account.Account) error {
	return f(ctx, acc)
}
