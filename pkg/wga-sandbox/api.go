// Package wga_sandbox 提供沙箱容器交互功能，支持在隔离环境中执行智能体任务。
package wga_sandbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/runner/opencode"
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/internal/sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
)

var manager = sandbox.NewManager()

type jsonErrorEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Run 在沙箱容器中执行智能体任务。
func Run(ctx context.Context, opts ...wga_sandbox_option.Option) (<-chan string, error) {
	var opt wga_sandbox_option.RunOption
	if err := opt.Apply(opts...); err != nil {
		return nil, fmt.Errorf("apply options failed: %w", err)
	}

	runID := opt.RunSession.RunID
	if err := manager.Create(ctx, runID, opt.Sandbox); err != nil {
		return nil, fmt.Errorf("create sandbox failed: %w", err)
	}

	logPrefix := fmt.Sprintf("[wga-sandbox][%s]", runID)
	if opt.AgentName != "" {
		logPrefix = fmt.Sprintf("[wga-sandbox][%s][%s]", runID, opt.AgentName)
	}

	log.Infof("%s %s", logPrefix, opt.CurrentTask)

	req := runner.RunRequest{
		RunSession:     opt.RunSession,
		Sandbox:        opt.Sandbox,
		Instruction:    opt.Instruction,
		OverallTask:    opt.OverallTask,
		CurrentTask:    opt.CurrentTask,
		Messages:       opt.Messages,
		InputDir:       opt.InputDir,
		OutputDir:      opt.OutputDir,
		Skills:         opt.Skills,
		Tools:          opt.Tools,
		ModelConfig:    opt.ModelConfig,
		EnableThinking: opt.EnableThinking,
	}

	sb, err := manager.Get(runID)
	if err != nil {
		return nil, fmt.Errorf("get sandbox failed: %w", err)
	}
	r := createRunner(opt.RunnerType, sb, req)

	outputCh := make(chan string, 1024)

	go func() {
		defer util.PrintPanicStack()
		defer close(outputCh)
		if !opt.SkipCleanup {
			defer func() { _ = manager.Cleanup(ctx, runID) }()
		}

		log.Infof("%s preparing...", logPrefix)
		if err := r.BeforeRun(ctx); err != nil {
			log.Errorf("%s prepare failed: %v", logPrefix, err)
			sendErrorEvent(outputCh, fmt.Sprintf("prepare failed: %v", err))
			return
		}

		log.Infof("%s running...", logPrefix)
		runnerOutputCh, err := r.Run(ctx)
		if err != nil {
			log.Errorf("%s run failed: %v", logPrefix, err)
			sendErrorEvent(outputCh, fmt.Sprintf("run failed: %v", err))
			return
		}

		for line := range runnerOutputCh {
			select {
			case outputCh <- line:
			case <-ctx.Done():
				return
			}
		}

		log.Infof("%s finishing...", logPrefix)
		if err := r.AfterRun(ctx); err != nil {
			log.Errorf("%s finish failed: %v", logPrefix, err)
			sendErrorEvent(outputCh, fmt.Sprintf("finish failed: %v", err))
			return
		}
		if opt.OutputDir != "" {
			log.Infof("%s output saved to: %s", logPrefix, opt.OutputDir)
		}
	}()

	return outputCh, nil
}

// Cleanup 清理指定 runID 的沙箱环境（用于 SkipCleanup=true 场景）。
func Cleanup(ctx context.Context, runID string) error {
	return manager.Cleanup(ctx, runID)
}

func sendErrorEvent(ch chan<- string, message string) {
	evt := jsonErrorEvent{Type: "error", Message: message}
	data, err := json.Marshal(evt)
	if err != nil {
		data = []byte(fmt.Sprintf(`{"type":"error","message":"%s"}`, message))
	}
	select {
	case ch <- string(data):
	default:
	}
}

func createRunner(t wga_sandbox_option.RunnerType, sb sandbox.Sandbox, req runner.RunRequest) runner.Runner {
	switch t {
	default:
		return opencode.NewRunner(sb, req)
	}
}
