// Package ag_ui_util 提供 AG-UI 协议的事件转换功能。
package ag_ui_util

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
)

// BaseState 转换器基础状态。
type BaseState struct {
	runID       string
	threadID    string
	messageID   string
	runStarted  bool
	runFinished bool
	textStarted bool
	inReasoning bool
}

// NewBaseState 创建基础状态。
func NewBaseState(runID, threadID string) BaseState {
	return BaseState{
		runID:    runID,
		threadID: threadID,
	}
}

func (s *BaseState) RunID() string     { return s.runID }
func (s *BaseState) ThreadID() string  { return s.threadID }
func (s *BaseState) MessageID() string { return s.messageID }

// EnsureRunStarted 确保已发送 RunStarted 事件。
func (s *BaseState) EnsureRunStarted() []aguievents.Event {
	if s.runStarted {
		return nil
	}
	s.runStarted = true
	return []aguievents.Event{aguievents.NewRunStartedEvent(s.threadID, s.runID)}
}

// StartTextMessage 开始文本消息。
func (s *BaseState) StartTextMessage() []aguievents.Event {
	if s.textStarted {
		return nil
	}
	s.textStarted = true
	return []aguievents.Event{aguievents.NewTextMessageStartEvent(s.messageID, aguievents.WithRole("assistant"))}
}

// EndTextMessage 结束文本消息。
func (s *BaseState) EndTextMessage() []aguievents.Event {
	if !s.textStarted {
		return nil
	}
	s.textStarted = false
	return []aguievents.Event{aguievents.NewTextMessageEndEvent(s.messageID)}
}

// StartReasoning 开始 reasoning，返回 reasoningStart 事件（仅第一次）。
func (s *BaseState) StartReasoning() []aguievents.Event {
	if s.inReasoning {
		return nil
	}
	s.inReasoning = true
	return []aguievents.Event{aguievents.NewTextMessageContentEvent(s.messageID, reasoningStart)}
}

// EndReasoning 结束 reasoning，返回 reasoningEnd 事件。
func (s *BaseState) EndReasoning() []aguievents.Event {
	if !s.inReasoning {
		return nil
	}
	s.inReasoning = false
	return []aguievents.Event{aguievents.NewTextMessageContentEvent(s.messageID, reasoningEnd)}
}

// FinishBase 生成基础结束事件。
func (s *BaseState) FinishBase() []aguievents.Event {
	if s.runFinished {
		return nil
	}
	s.runFinished = true
	var events []aguievents.Event
	events = append(events, s.EnsureRunStarted()...)
	events = append(events, s.EndReasoning()...)
	events = append(events, s.EndTextMessage()...)
	events = append(events, aguievents.NewRunFinishedEvent(s.threadID, s.runID))
	return events
}

// EventsToJSONChannel 将事件流转换为 JSON 字符串流。
func EventsToJSONChannel(ctx context.Context, in <-chan aguievents.Event) <-chan string {
	out := make(chan string, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-in:
				if !ok {
					return
				}
				if data, err := json.Marshal(evt); err == nil {
					out <- string(data)
				}
			}
		}
	}()
	return out
}

const (
	reasoningStart = "\n> 💭 \n> "
	reasoningEnd   = "\n\n"
)

// RemoveReasoningContent 移除文本中的推理内容。
func RemoveReasoningContent(content string) string {
	for {
		idx := strings.Index(content, reasoningStart)
		if idx == -1 {
			return content
		}
		endIdx := strings.Index(content[idx:], reasoningEnd)
		if endIdx == -1 {
			return content[:idx]
		}
		content = content[:idx] + content[idx+endIdx+len(reasoningEnd):]
	}
}
