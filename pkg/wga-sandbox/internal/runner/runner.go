// Package runner 提供智能体运行器接口。
package runner

import (
	"context"

	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
)

// Runner 智能体运行器接口。
type Runner interface {
	BeforeRun(ctx context.Context) error
	Run(ctx context.Context) (<-chan string, error)
	AfterRun(ctx context.Context) error
}

// RunRequest 运行请求参数。
type RunRequest struct {
	RunSession     wga_sandbox_option.RunSession
	Sandbox        wga_sandbox_option.SandboxConfig
	Instruction    string
	OverallTask    string
	CurrentTask    string
	InputDir       string
	OutputDir      string
	Messages       []wga_sandbox_option.Message
	Skills         []wga_sandbox_option.Skill
	Tools          []wga_sandbox_option.Tool
	ModelConfig    wga_sandbox_option.ModelConfig
	EnableThinking bool
}
