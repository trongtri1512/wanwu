package service

import (
	"fmt"
	"net/http"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func ModelMultiModalRerank(ctx *gin.Context, modelID string, req *mp_common.MultiModalRerankReq) {
	// modelInfo by modelID
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelID})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if !modelInfo.IsActive {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFModelStatus, modelInfo.ModelId))
		return
	}

	// 校验model字段
	if req != nil {
		if req.Model != "" && req.Model != modelInfo.Model {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v multiModalRerank err: model mismatch!", modelInfo.ModelId)))
			return
		}
	}

	// jina 传参为不带前缀base64
	if modelInfo.Provider == mp.ProviderJina {
		for i := range req.Documents {
			if req.Documents[i].Image != "" {
				req.Documents[i].Image, _ = util.CheckAndRemoveBase64Prefix(req.Documents[i].Image)
			}
		}
	}
	// multiModalRerank config
	multiModalRerank, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v multiModalRerank err: %v", modelInfo.ModelId, err)))
		return
	}
	iMultiModalRerank, ok := multiModalRerank.(mp.IMultiModalRerank)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v multiModalRerank err: invalid provider", modelInfo.ModelId)))
		return
	}
	// multiModalRerank
	multiModalRerankReq, err := iMultiModalRerank.NewReq(req)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v multiModalRerank NewReq err: %v", modelInfo.ModelId, err)))
		return
	}
	startTime := time.Now()
	resp, err := iMultiModalRerank.MultiModalRerank(ctx.Request.Context(), multiModalRerankReq)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v multiModalRerank err: %v", modelInfo.ModelId, err)))
		return
	}
	if data, ok := resp.ConvertResp(); ok {
		if data.Model == "" {
			data.Model = modelInfo.Model
		}
		status := http.StatusOK
		ctx.Set(gin_util.STATUS, status)
		//ctx.Set(config.RESULT, resp.String())
		ctx.JSON(status, data)
		costs := int(time.Since(startTime).Milliseconds())
		recordModelStatistic(ctx, modelInfo, true,
			data.Usage.PromptTokens, data.Usage.CompletionTokens, data.Usage.TotalTokens, costs, 0, false)
		return
	}
	recordModelStatistic(ctx, modelInfo, false, 0, 0, 0, 0, 0, false)
	gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v multiModalRerank err: invalid resp", modelInfo.ModelId)))
}
