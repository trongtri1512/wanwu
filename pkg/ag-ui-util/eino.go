package ag_ui_util

import (
	"context"

	"github.com/UnicomAI/wanwu/pkg/util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino/adk"
	"github.com/google/uuid"
)

// EinoTranslator eino AgentEvent 转换器。
type EinoTranslator struct {
	BaseState
	toolCallIDs map[string]bool
}

// NewEinoTranslator 创建 eino 转换器。
func NewEinoTranslator(runID, threadID string) *EinoTranslator {
	return &EinoTranslator{
		BaseState:   NewBaseState(runID, threadID),
		toolCallIDs: make(map[string]bool),
	}
}

// Translate 转换单个 eino AgentEvent。
func (t *EinoTranslator) Translate(_ context.Context, event *adk.AgentEvent) []aguievents.Event {
	if t.runFinished {
		return nil
	}

	if t.messageID == "" {
		t.messageID = uuid.NewString()
	}

	if event.Err != nil {
		return nil
	}

	if event.Action != nil && event.Action.Exit {
		return t.Finish()
	}

	if event.Output == nil || event.Output.MessageOutput == nil {
		return nil
	}

	msg := event.Output.MessageOutput.Message
	if msg == nil {
		return nil
	}

	hasContent := msg.Content != "" || msg.ReasoningContent != "" || len(msg.ToolCalls) > 0
	if !hasContent {
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.StartTextMessage()...)

	if msg.ReasoningContent != "" {
		content := formatReasoningContent(msg.ReasoningContent)
		events = append(events, aguievents.NewTextMessageContentEvent(t.messageID, content))
	}

	if msg.Content != "" {
		events = append(events, aguievents.NewTextMessageContentEvent(t.messageID, msg.Content))
	}

	for _, tc := range msg.ToolCalls {
		if !t.toolCallIDs[tc.ID] {
			events = append(events, aguievents.NewToolCallStartEvent(tc.ID, tc.Function.Name, aguievents.WithParentMessageID(t.messageID)))
			events = append(events, aguievents.NewToolCallArgsEvent(tc.ID, tc.Function.Arguments))
			events = append(events, aguievents.NewToolCallEndEvent(tc.ID))
			t.toolCallIDs[tc.ID] = true
		}
	}

	return events
}

// Finish 生成结束事件。
func (t *EinoTranslator) Finish() []aguievents.Event {
	return t.FinishBase()
}

// TranslateStream 转换事件流。
func (t *EinoTranslator) TranslateStream(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent]) <-chan aguievents.Event {
	out := make(chan aguievents.Event, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(out)
		for {
			event, ok := iter.Next()
			if !ok {
				for _, evt := range t.Finish() {
					out <- evt
				}
				return
			}
			for _, evt := range t.Translate(ctx, event) {
				select {
				case out <- evt:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
