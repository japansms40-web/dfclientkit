// Package runlog 把 taskrunner 的处理事件格式化为运行日志文案，供前端渲染着色。
package runlog

import (
	"fmt"

	"dfclientkit/taskrunner"
)

// Kind 标识一条日志在前端应使用的着色分类，前端据此映射到主题色。
type Kind string

const (
	KindStart   Kind = "start"
	KindSuccess Kind = "success"
	KindFailure Kind = "failure"
	KindSwitch  Kind = "switch"
	KindInfo    Kind = "info"
)

// TagFor 返回某类事件的日志标签、着色分类，以及正文是否也要跟着上色
// （失败/换号的正文本身也标红/标黄，其余类型正文用默认前景色）。
// EventRoundProgress 不产生日志行，调用方应在此之前过滤掉。
func TagFor(k taskrunner.EventKind) (tag string, kind Kind, highlightMessage bool) {
	switch k {
	case taskrunner.EventAttemptStart:
		return "[开始]", KindStart, false
	case taskrunner.EventAttemptSuccess:
		return "[成功]", KindSuccess, false
	case taskrunner.EventAttemptFailure:
		return "[失败]", KindFailure, true
	case taskrunner.EventAccountSwitch:
		return "[换号]", KindSwitch, true
	case taskrunner.EventRoundStart, taskrunner.EventRoundDone:
		return "[轮次]", KindInfo, false
	default:
		return "[信息]", KindInfo, false
	}
}

// LineFor 把一条事件格式化为日志正文（不含时间戳/标签）。verb 是这次任务的业务
// 动作动词（比如"发布"/"注册"/"采集"），传空字符串时用缺省的"处理"。
func LineFor(e taskrunner.Event, verb string) string {
	if verb == "" {
		verb = "处理"
	}
	switch e.Kind {
	case taskrunner.EventAttemptStart:
		return fmt.Sprintf("账号 %s 开始%s《%s》", e.CK, verb, e.ItemLabel)
	case taskrunner.EventAttemptSuccess:
		return fmt.Sprintf("账号 %s %s《%s》成功: %s", e.CK, verb, e.ItemLabel, e.Result)
	case taskrunner.EventAttemptFailure:
		return fmt.Sprintf("账号 %s %s《%s》失败: %v", e.CK, verb, e.ItemLabel, e.Err)
	case taskrunner.EventAccountSwitch:
		return fmt.Sprintf("账号 %s 换号: %v", e.CK, e.Err)
	case taskrunner.EventRoundStart:
		return fmt.Sprintf("第 %d 轮开始，共 %d 个账号", e.Round, e.RoundTotal)
	case taskrunner.EventRoundDone:
		return fmt.Sprintf("第 %d 轮结束", e.Round)
	default:
		return e.CK
	}
}
