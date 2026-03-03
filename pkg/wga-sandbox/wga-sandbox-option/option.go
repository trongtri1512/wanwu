// Package wga_sandbox_option 提供 wga_sandbox 的选项配置。
package wga_sandbox_option

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/getkin/kin-openapi/openapi3"
)

// ============================================================================
// 常量
// ============================================================================

const (
	SandboxTypeReuse   SandboxType = "reuse"   // 复用容器模式
	SandboxTypeOneshot SandboxType = "oneshot" // 一次性容器模式
)

const (
	RunnerTypeOpencode RunnerType = "opencode" // opencode 智能体（默认）
)

const (
	sandboxAPIPort  = 8080 // sandbox API 端口
	opencodeAPIPort = 4096 // opencode API 端口
)

// ============================================================================
// 类型 - 配置
// ============================================================================

// ModelConfig 模型配置。
type ModelConfig struct {
	Provider     string // 提供商标识
	ProviderName string // 提供商显示名称
	BaseURL      string // API 基础地址
	APIKey       string // API 密钥
	Model        string // 模型标识
	ModelName    string // 模型显示名称
}

// Message 消息结构。
type Message struct {
	Role    string // 角色：user, assistant, system
	Content string // 消息内容
}

// Tool 工具配置。
type Tool struct {
	OpenAPI3Schema *openapi3.T // OpenAPI 3.0 schema 文档
	OperationIDs   []string    // 允许的 operations，为空则全部允许
	APIAuth        *openapi3_util.Auth
	Name           string // 工具名称，从 schema 的 info.title 自动读取
}

// Skill 技能配置。
type Skill struct {
	Dir string // skill 目录路径
}

// RunSession 执行会话标识。
type RunSession struct {
	ThreadID string // 对话会话 ID
	RunID    string // 执行请求 ID
}

// SandboxType 沙箱类型。
type SandboxType string

// RunnerType 运行器类型。
type RunnerType string

// ============================================================================
// 类型 - SandboxConfig
// ============================================================================

// SandboxConfig 沙箱配置。
type SandboxConfig struct {
	sandboxType SandboxType
	host        string // localhost 或容器名
	imageName   string // oneshot 模式用
}

func (c SandboxConfig) Type() SandboxType {
	return c.sandboxType
}

func (c SandboxConfig) Host() string {
	return c.host
}

func (c SandboxConfig) ImageName() string {
	return c.imageName
}

func (c SandboxConfig) APIEndpoint() string {
	return fmt.Sprintf("http://%s:%d", c.host, sandboxAPIPort)
}

func (c SandboxConfig) OpencodeEndpoint() string {
	return fmt.Sprintf("http://%s:%d", c.host, opencodeAPIPort)
}

// ============================================================================
// 类型 - Option/RunOption
// ============================================================================

// Option 选项接口。
type Option interface {
	apply(*RunOption) error
}

// OptionFunc 选项函数。
type OptionFunc func(*RunOption) error

func (f OptionFunc) apply(opts *RunOption) error {
	return f(opts)
}

// RunOption 运行选项。
type RunOption struct {
	RunSession     RunSession
	ModelConfig    ModelConfig
	Sandbox        SandboxConfig
	RunnerType     RunnerType
	Instruction    string
	OverallTask    string
	CurrentTask    string
	InputDir       string
	OutputDir      string
	Messages       []Message
	Skills         []Skill
	Tools          []Tool
	EnableThinking bool
	SkipCleanup    bool
	AgentName      string
}

func (o *RunOption) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt.apply(o); err != nil {
			return err
		}
	}
	if o.RunSession.ThreadID == "" {
		o.RunSession.ThreadID = uuid.New().String()
	}
	if o.RunSession.RunID == "" {
		o.RunSession.RunID = uuid.New().String()
	}
	if o.Sandbox.Type() == "" {
		o.Sandbox.sandboxType = SandboxTypeReuse
	}
	if o.Sandbox.Host() == "" {
		return fmt.Errorf("sandbox requires host")
	}
	if o.Sandbox.Type() == SandboxTypeOneshot && o.Sandbox.ImageName() == "" {
		return fmt.Errorf("oneshot sandbox requires image name")
	}
	return nil
}

// ============================================================================
// 构造函数
// ============================================================================

// SandboxReuse 创建复用容器模式的沙箱配置。
func SandboxReuse(host string) SandboxConfig {
	return SandboxConfig{
		sandboxType: SandboxTypeReuse,
		host:        host,
	}
}

// SandboxOneshot 创建一次性容器模式的沙箱配置。
func SandboxOneshot(imageName string) SandboxConfig {
	return SandboxConfig{
		sandboxType: SandboxTypeOneshot,
		imageName:   imageName,
	}
}

// ============================================================================
// 选项函数
// ============================================================================

func WithModelConfig(config ModelConfig) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.ModelConfig = config
		return nil
	})
}

func WithRunSession(session RunSession) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.RunSession = session
		return nil
	})
}

func WithSandbox(cfg SandboxConfig) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.Sandbox = cfg
		return nil
	})
}

func WithRunnerType(runnerType RunnerType) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.RunnerType = runnerType
		return nil
	})
}

func WithInstruction(instruction string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.Instruction = instruction
		return nil
	})
}

func WithOverallTask(overallTask string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.OverallTask = overallTask
		return nil
	})
}

func WithCurrentTask(currentTask string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.CurrentTask = currentTask
		return nil
	})
}

func WithInputDir(inputDir string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.InputDir = inputDir
		return nil
	})
}

func WithOutputDir(outputDir string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.OutputDir = outputDir
		return nil
	})
}

func WithMessages(messages []Message) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.Messages = messages
		return nil
	})
}

func WithSkills(skills []Skill) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.Skills = skills
		return nil
	})
}

func WithTools(tools []Tool) Option {
	return OptionFunc(func(opts *RunOption) error {
		ctx := context.Background()
		for i := range tools {
			if tools[i].OpenAPI3Schema == nil {
				return fmt.Errorf("tool schema is required")
			}
			if err := openapi3_util.ValidateDoc(ctx, tools[i].OpenAPI3Schema); err != nil {
				return fmt.Errorf("invalid tool schema: %w", err)
			}
			if tools[i].APIAuth != nil && tools[i].APIAuth.Type != "none" && tools[i].APIAuth.Value == "" {
				return fmt.Errorf("tool [%s] auth value is empty", tools[i].Name)
			}
			if tools[i].Name == "" {
				tools[i].Name = tools[i].OpenAPI3Schema.Info.Title
			}
			if len(tools[i].OperationIDs) > 0 {
				tools[i].OpenAPI3Schema = openapi3_util.FilterDocOperations(tools[i].OpenAPI3Schema, tools[i].OperationIDs)
			}
			opts.Tools = append(opts.Tools, tools[i])
		}
		return nil
	})
}

func WithEnableThinking(enable bool) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.EnableThinking = enable
		return nil
	})
}

func WithSkipCleanup(skip bool) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.SkipCleanup = skip
		return nil
	})
}

func WithAgentName(name string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.AgentName = name
		return nil
	})
}
