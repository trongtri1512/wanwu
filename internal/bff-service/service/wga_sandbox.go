package service

import (
	"fmt"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/gin-gonic/gin"
)

func WgaSandboxRun(ctx *gin.Context, req *request.WgaSandboxRunReq) error {
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: req.Model.ModelId})
	if err != nil {
		return err
	}
	if !modelInfo.IsActive {
		return grpc_util.ErrorStatus(err_code.Code_BFFModelStatus, modelInfo.ModelId)
	}

	endpoint := mp.ToModelEndpoint(modelInfo.ModelId, modelInfo.Model)
	modelURL, _ := endpoint["model_url"].(string)
	modelConfig := wga_sandbox_option.ModelConfig{
		Provider:     modelInfo.Provider,
		ProviderName: modelInfo.Provider,
		BaseURL:      modelURL,
		APIKey:       "",
		Model:        modelInfo.Model,
		ModelName:    modelInfo.DisplayName,
	}

	sandboxCfg := config.Cfg().WgaSandbox.Sandbox
	var sandbox wga_sandbox_option.SandboxConfig
	if sandboxCfg.Type == "oneshot" {
		sandbox = wga_sandbox_option.SandboxOneshot(sandboxCfg.ImageName)
	} else {
		sandbox = wga_sandbox_option.SandboxReuse(sandboxCfg.Host)
	}

	opts := []wga_sandbox_option.Option{
		wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{
			ThreadID: req.ThreadID,
			RunID:    req.RunID,
		}),
		wga_sandbox_option.WithModelConfig(modelConfig),
		wga_sandbox_option.WithSandbox(sandbox),
		wga_sandbox_option.WithCurrentTask(req.CurrentTask),
		wga_sandbox_option.WithInstruction(req.Instruction),
		wga_sandbox_option.WithOverallTask(req.OverallTask),
		wga_sandbox_option.WithEnableThinking(req.EnableThinking),
		wga_sandbox_option.WithSkipCleanup(req.SkipCleanup),
		wga_sandbox_option.WithAgentName(req.AgentName),
		wga_sandbox_option.WithInputDir(req.InputDir),
		wga_sandbox_option.WithOutputDir(req.OutputDir),
	}

	if len(req.Messages) > 0 {
		messages := make([]wga_sandbox_option.Message, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = wga_sandbox_option.Message{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		opts = append(opts, wga_sandbox_option.WithMessages(messages))
	}

	_, outputCh, err := wga_sandbox.Run(ctx.Request.Context(), opts...)
	if err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("wga sandbox run failed: %v", err))
	}

	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[WGA][Sandbox] model %s run", req.Model.ModelId), sse_util.DONE_MSG).
		WriteStream(outputCh, nil, buildWgaSandboxLineProcessor(), nil)
	return nil
}

func WgaSandboxCleanup(ctx *gin.Context, req *request.WgaSandboxCleanupReq) error {
	if err := wga_sandbox.Cleanup(ctx.Request.Context(), req.RunID); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("wga sandbox cleanup failed: %v", err))
	}
	return nil
}

func buildWgaSandboxLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return "data: " + lineText + "\n\n", false, nil
	}
}
