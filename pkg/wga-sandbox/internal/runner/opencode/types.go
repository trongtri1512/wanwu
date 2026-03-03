package opencode

import "encoding/json"

// ============================================================================
// Opencode 输出类型（导出）
// ============================================================================

// OpencodeEventType 事件类型。
type OpencodeEventType string

// 事件类型常量。
const (
	OpencodeEventTypeStepStart  OpencodeEventType = "step_start"
	OpencodeEventTypeStepFinish OpencodeEventType = "step_finish"
	OpencodeEventTypeText       OpencodeEventType = "text"
	OpencodeEventTypeToolUse    OpencodeEventType = "tool_use"
	OpencodeEventTypeReasoning  OpencodeEventType = "reasoning"
	OpencodeEventTypeSnapshot   OpencodeEventType = "snapshot"
	OpencodeEventTypePatch      OpencodeEventType = "patch"
	OpencodeEventTypeFile       OpencodeEventType = "file"
	OpencodeEventTypeAgent      OpencodeEventType = "agent"
	OpencodeEventTypeRetry      OpencodeEventType = "retry"
	OpencodeEventTypeSubtask    OpencodeEventType = "subtask"
	OpencodeEventTypeCompaction OpencodeEventType = "compaction"
)

// OpencodeEvent 输出事件结构。
type OpencodeEvent struct {
	Type      OpencodeEventType `json:"type"`
	Timestamp int64             `json:"timestamp"`
	SessionID string            `json:"sessionID"`
	Part      json.RawMessage   `json:"part"`
}

// ============================================================================
// 事件部分类型
// ============================================================================

// textPart 文本部分。
type textPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// reasoningPart 推理部分。
type reasoningPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// toolPart 工具调用部分。
type toolPart struct {
	Type   string    `json:"type"`
	CallID string    `json:"callID,omitempty"`
	Tool   string    `json:"tool"`
	State  toolState `json:"state"`
}

// toolState 工具调用状态。
type toolState struct {
	Status string      `json:"status"`
	Input  interface{} `json:"input,omitempty"`
	Output string      `json:"output,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// ============================================================================
// SSE 事件类型（内部使用）
// ============================================================================

// sseEvent SSE 事件结构。
type sseEvent struct {
	Directory string          `json:"directory"`
	Payload   sseEventPayload `json:"payload"`
}

// sseEventPayload SSE 事件负载。
type sseEventPayload struct {
	Type       string        `json:"type"`
	Properties sseEventProps `json:"properties"`
}

// sseEventProps SSE 事件属性。
type sseEventProps struct {
	SessionID string         `json:"sessionID"`
	Part      sseEventPart   `json:"part"`
	Delta     string         `json:"delta"`
	Status    sseEventStatus `json:"status"`
}

// sseEventPart SSE 事件部分。
type sseEventPart struct {
	ID        string       `json:"id"`
	SessionID string       `json:"sessionID"`
	MessageID string       `json:"messageID"`
	Type      string       `json:"type"`
	Text      string       `json:"text"`
	CallID    string       `json:"callID"`
	Tool      string       `json:"tool"`
	State     sseToolState `json:"state"`
}

// sseToolState 工具调用状态。
type sseToolState struct {
	Status string      `json:"status"`
	Input  interface{} `json:"input,omitempty"`
	Output string      `json:"output,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// sseEventStatus SSE 事件状态。
type sseEventStatus struct {
	Type string `json:"type"`
}
