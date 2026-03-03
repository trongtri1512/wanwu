// Package option 提供智能体运行选项的内部实现。
package option

import (
	"fmt"

	"github.com/google/uuid"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
)

// ============================================================================
// 类型 - 配置
// ============================================================================

// ModelConfig 模型配置。
type ModelConfig struct {
	Model       string              // 模型标识
	ApiKey      string              // API 密钥
	EndpointUrl string              // API 端点地址
	Params      mp_common.LLMParams // 模型参数
}

// ToolConfig 工具配置。
type ToolConfig struct {
	Title   string                  // 工具标题（对应 OpenAPI schema 的 info.title）
	APIAuth *util.ApiAuthWebRequest // API 认证配置
}

// WorkspaceConfig 工作空间配置。
type WorkspaceConfig struct {
	InputDir  string // 输入目录路径
	OutputDir string // 输出目录路径
}

// RunSession 执行会话标识。
type RunSession struct {
	ThreadID string // 对话会话 ID
	RunID    string // 执行请求 ID
}

// Message 消息。
type Message struct {
	Role    string // 角色：user, assistant, system
	Content string // 消息内容
}

// ============================================================================
// Option/Options
// ============================================================================

// Option 选项接口。
type Option interface {
	apply(*Options) error
}

// optionFunc 选项函数。
type optionFunc func(*Options) error

func (f optionFunc) apply(opts *Options) error {
	return f(opts)
}

// Options 智能体运行选项。
type Options struct {
	Model      ModelConfig     // 模型配置
	Tools      []ToolConfig    // 工具配置列表
	Workspace  WorkspaceConfig // 工作空间配置
	RunSession RunSession      // 执行会话标识
	Messages   []Message       // 历史消息
}

// Apply 应用选项。
// 如果 ThreadID 或 RunID 为空，自动生成 UUID。
func (options *Options) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt.apply(options); err != nil {
			return err
		}
	}
	if options.RunSession.ThreadID == "" {
		options.RunSession.ThreadID = uuid.New().String()
	}
	if options.RunSession.RunID == "" {
		options.RunSession.RunID = uuid.New().String()
	}
	return nil
}

// ============================================================================
// Check
// ============================================================================

// CheckResult 条件检查结果。
type CheckResult struct {
	Model          CheckModel          // 模型检查结果
	ToolCategories []CheckToolCategory // 工具类别检查结果
}

// CheckModel 模型检查结果。
type CheckModel struct {
	Meet bool // 是否满足条件
}

// CheckToolCategory 工具类别检查结果。
type CheckToolCategory struct {
	Category  string      // 工具类别类型
	Condition string      // 工具类别条件
	Meet      bool        // 是否满足条件
	Tools     []CheckTool // 工具检查结果
}

// CheckTool 工具检查结果。
type CheckTool struct {
	Title string // 工具标题
	Meet  bool   // 是否满足条件
}

// CheckCondition 检查智能体运行条件是否满足。
func (options *Options) CheckCondition(cfg *config.Agent) (*CheckResult, error) {
	model := CheckModel{Meet: true}
	if err := options.checkModel(); err != nil {
		model.Meet = false
	}
	conditions, err := options.checkToolsCondition(cfg.ToolCategories)
	if err != nil {
		return nil, err
	}
	return &CheckResult{
		Model:          model,
		ToolCategories: conditions,
	}, nil
}

// ============================================================================
// 选项函数
// ============================================================================

// WithModelConfig 设置模型配置。
func WithModelConfig(model ModelConfig) Option {
	return optionFunc(func(opts *Options) error {
		opts.Model = model
		return nil
	})
}

// WithToolConfig 添加工具配置，工具标题不能重复。
func WithToolConfig(tool ToolConfig) Option {
	return optionFunc(func(opts *Options) error {
		if tool.APIAuth != nil {
			if err := tool.APIAuth.Check(); err != nil {
				return fmt.Errorf("tool (%v) check auth err: %v", tool.Title, err)
			}
		}
		for _, toolOpt := range opts.Tools {
			if toolOpt.Title == tool.Title {
				return fmt.Errorf("tool (%v) already exist", tool.Title)
			}
		}
		opts.Tools = append(opts.Tools, tool)
		return nil
	})
}

// WithWorkspaceConfig 设置工作空间配置。
func WithWorkspaceConfig(workspace WorkspaceConfig) Option {
	return optionFunc(func(opts *Options) error {
		opts.Workspace = workspace
		return nil
	})
}

// WithRunSession 设置执行会话标识（ThreadID 和 RunID）。
func WithRunSession(session RunSession) Option {
	return optionFunc(func(opts *Options) error {
		opts.RunSession = session
		return nil
	})
}

// WithMessages 设置历史消息。
func WithMessages(messages []Message) Option {
	return optionFunc(func(opts *Options) error {
		opts.Messages = messages
		return nil
	})
}
