// Package factory 提供智能体实例创建功能。
package factory

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
)

// NewAgent 创建智能体实例。
func NewAgent(ctx context.Context, cfg *config.Agent, query string, options option.Options) (adk.Agent, error) {
	return newAgent(ctx, cfg, query, options)
}

func newAgent(ctx context.Context, cfg *config.Agent, query string, options option.Options) (adk.Agent, error) {
	switch cfg.Type {
	case config.AgentTypeReAct:
		return newReactAgent(ctx, cfg, query, options)
	case config.AgentTypeSandbox:
		return newSandboxAgentImpl(ctx, cfg, query, options)
	case config.AgentTypeSequential:
		return newSequentialAgent(ctx, cfg, query, options)
	case config.AgentTypeLoop:
		return newLoopAgent(ctx, cfg, query, options)
	case config.AgentTypeParallel:
		return newParallelAgent(ctx, cfg, query, options)
	case config.AgentTypeDeep:
		return newDeepAgent(ctx, cfg, query, options)
	case config.AgentTypeSupervisor:
		return newSupervisorAgent(ctx, cfg, query, options)
	default:
		return nil, fmt.Errorf("agent (%v) type (%v) unsupported", cfg.ID, cfg.Type)
	}
}

func newReactAgent(ctx context.Context, cfg *config.Agent, _ string, options option.Options) (adk.Agent, error) {
	instruction, err := options.FormatInstruction(ctx, cfg)
	if err != nil {
		return nil, err
	}
	model, err := options.ToChatModel(ctx)
	if err != nil {
		return nil, err
	}
	tools, err := options.ToToolsConfig(cfg.ToolCategories)
	if err != nil {
		return nil, err
	}
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		Instruction: instruction,
		Model:       model,
		ToolsConfig: tools,

		MaxIterations: cfg.Configure.MaxIterations,
	})
}

func newSequentialAgent(ctx context.Context, cfg *config.Agent, query string, options option.Options) (adk.ResumableAgent, error) {
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, query, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		SubAgents:   subAgents,
	})
}

func newLoopAgent(ctx context.Context, cfg *config.Agent, query string, options option.Options) (adk.ResumableAgent, error) {
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, query, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return adk.NewLoopAgent(ctx, &adk.LoopAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		SubAgents:   subAgents,

		MaxIterations: cfg.Configure.MaxIterations,
	})
}

func newParallelAgent(ctx context.Context, cfg *config.Agent, query string, options option.Options) (adk.ResumableAgent, error) {
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, query, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
		Name:        cfg.ID,
		Description: cfg.Description,
		SubAgents:   subAgents,
	})
}

func newDeepAgent(ctx context.Context, cfg *config.Agent, query string, options option.Options) (adk.ResumableAgent, error) {
	instruction, err := options.FormatInstruction(ctx, cfg)
	if err != nil {
		return nil, err
	}
	model, err := options.ToChatModel(ctx)
	if err != nil {
		return nil, err
	}
	tools, err := options.ToToolsConfig(cfg.ToolCategories)
	if err != nil {
		return nil, err
	}
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, query, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return deep.New(ctx, &deep.Config{
		Name:        cfg.ID,
		Description: cfg.Description,
		Instruction: instruction,
		ChatModel:   model,
		ToolsConfig: tools,
		SubAgents:   subAgents,

		MaxIteration:      cfg.Configure.MaxIterations,
		WithoutWriteTodos: true,
	})
}

func newSupervisorAgent(ctx context.Context, cfg *config.Agent, query string, options option.Options) (adk.Agent, error) {
	agent, err := newReactAgent(ctx, cfg, query, options)
	if err != nil {
		return nil, err
	}
	var subAgents []adk.Agent
	for _, subCfg := range cfg.SubAgents {
		subAgent, err := newAgent(ctx, subCfg, query, options)
		if err != nil {
			return nil, err
		}
		subAgents = append(subAgents, subAgent)
	}
	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: agent,
		SubAgents:  subAgents,
	})
}
