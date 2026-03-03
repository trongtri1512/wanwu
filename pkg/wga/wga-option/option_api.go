package wga_option

import (
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
)

type Option = option.Option
type ModelConfig = option.ModelConfig
type ToolConfig = option.ToolConfig
type WorkspaceConfig = option.WorkspaceConfig
type RunSession = option.RunSession
type Message = option.Message

type CheckResult = option.CheckResult

// WithModelConfig 设置模型配置。
func WithModelConfig(model ModelConfig) Option {
	return option.WithModelConfig(model)
}

// WithToolConfig 添加工具配置，工具标题不能重复。
func WithToolConfig(tool ToolConfig) Option {
	return option.WithToolConfig(tool)
}

// WithWorkspaceConfig 设置工作空间配置。
func WithWorkspaceConfig(workspace WorkspaceConfig) Option {
	return option.WithWorkspaceConfig(workspace)
}

// WithRunSession 设置执行会话标识（ThreadID 和 RunID）。
func WithRunSession(session RunSession) Option {
	return option.WithRunSession(session)
}

// WithMessages 设置历史消息。
func WithMessages(messages []Message) Option {
	return option.WithMessages(messages)
}
