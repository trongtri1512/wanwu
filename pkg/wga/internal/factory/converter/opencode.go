package converter

import (
	"encoding/json"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	"github.com/cloudwego/eino/schema"
)

// opencodeConverter 实现 opencode runner 的事件转换。
type opencodeConverter struct {
	handlers map[wga_sandbox.OpencodeEventType]partParser
}

func newOpencodeConverter() *opencodeConverter {
	return &opencodeConverter{
		handlers: map[wga_sandbox.OpencodeEventType]partParser{
			wga_sandbox.OpencodeEventTypeStepStart:  parseSkipPart,
			wga_sandbox.OpencodeEventTypeStepFinish: parseSkipPart,
			wga_sandbox.OpencodeEventTypeText:       parseTextPart,
			wga_sandbox.OpencodeEventTypeReasoning:  parseReasoningPart,
			wga_sandbox.OpencodeEventTypeToolUse:    parseToolUsePart,
			wga_sandbox.OpencodeEventTypeFile:       parseFilePart,
			wga_sandbox.OpencodeEventTypeSnapshot:   parseSnapshotPart,
			wga_sandbox.OpencodeEventTypeAgent:      parseAgentPart,
			wga_sandbox.OpencodeEventTypePatch:      parsePatchPart,
			wga_sandbox.OpencodeEventTypeRetry:      parseRetryPart,
			"error":                                 parseErrorPart,
		},
	}
}

func (c *opencodeConverter) Convert(line string) (*schema.Message, error) {
	event, err := wga_sandbox.ParseOpencodeEvent([]byte(line))
	if err != nil {
		return nil, err
	}

	handler := c.handlers[event.Type]
	if handler == nil {
		log.Warnf("[converter][opencode] unknown event type: %s", event.Type)
		return nil, nil
	}

	content, err := handler(event.Part)
	if err != nil {
		return nil, err
	}
	if content.skip {
		return nil, nil
	}

	return &schema.Message{
		Role:             schema.Assistant,
		Content:          content.content,
		ReasoningContent: content.reasoningContent,
		ToolCalls:        content.toolCalls,
	}, nil
}

type messageContent struct {
	content          string
	reasoningContent string
	toolCalls        []schema.ToolCall
	skip             bool
}

type partParser func(json.RawMessage) (messageContent, error)

func parseSkipPart(_ json.RawMessage) (messageContent, error) {
	return messageContent{skip: true}, nil
}

func parseTextPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeTextPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: p.Text}, nil
}

func parseReasoningPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeReasoningPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{reasoningContent: p.Text}, nil
}

func parseToolUsePart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeToolPart(part)
	if err != nil {
		return messageContent{}, err
	}
	input := ""
	if p.State.Input != nil {
		data, _ := json.Marshal(p.State.Input)
		input = string(data)
	}
	return messageContent{
		toolCalls: []schema.ToolCall{
			{
				ID:   p.CallID,
				Type: "function",
				Function: schema.FunctionCall{
					Name:      p.Tool,
					Arguments: input,
				},
			},
		},
	}, nil
}

func parseFilePart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeFilePart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[file] " + p.Filename}, nil
}

func parseSnapshotPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeSnapshotPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[snapshot] " + p.ID}, nil
}

func parseAgentPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodeAgentPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[agent] " + p.Name}, nil
}

func parsePatchPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodePartPatchPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[patch] " + p.ID}, nil
}

func parseRetryPart(part json.RawMessage) (messageContent, error) {
	p, err := wga_sandbox.ParseOpencodePartRetryPart(part)
	if err != nil {
		return messageContent{}, err
	}
	return messageContent{content: "[retry] " + p.ID}, nil
}

type opencodeErrorPart struct {
	Error struct {
		Name string `json:"name"`
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	} `json:"error"`
}

func parseErrorPart(part json.RawMessage) (messageContent, error) {
	var p opencodeErrorPart
	if err := json.Unmarshal(part, &p); err != nil {
		return messageContent{}, err
	}
	msg := "[error] " + p.Error.Name
	if p.Error.Data.Message != "" {
		msg += ": " + p.Error.Data.Message
	}
	return messageContent{content: msg}, nil
}
