package ag_ui_util

import (
	"context"
	"io"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
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

// TranslateStream 转换事件流。
func (t *EinoTranslator) TranslateStream(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent]) <-chan aguievents.Event {
	out := make(chan aguievents.Event, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(out)
		defer func() {
			for _, evt := range t.FinishBase() {
				select {
				case out <- evt:
				case <-ctx.Done():
					return
				}
			}
		}()

		for {
			event, ok := iter.Next()
			if !ok {
				return
			}

			if event.Err != nil {
				// 发送错误信息作为文本消息
				errMsg := &schema.Message{
					Role:    schema.Assistant,
					Content: "[error] " + event.Err.Error(),
				}
				for _, evt := range t.translateMessage(errMsg) {
					select {
					case out <- evt:
					case <-ctx.Done():
						return
					}
				}
				return
			}

			if event.Action != nil && event.Action.Exit {
				return
			}

			if event.Output == nil || event.Output.MessageOutput == nil {
				continue
			}

			if t.messageID == "" {
				t.messageID = uuid.NewString()
			}

			msgOutput := event.Output.MessageOutput

			if msgOutput.IsStreaming {
				// 流式模式：边接收边发送
				t.translateStream(ctx, msgOutput, out)
			} else {
				// 非流式模式
				for _, evt := range t.translateMessage(msgOutput.Message) {
					select {
					case out <- evt:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}

// translateStream 处理流式消息，边接收边发送。
func (t *EinoTranslator) translateStream(ctx context.Context, msgOutput *adk.MessageVariant, out chan<- aguievents.Event) {
	if msgOutput.MessageStream == nil {
		return
	}
	defer msgOutput.MessageStream.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		frame, err := msgOutput.MessageStream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}

		for _, evt := range t.translateMessage(frame) {
			select {
			case out <- evt:
			case <-ctx.Done():
				return
			}
		}
	}
}

// translateMessage 转换消息（支持非流式和流式帧）。
func (t *EinoTranslator) translateMessage(msg *schema.Message) []aguievents.Event {
	if msg == nil {
		return nil
	}

	// 处理工具调用结果
	if msg.Role == schema.Tool && msg.ToolCallID != "" {
		var events []aguievents.Event
		events = append(events, t.EnsureRunStarted()...)
		events = append(events, aguievents.NewToolCallResultEvent(uuid.NewString(), msg.ToolCallID, msg.Content))
		return events
	}

	hasContent := msg.Content != "" || msg.ReasoningContent != "" || len(msg.ToolCalls) > 0
	if !hasContent {
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.StartTextMessage()...)

	// 处理 reasoning
	if msg.ReasoningContent != "" {
		events = append(events, t.StartReasoning()...)
		content := strings.ReplaceAll(msg.ReasoningContent, "\n", "\n> ")
		events = append(events, aguievents.NewTextMessageContentEvent(t.messageID, content))
	}

	// 处理普通内容，结束 reasoning 状态
	if msg.Content != "" {
		events = append(events, t.EndReasoning()...)
		events = append(events, aguievents.NewTextMessageContentEvent(t.messageID, msg.Content))
	}

	// 处理工具调用
	for _, tc := range msg.ToolCalls {
		// 跳过无效的 ToolCall
		if tc.ID == "" || tc.Function.Name == "" {
			continue
		}
		if !t.toolCallIDs[tc.ID] {
			events = append(events, aguievents.NewToolCallStartEvent(tc.ID, tc.Function.Name, aguievents.WithParentMessageID(t.messageID)))
			if tc.Function.Arguments != "" {
				events = append(events, aguievents.NewToolCallArgsEvent(tc.ID, tc.Function.Arguments))
			}
			events = append(events, aguievents.NewToolCallEndEvent(tc.ID))
			t.toolCallIDs[tc.ID] = true
		}
	}

	return events
}
