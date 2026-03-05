package service

import (
	"fmt"
	"os"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/gin-gonic/gin"
)

// inputdir需要，threadID, runID不需要
func RunSkillCreator(ctx *gin.Context, modelConfig wga_sandbox_option.ModelConfig, inputDir, outputDir string, messages []wga_sandbox_option.Message) (<-chan string, error) {
	skillCreatorCfg := config.Cfg().SkillCreator

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory failed: %w", err)
	}

	opts := buildSkillCreatorOptions(modelConfig, inputDir, outputDir, messages, skillCreatorCfg)

	_, jsonCh, err := wga_sandbox.Run(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("run sandbox failed: %w", err)
	}

	filteredCh := filterOpencodeEvents(jsonCh)
	return filteredCh, nil
}

func buildSkillCreatorOptions(modelConfig wga_sandbox_option.ModelConfig, inputDir, outputDir string, messages []wga_sandbox_option.Message, skillCreatorCfg config.SkillCreatorConfig) []wga_sandbox_option.Option {

	opts := []wga_sandbox_option.Option{
		wga_sandbox_option.WithModelConfig(modelConfig),
		wga_sandbox_option.WithSkipCleanup(false),
		wga_sandbox_option.WithOutputDir(outputDir),
		wga_sandbox_option.WithEnableThinking(skillCreatorCfg.Agent.EnableThinking),
		wga_sandbox_option.WithInputDir(inputDir),
	}

	if skillCreatorCfg.Agent.Instruction != "" {
		opts = append(opts, wga_sandbox_option.WithInstruction(skillCreatorCfg.Agent.Instruction))
	}

	sandboxCfg := config.Cfg().WgaSandbox.Sandbox
	switch sandboxCfg.Type {
	case string(wga_sandbox_option.SandboxTypeOneshot):
		opts = append(opts, wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxOneshot(sandboxCfg.ImageName)))
	default:
		opts = append(opts, wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxReuse(sandboxCfg.Host)))
	}

	if len(messages) > 0 {
		optsMessages := make([]wga_sandbox_option.Message, len(messages))
		for i, msg := range messages {
			optsMessages[i] = wga_sandbox_option.Message{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		opts = append(opts, wga_sandbox_option.WithMessages(optsMessages))
	}

	if len(skillCreatorCfg.Skills) > 0 {
		skills := make([]wga_sandbox_option.Skill, len(skillCreatorCfg.Skills))
		for i, skill := range skillCreatorCfg.Skills {
			skills[i] = wga_sandbox_option.Skill{Dir: skill.Dir}
		}
		opts = append(opts, wga_sandbox_option.WithSkills(skills))
	}

	return opts
}

func filterOpencodeEvents(jsonCh <-chan string) <-chan string {
	resultCh := make(chan string, 10)

	go func() {
		defer util.PrintPanicStack()
		defer close(resultCh)
		for line := range jsonCh {
			event, err := wga_sandbox.ParseOpencodeEvent([]byte(line))
			if err != nil {
				continue
			}
			switch event.Type {
			case wga_sandbox.OpencodeEventTypeReasoning:
				continue
			case wga_sandbox.OpencodeEventTypeText:
				opencodeTextPart, err := wga_sandbox.ParseOpencodeTextPart(event.Part)
				if err != nil || opencodeTextPart.Text == "" {
					continue
				}
				msg := fmt.Sprintf(`{"response": "%s"}`, opencodeTextPart.Text)
				resultCh <- msg
			case wga_sandbox.OpencodeEventTypeToolUse:
				toolPart, err := wga_sandbox.ParseOpencodeToolPart(event.Part)
				if err != nil {
					continue
				}
				resultCh <- fmt.Sprintf(`{"response": "工具名称: %s"}`, toolPart.Tool)
				resultCh <- fmt.Sprintf(`{"response": "<tool>\n\n%s工具参数：\n%s\n%s\n\n\\"}`, "```", toolPart.State.Input, "```")
				if toolPart.State.Output != "" {
					resultCh <- fmt.Sprintf(`{"response": "%s%s 调用结果：\n %s %s"}`, "```", toolPart.Tool, toolPart.State.Output, "```")
				}
				resultCh <- `{"response": "</tool>"}`
			}
		}
	}()

	return resultCh
}
