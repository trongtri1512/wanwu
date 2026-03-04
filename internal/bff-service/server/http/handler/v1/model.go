package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/gin-gonic/gin"
)

// ImportModel
//
//	@Tags			model
//	@Summary		模型导入
//	@Description	第三方模型的导入
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ImportOrUpdateModelRequest	true	"导入模型信息"
//	@Success		200		{object}	response.Response
//	@Router			/model [post]
func ImportModel(ctx *gin.Context) {
	var req request.ImportOrUpdateModelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ImportModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateModel
//
//	@Tags		model
//	@Summary	导入模型更新
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.ImportOrUpdateModelRequest	true	"模型变更信息"
//	@Success	200		{object}	response.Response
//	@Router		/model [put]
func UpdateModel(ctx *gin.Context) {
	var req request.ImportOrUpdateModelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteModel
//
//	@Tags		model
//	@Summary	导入模型删除
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.DeleteModelRequest	true	"模型删除key"
//	@Success	200		{object}	response.Response
//	@Router		/model [delete]
func DeleteModel(ctx *gin.Context) {
	var req request.DeleteModelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// GetModel
//
//	@Tags		model
//	@Summary	‌查询单个模型
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.GetModelRequest	true	"模型ID"
//	@Success	200		{object}	response.Response{data=response.ModelInfo}
//	@Router		/model [get]
func GetModel(ctx *gin.Context) {
	var req request.GetModelRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetModel(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ListModels
//
//	@Tags		model
//	@Summary	导入模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		modelType	query		string	false	"模型类型"	Enums(llm,embedding,rerank)
//	@Param		provider	query		string	false	"模型供应商"
//	@Param		displayName	query		string	false	"模型显示名称"
//	@Param		isActive	query		string	false	"启用状态（true: 启用）"
//	@Param		filterScope	query		string	false	"模型作用域类型: public,private"
//	@Param		scopeType	query		string	false	"模型公开范围: 1,2,3"
//	@Success	200			{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/list [get]
func ListModels(ctx *gin.Context) {
	var req request.ListModelsRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListModels(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeModelStatus
//
//	@Tags		model
//	@Summary	模型启用/关闭
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.ModelStatusRequest	true	"启用/关闭 的模型信息"
//	@Success	200		{object}	response.Response
//	@Router		/model/status [put]
func ChangeModelStatus(ctx *gin.Context) {
	var req request.ModelStatusRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ChangeModelStatus(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// ListLlmModels
//
//	@Tags		model
//	@Summary	llm模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/llm [get]
func ListLlmModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeLLM,
	})
	gin_util.Response(ctx, resp, err)
}

// ListRerankModels
//
//	@Tags		model
//	@Summary	rerank模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/rerank [get]
func ListRerankModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeTextRerank,
	})
	gin_util.Response(ctx, resp, err)
}

// ListMultiRerankModels
//
//	@Tags		model
//	@Summary	多模态rerank模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/multi-rerank [get]
func ListMultiRerankModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeMultiRerank,
	})
	gin_util.Response(ctx, resp, err)
}

// ListEmbeddingModels
//
//	@Tags		model
//	@Summary	embedding模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/embedding [get]
func ListEmbeddingModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeTextEmbedding,
	})
	gin_util.Response(ctx, resp, err)
}

// ListMultiEmbeddingModels
//
//	@Tags		model
//	@Summary	多模态embedding模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/multi-embedding [get]
func ListMultiEmbeddingModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeMultiEmbedding,
	})
	gin_util.Response(ctx, resp, err)
}

// ListOcrModels
//
//	@Tags		model
//	@Summary	ocr模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/ocr [get]
func ListOcrModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeOcr,
	})
	gin_util.Response(ctx, resp, err)
}

// ListPdfParserModels
//
//	@Tags		model
//	@Summary	pdf文档解析模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/pdf-parser [get]
func ListPdfParserModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypePdfParser,
	})
	gin_util.Response(ctx, resp, err)
}

// ListGuiModels
//
//	@Tags		model
//	@Summary	gui模型列表
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/gui [get]
func ListGuiModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeGui,
	})
	gin_util.Response(ctx, resp, err)
}

// ListSyncAsrModels
//
//	@Tags		model
//	@Summary	同步asr语音识别模型列表展示
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/asr [get]
func ListSyncAsrModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeSyncAsr,
	})
	gin_util.Response(ctx, resp, err)
}

// ListText2ImageModels
//
//	@Tags		model
//	@Summary	文生图模型列表展示
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Response{data=response.ListResult{list=response.ModelInfo}}
//	@Router		/model/select/text2image [get]
func ListText2ImageModels(ctx *gin.Context) {
	resp, err := service.ListTypeModels(ctx, getUserID(ctx), getOrgID(ctx), &request.ListTypeModelsRequest{
		ModelType: mp.ModelTypeText2Image,
	})
	gin_util.Response(ctx, resp, err)
}
