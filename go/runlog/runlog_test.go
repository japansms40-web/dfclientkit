package runlog

import (
	"errors"
	"testing"

	"dfclientkit/taskrunner"
)

func TestTagForKindsAndHighlight(t *testing.T) {
	cases := []struct {
		kind     taskrunner.EventKind
		wantTag  string
		wantKind Kind
		wantHi   bool
	}{
		{taskrunner.EventAttemptStart, "[开始]", KindStart, false},
		{taskrunner.EventAttemptSuccess, "[成功]", KindSuccess, false},
		{taskrunner.EventAttemptFailure, "[失败]", KindFailure, true},
		{taskrunner.EventAccountSwitch, "[换号]", KindSwitch, true},
		{taskrunner.EventRoundStart, "[轮次]", KindInfo, false},
		{taskrunner.EventRoundDone, "[轮次]", KindInfo, false},
	}
	for _, c := range cases {
		tag, kind, hi := TagFor(c.kind)
		if tag != c.wantTag || kind != c.wantKind || hi != c.wantHi {
			t.Errorf("TagFor(%v) = %q/%v/%v, want %q/%v/%v", c.kind, tag, kind, hi, c.wantTag, c.wantKind, c.wantHi)
		}
	}
}

func TestLineForFormattingWithVerb(t *testing.T) {
	cases := []struct {
		event taskrunner.Event
		verb  string
		want  string
	}{
		{taskrunner.Event{Kind: taskrunner.EventAttemptStart, CK: "ck1", ItemLabel: "hello"}, "发布", "账号 ck1 开始发布《hello》"},
		{taskrunner.Event{Kind: taskrunner.EventAttemptSuccess, CK: "ck1", ItemLabel: "hello", Result: "ok"}, "发布", "账号 ck1 发布《hello》成功: ok"},
		{taskrunner.Event{Kind: taskrunner.EventAttemptFailure, CK: "ck1", ItemLabel: "hello", Err: errors.New("boom")}, "发布", "账号 ck1 发布《hello》失败: boom"},
		{taskrunner.Event{Kind: taskrunner.EventAccountSwitch, CK: "ck1", Err: errors.New("连续失败达到换号阈值")}, "发布", "账号 ck1 换号: 连续失败达到换号阈值"},
		{taskrunner.Event{Kind: taskrunner.EventRoundStart, Round: 2, RoundTotal: 5}, "发布", "第 2 轮开始，共 5 个账号"},
		{taskrunner.Event{Kind: taskrunner.EventRoundDone, Round: 2}, "发布", "第 2 轮结束"},
	}
	for _, c := range cases {
		if got := LineFor(c.event, c.verb); got != c.want {
			t.Errorf("LineFor(%+v, %q) = %q, want %q", c.event, c.verb, got, c.want)
		}
	}
}

func TestLineForDefaultsVerbToProcess(t *testing.T) {
	e := taskrunner.Event{Kind: taskrunner.EventAttemptStart, CK: "ck1", ItemLabel: "hello"}
	want := "账号 ck1 开始处理《hello》"
	if got := LineFor(e, ""); got != want {
		t.Errorf("LineFor(_, \"\") = %q, want %q", got, want)
	}
}
