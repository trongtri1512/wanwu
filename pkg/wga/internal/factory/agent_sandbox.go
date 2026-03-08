package factory

import (
	"context"

	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_converter "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-converter"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
	"github.com/cloudwego/eino/adk"
)

// sandboxAgent 在沙箱容器中执行的智能体。
type sandboxAgent struct {
	cfg     *config.Agent
	query   string
	options option.Options
}

func newSandboxAgentImpl(_ context.Context, cfg *config.Agent, query string, options option.Options) (adk.Agent, error) {
	return &sandboxAgent{
		cfg:     cfg,
		query:   query,
		options: options,
	}, nil
}

func (a *sandboxAgent) Name(_ context.Context) string {
	return a.cfg.ID
}

func (a *sandboxAgent) Description(_ context.Context) string {
	return a.cfg.Description
}

func (a *sandboxAgent) Run(ctx context.Context, _ *adk.AgentInput, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	sandboxOpts := a.buildSandboxOpts(a.query)

	_, outputCh, err := wga_sandbox.Run(ctx, sandboxOpts...)
	if err != nil {
		return wga_sandbox_converter.ConvertToEinoIteratorWithError(ctx, wga_sandbox_option.RunnerTypeOpencode, err)
	}

	return wga_sandbox_converter.ConvertToEinoIterator(ctx, wga_sandbox_option.RunnerTypeOpencode, outputCh)
}

func (a *sandboxAgent) buildSandboxOpts(overallTask string) []wga_sandbox_option.Option {
	opts := []wga_sandbox_option.Option{
		wga_sandbox_option.WithModelConfig(wga_sandbox_option.ModelConfig{
			Provider:     a.options.Model.Provider,
			ProviderName: a.options.Model.ProviderName,
			BaseURL:      a.options.Model.BaseURL,
			APIKey:       a.options.Model.APIKey,
			Model:        a.options.Model.Model,
			ModelName:    a.options.Model.ModelName,
			Params:       a.options.Model.Params,
		}),
		wga_sandbox_option.WithInstruction(a.cfg.Prompt),
		wga_sandbox_option.WithEnableThinking(a.cfg.Configure.EnableThinking),
		wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
			ThreadID: a.options.RunSession.ThreadID,
			RunID:    a.options.RunSession.RunID,
		}),
		wga_sandbox_option.WithSkipCleanup(true),
		wga_sandbox_option.WithAgentName(a.cfg.ID),
	}

	if overallTask != "" {
		opts = append(opts, wga_sandbox_option.WithOverallTask(overallTask))
	}

	// 传递历史消息
	if len(a.options.Messages) > 0 {
		messages := make([]wga_sandbox_option.Message, len(a.options.Messages))
		for i, msg := range a.options.Messages {
			messages[i] = wga_sandbox_option.Message{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		opts = append(opts, wga_sandbox_option.WithMessages(messages))
	}

	// 传递技能
	if len(a.cfg.Skills) > 0 {
		skills := make([]wga_sandbox_option.Skill, len(a.cfg.Skills))
		for i, skill := range a.cfg.Skills {
			skills[i] = wga_sandbox_option.Skill{
				Dir: skill.Dir,
			}
		}
		opts = append(opts, wga_sandbox_option.WithSkills(skills))
	}

	if a.options.Workspace.InputDir != "" {
		opts = append(opts, wga_sandbox_option.WithInputDir(a.options.Workspace.InputDir))
	}
	if a.options.Workspace.OutputDir != "" {
		opts = append(opts, wga_sandbox_option.WithOutputDir(a.options.Workspace.OutputDir))
	}

	if a.cfg.Configure.Sandbox != nil {
		cfg := a.cfg.Configure.Sandbox
		switch cfg.Type {
		case "oneshot":
			opts = append(opts, wga_sandbox_option.WithSandbox(
				wga_sandbox_option.SandboxOneshot(cfg.ImageName),
			))
		default:
			opts = append(opts, wga_sandbox_option.WithSandbox(
				wga_sandbox_option.SandboxReuse(cfg.Host),
			))
		}
	}

	var tools []wga_sandbox_option.Tool
	for _, tc := range a.cfg.ToolCategories {
		for _, toolCfg := range tc.Tools {
			var auth *openapi3_util.Auth
			for _, toolOpt := range a.options.Tools {
				if toolOpt.Title == toolCfg.Doc.Info.Title {
					if converted, err := toolOpt.APIAuth.ToOpenapiAuth(); err == nil {
						auth = converted
					}
					break
				}
			}
			var operationIDs []string
			for _, op := range toolCfg.Operations {
				operationIDs = append(operationIDs, op.OperationID)
			}
			tools = append(tools, wga_sandbox_option.Tool{
				OpenAPI3Schema: toolCfg.Doc,
				OperationIDs:   operationIDs,
				APIAuth:        auth,
			})
		}
	}
	if len(tools) > 0 {
		opts = append(opts, wga_sandbox_option.WithTools(tools))
	}

	return opts
}
