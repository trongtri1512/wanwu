package ag_ui_util

import (
	"context"
	"encoding/json"

	"github.com/UnicomAI/wanwu/pkg/log"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/google/uuid"
	"github.com/sst/opencode-sdk-go"
)

// opencode 事件类型（内部使用）。
type opencodeEventType string

const (
	opencodeEventTypeText      opencodeEventType = "text"
	opencodeEventTypeToolUse   opencodeEventType = "tool_use"
	opencodeEventTypeReasoning opencodeEventType = "reasoning"
)

// opencodeEvent opencode 事件结构（内部使用）。
type opencodeEvent struct {
	Type opencodeEventType `json:"type"`
	Part json.RawMessage   `json:"part"`
}

// opencodeErrorPart opencode 错误事件内容（内部使用）。
type opencodeErrorPart struct {
	Error struct {
		Name string `json:"name"`
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	} `json:"error"`
}

// OpencodeTranslator opencode 事件转换器。
type OpencodeTranslator struct {
	BaseState
	activeToolCalls map[string]bool
}

// NewOpencodeTranslator 创建 opencode 转换器。
func NewOpencodeTranslator(runID, threadID string) *OpencodeTranslator {
	return &OpencodeTranslator{
		BaseState:       NewBaseState(runID, threadID),
		activeToolCalls: make(map[string]bool),
	}
}

// Translate 转换单个 opencode 事件。
func (t *OpencodeTranslator) Translate(_ context.Context, line string) []aguievents.Event {
	var evt opencodeEvent
	if err := json.Unmarshal([]byte(line), &evt); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse opencode event: %v", t.runID, err)
		return nil
	}

	if t.runFinished {
		log.Warnf("[ag-ui-util][%s] stream already finished, ignoring event: %s", t.runID, evt.Type)
		return nil
	}

	if t.messageID == "" {
		t.messageID = uuid.NewString()
	}

	var events []aguievents.Event

	switch evt.Type {
	case opencodeEventTypeText:
		events = t.translateText(evt.Part)
	case opencodeEventTypeToolUse:
		events = t.translateToolUse(evt.Part)
	case opencodeEventTypeReasoning:
		events = t.translateReasoning(evt.Part)
	case "error":
		events = t.translateError(evt.Part)
	}

	return events
}

// Finish 生成结束事件。
func (t *OpencodeTranslator) Finish() []aguievents.Event {
	return t.FinishBase()
}

// TranslateStream 转换事件流。
func (t *OpencodeTranslator) TranslateStream(ctx context.Context, in <-chan string) <-chan aguievents.Event {
	return TranslateStream(ctx, t, in)
}

func (t *OpencodeTranslator) translateText(partData json.RawMessage) []aguievents.Event {
	var part opencode.TextPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse text part: %v", t.runID, err)
		return nil
	}

	if part.Text == "" {
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.StartTextMessage()...)
	events = append(events, aguievents.NewTextMessageContentEvent(t.messageID, part.Text))
	return events
}

func (t *OpencodeTranslator) translateReasoning(partData json.RawMessage) []aguievents.Event {
	var part opencode.ReasoningPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse reasoning part: %v", t.runID, err)
		return nil
	}

	if part.Text == "" {
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.StartTextMessage()...)

	reasoningContent := formatReasoningContent(part.Text)
	events = append(events, aguievents.NewTextMessageContentEvent(t.messageID, reasoningContent))
	return events
}

func (t *OpencodeTranslator) translateError(partData json.RawMessage) []aguievents.Event {
	var part opencodeErrorPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse error part: %v", t.runID, err)
		return nil
	}

	msg := "[error] " + part.Error.Name
	if part.Error.Data.Message != "" {
		msg += ": " + part.Error.Data.Message
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.StartTextMessage()...)
	events = append(events, aguievents.NewTextMessageContentEvent(t.messageID, msg))
	return events
}

func (t *OpencodeTranslator) translateToolUse(partData json.RawMessage) []aguievents.Event {
	var part opencode.ToolPart
	if err := json.Unmarshal(partData, &part); err != nil {
		log.Warnf("[ag-ui-util][%s] failed to parse tool_use part: %v", t.runID, err)
		return nil
	}

	var events []aguievents.Event
	events = append(events, t.EnsureRunStarted()...)
	events = append(events, t.StartTextMessage()...)

	toolCallID := part.CallID
	if toolCallID == "" {
		toolCallID = part.ID
	}

	switch part.State.Status {
	case opencode.ToolPartStateStatusPending, opencode.ToolPartStateStatusRunning:
		if !t.activeToolCalls[toolCallID] {
			events = append(events, aguievents.NewToolCallStartEvent(toolCallID, part.Tool, aguievents.WithParentMessageID(t.messageID)))
			t.activeToolCalls[toolCallID] = true
			if input := t.getToolInput(part.State); input != "" {
				events = append(events, aguievents.NewToolCallArgsEvent(toolCallID, input))
			}
		}

	case opencode.ToolPartStateStatusCompleted:
		if !t.activeToolCalls[toolCallID] {
			events = append(events, aguievents.NewToolCallStartEvent(toolCallID, part.Tool, aguievents.WithParentMessageID(t.messageID)))
			if input := t.getToolInput(part.State); input != "" {
				events = append(events, aguievents.NewToolCallArgsEvent(toolCallID, input))
			}
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
		} else {
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
			delete(t.activeToolCalls, toolCallID)
		}
		resultMessageID := uuid.NewString()
		events = append(events, aguievents.NewToolCallResultEvent(resultMessageID, toolCallID, part.State.Output))

	case opencode.ToolPartStateStatusError:
		if !t.activeToolCalls[toolCallID] {
			events = append(events, aguievents.NewToolCallStartEvent(toolCallID, part.Tool, aguievents.WithParentMessageID(t.messageID)))
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
		} else {
			events = append(events, aguievents.NewToolCallEndEvent(toolCallID))
			delete(t.activeToolCalls, toolCallID)
		}
		resultMessageID := uuid.NewString()
		events = append(events, aguievents.NewToolCallResultEvent(resultMessageID, toolCallID, part.State.Error))
	}

	return events
}

func (t *OpencodeTranslator) getToolInput(state opencode.ToolPartState) string {
	if state.Input == nil {
		return ""
	}
	switch v := state.Input.(type) {
	case string:
		return v
	case map[string]interface{}:
		data, _ := json.Marshal(v)
		return string(data)
	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}
