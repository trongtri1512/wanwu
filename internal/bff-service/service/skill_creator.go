package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/gin-gonic/gin"
)

func RunSkillCreator(ctx *gin.Context, modelConfig wga_sandbox_option.ModelConfig, runId, inputDir, outputDir string, messages []adk.Message) (<-chan string, error) {
	skillCreatorCfg := config.Cfg().SkillCreator

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory failed: %w", err)
	}

	opts := buildSkillCreatorOptions(modelConfig, runId, inputDir, outputDir, messages, skillCreatorCfg)

	_, jsonCh, err := wga_sandbox.Run(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("run sandbox failed: %w", err)
	}

	filteredCh := filterOpencodeEvents(jsonCh)
	return filteredCh, nil
}

func buildSkillCreatorOptions(modelConfig wga_sandbox_option.ModelConfig, runId, inputDir, outputDir string, messages []adk.Message, skillCreatorCfg config.SkillCreatorConfig) []wga_sandbox_option.Option {

	opts := []wga_sandbox_option.Option{
		wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{RunID: runId}),
		wga_sandbox_option.WithModelConfig(modelConfig),
		wga_sandbox_option.WithOutputDir(outputDir),
		wga_sandbox_option.WithMessages(messages),
		wga_sandbox_option.WithEnableThinking(skillCreatorCfg.EnableThinking),
	}

	if inputDir != "" {
		opts = append(opts, wga_sandbox_option.WithInputDir(filepath.Clean(inputDir)+"/."))
	}

	if skillCreatorCfg.Instruction != "" {
		opts = append(opts, wga_sandbox_option.WithInstruction(skillCreatorCfg.Instruction))
	}

	sandboxCfg := config.Cfg().WgaSandbox.Sandbox
	switch sandboxCfg.Type {
	case string(wga_sandbox_option.SandboxTypeOneshot):
		opts = append(opts, wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxOneshot(sandboxCfg.ImageName)))
	default:
		opts = append(opts, wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxReuse(sandboxCfg.Host)))
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
	resultCh := make(chan string, 1024)

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
				resultCh <- opencodeTextPart.Text
			case wga_sandbox.OpencodeEventTypeToolUse:
				toolPart, err := wga_sandbox.ParseOpencodeToolPart(event.Part)
				if err != nil {
					continue
				}
				input, _ := json.Marshal(toolPart.State.Input)
				resultCh <- fmt.Sprintf("工具名称: %s", toolPart.Tool)
				resultCh <- fmt.Sprintf("<tool>\n\n%s工具参数: \n%s\n%s\n\n\\", "```", string(input), "```")
				// 工具调用结果输出可能较长，暂不输出
				// if toolPart.State.Output != "" {
				// 	resultCh <- fmt.Sprintf("%s%s 调用结果：\n %s %s", "```", toolPart.Tool, toolPart.State.Output, "```")
				// }
				resultCh <- "</tool>"
			case wga_sandbox.OpencodeEventTypeError:
				opencodeErrorPart, err := wga_sandbox.ParseOpencodeErrorPart(event.Part)
				if err != nil {
					continue
				}
				resultCh <- fmt.Sprintf("%v: %v", opencodeErrorPart.Error.Name, opencodeErrorPart.Error.Data.Message)
			}
		}
	}()

	return resultCh
}
