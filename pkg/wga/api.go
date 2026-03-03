// Package wga 提供 AI 智能体的统一管理和执行接口。
//
// 支持多种智能体类型：react、sandbox、sequential、loop、parallel、deep、supervisor。
package wga

import (
	"context"
	"errors"
	"fmt"

	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/factory"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

var (
	ErrWgaNotInit     = errors.New("wga not init")
	ErrWgaAlreadyInit = errors.New("wga already init")
)

var _agents []*config.Agent

// Init 初始化智能体配置。
func Init(ctx context.Context, configPath string) error {
	if _agents != nil {
		return ErrWgaAlreadyInit
	}
	agents, err := config.LoadAgents(ctx, configPath)
	if err != nil {
		return err
	}
	_agents = agents
	return nil
}

// CheckOptions 检查智能体运行条件是否满足。
func CheckOptions(_ context.Context, id string, opts ...option.Option) (*wga_option.CheckResult, error) {
	agentCfg, err := getAgent(id)
	if err != nil {
		return nil, err
	}
	var options option.Options
	if err := options.Apply(opts...); err != nil {
		return nil, err
	}
	return options.CheckCondition(agentCfg)
}

// Run 执行智能体任务，返回事件迭代器。
func Run(ctx context.Context, id, query string, opts ...option.Option) (*adk.AsyncIterator[*adk.AgentEvent], error) {
	agentCfg, err := getAgent(id)
	if err != nil {
		return nil, err
	}
	var options option.Options
	if err := options.Apply(opts...); err != nil {
		return nil, err
	}
	agent, err := factory.NewAgent(ctx, agentCfg, query, options)
	if err != nil {
		return nil, err
	}
	input := &adk.AgentInput{
		Messages: []adk.Message{schema.UserMessage(query)},
	}
	return agent.Run(ctx, input), nil
}

func getAgent(id string) (*config.Agent, error) {
	for _, agent := range _agents {
		if agent.ID == id {
			return agent, nil
		}
	}
	return nil, fmt.Errorf("agent (%s) not found", id)
}

// Cleanup 清理指定 runID 的沙箱工作目录。
func Cleanup(ctx context.Context, runID string) error {
	return wga_sandbox.Cleanup(ctx, runID)
}
