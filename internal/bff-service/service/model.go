package service

import (
	"fmt"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type ModelInfoOptions struct {
	UserId string
}

func DefaultModelInfoOptions() *ModelInfoOptions {
	return &ModelInfoOptions{
		UserId: "",
	}
}

func ImportModel(ctx *gin.Context, userId, orgId string, req *request.ImportOrUpdateModelRequest) error {
	clientReq, err := parseImportAndUpdateClientReq(userId, orgId, req)
	if err != nil {
		return err
	}
	if err = ValidateModel(ctx, clientReq); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("An error occurred during model import validation: Invalid model: %v, err : %v", clientReq.Model, err))
	}
	_, err = model.ImportModel(ctx.Request.Context(), clientReq)
	if err != nil {
		return err
	}
	return nil
}

func UpdateModel(ctx *gin.Context, userId, orgId string, req *request.ImportOrUpdateModelRequest) error {
	if req.ModelId == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "modelId cannot be empty")
	}
	clientReq, err := parseImportAndUpdateClientReq(userId, orgId, req)
	if err != nil {
		return err
	}
	if err = ValidateModel(ctx, clientReq); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("An error occurred during model update validation: Invalid model: %v, err : %v", clientReq.Model, err))
	}
	_, err = model.UpdateModel(ctx, clientReq)
	if err != nil {
		return err
	}
	return nil
}

func DeleteModel(ctx *gin.Context, userId, orgId string, req *request.DeleteModelRequest) error {
	_, err := model.DeleteModel(ctx.Request.Context(), &model_service.DeleteModelReq{
		ModelId: req.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetModel(ctx *gin.Context, userId, orgId string, req *request.GetModelRequest) (*response.ModelInfo, error) {
	resp, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{
		ModelId: req.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return nil, err
	}
	return toModelInfo(ctx, resp, &ModelInfoOptions{UserId: userId})
}

func GetModelById(ctx *gin.Context, req *request.GetModelRequest) (*response.ModelInfo, error) {
	return GetModel(ctx, "", "", req)
}

func ListModels(ctx *gin.Context, userId, orgId string, req *request.ListModelsRequest) (*response.ListResult, error) {
	resp, err := model.ListModels(ctx.Request.Context(), &model_service.ListModelsReq{
		Provider:    req.Provider,
		ModelType:   req.ModelType,
		DisplayName: req.DisplayName,
		IsActive:    req.IsActive,
		UserId:      userId,
		OrgId:       orgId,
		FilterScope: req.FilterScope,
		ScopeType:   req.ScopeType,
	})
	if err != nil {
		return nil, err
	}
	list, err := toModelInfos(ctx, resp.Models, &ModelInfoOptions{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{
		List:  list,
		Total: resp.Total,
	}, nil
}

func ChangeModelStatus(ctx *gin.Context, userId, orgId string, req *request.ModelStatusRequest) error {
	_, err := model.ChangeModelStatus(ctx.Request.Context(), &model_service.ModelStatusReq{
		ModelId:  req.ModelId,
		IsActive: req.IsActive,
		UserId:   userId,
		OrgId:    orgId,
	})
	if err != nil {
		return err
	}
	return nil
}

func ListTypeModels(ctx *gin.Context, userId, orgId string, req *request.ListTypeModelsRequest) (*response.ListResult, error) {
	resp, err := model.ListTypeModels(ctx.Request.Context(), &model_service.ListTypeModelsReq{
		ModelType: req.ModelType,
		UserId:    userId,
		OrgId:     orgId,
	})
	if err != nil {
		return nil, err
	}
	list, err := toModelInfos(ctx, resp.Models, &ModelInfoOptions{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{
		List:  list,
		Total: resp.Total,
	}, nil
}

func CheckModelUserPermission(ctx *gin.Context, userId, orgId string, modelIds []string) ([]string, error) {
	resp, err := model.GetModelByIds(ctx.Request.Context(), &model_service.GetModelByIdsReq{ModelIds: modelIds})
	if err != nil {
		return nil, err
	}
	// 创建模型ID到模型信息的映射
	modelMap := make(map[string]*model_service.ModelInfo)
	for _, model := range resp.Models {
		modelMap[model.ModelId] = model
	}
	// 校验所有传入的modelIds，收集有权限的模型ID
	var authorizedModelIds []string
	var unauthorizedModelId string
	for _, modelId := range modelIds {
		model, exists := modelMap[modelId]
		if !exists {
			// 模型不存在
			unauthorizedModelId = modelId
			continue
		}
		// 校验模型权限
		var hasPermission bool
		switch model.GetScopeType() {
		case config.ModelScopeTypePrivate: // 私有
			hasPermission = (model.UserId == userId) && (model.OrgId == orgId)
		case config.ModelScopeTypePublic: // 公开
			hasPermission = true // 公开模型，任何人都可以访问
		case config.ModelScopeTypeOrg: // 指定组织可见
			hasPermission = (model.OrgId == orgId)
		default:
			hasPermission = (model.UserId == userId) && (model.OrgId == orgId)
		}
		if hasPermission {
			authorizedModelIds = append(authorizedModelIds, modelId)
		} else {
			unauthorizedModelId = modelId
		}
	}

	if unauthorizedModelId != "" {
		return authorizedModelIds, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_model_perm", unauthorizedModelId)
	}
	return authorizedModelIds, nil
}

func GetModelIdByUuid(ctx *gin.Context, uuid string) (string, error) {
	resp, err := model.GetModelByUuid(ctx, &model_service.GetModelByUuidReq{Uuid: uuid})
	if err != nil {
		return "", err
	}
	return resp.ModelId, nil
}

// --- internal ---

func parseImportAndUpdateClientReq(userId, orgId string, req *request.ImportOrUpdateModelRequest) (*model_service.ModelInfo, error) {
	if req.ScopeType == config.ModelScopeTypePublic {
		if userId != config.SystemAdminUserID || orgId != config.TopOrgID {
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "Only system administrators can make the model public")
		}
	}
	importSource := req.ImportSource
	if importSource == "" {
		importSource = "external"
	}
	clientReq := &model_service.ModelInfo{
		Provider:      req.Provider,
		ModelId:       req.ModelId,
		ModelType:     req.ModelType,
		Model:         req.Model,
		DisplayName:   req.DisplayName,
		ModelIconPath: req.Avatar.Key,
		PublishDate:   req.PublishDate,
		UserId:        userId,
		OrgId:         orgId,
		IsActive:      true,
		ModelDesc:     req.ModelDesc,
		ScopeType:     req.ScopeType,
		ImportSource:  importSource,
	}
	configStr, err := req.ConfigString()
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	clientReq.ProviderConfig = configStr
	return clientReq, nil
}

func toModelInfos(ctx *gin.Context, models []*model_service.ModelInfo, opts ...*ModelInfoOptions) ([]*response.ModelInfo, error) {
	var ret []*response.ModelInfo
	for _, m := range models {
		info, err := toModelInfo(ctx, m, opts...)
		if err != nil {
			return nil, err
		}
		ret = append(ret, info)
	}
	return ret, nil
}

func toModelInfo(ctx *gin.Context, modelInfo *model_service.ModelInfo, opts ...*ModelInfoOptions) (*response.ModelInfo, error) {
	modelConfig, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v get model config err: %v", modelInfo.ModelId, err))
	}

	// 获取模型全量标签（独立函数）
	allModelTags, err := getModelAllTags(modelInfo)
	if err != nil {
		return nil, err
	}

	// 判断模型是否支持 编辑
	option := DefaultModelInfoOptions()
	if len(opts) > 0 && opts[0] != nil {
		if opts[0].UserId != "" {
			option.UserId = opts[0].UserId
		}
	}
	allowEdit := modelInfo.UserId == option.UserId

	res := &response.ModelInfo{
		ModelId:      modelInfo.ModelId,
		Uuid:         modelInfo.Uuid,
		Provider:     modelInfo.Provider,
		Model:        modelInfo.Model,
		ModelType:    modelInfo.ModelType,
		DisplayName:  modelInfo.DisplayName,
		Avatar:       CacheAvatar(ctx, modelInfo.ModelIconPath, true),
		PublishDate:  modelInfo.PublishDate,
		IsActive:     modelInfo.IsActive,
		UserId:       modelInfo.UserId,
		OrgId:        modelInfo.OrgId,
		CreatedAt:    util.Time2Str(modelInfo.CreatedAt),
		UpdatedAt:    util.Time2Str(modelInfo.UpdatedAt),
		ModelDesc:    modelInfo.ModelDesc,
		Config:       modelConfig,
		Tags:         allModelTags,
		ScopeType:    modelInfo.ScopeType,
		AllowEdit:    allowEdit,
		ImportSource: modelInfo.ImportSource,
	}
	if res.DisplayName == "" {
		res.DisplayName = res.Model
	}
	return res, nil
}

// getModelAllTags 获取模型所有标签
func getModelAllTags(modelInfo *model_service.ModelInfo) ([]mp_common.Tag, error) {
	var allModelTags []mp_common.Tag

	// - 公开范围标签：如私有/公开/内部测试等
	scopeTags := mp_common.GetTagsByScopeType(modelInfo.ScopeType)
	allModelTags = append(allModelTags, scopeTags...)

	// - 导入来源标签：如外部URL/内置模型等
	sourceTags := mp_common.GetTagsByImportSource(modelInfo.ImportSource)
	allModelTags = append(allModelTags, sourceTags...)

	// - 基础属性标签：如模型类型(LLM/rerank)、最大token数、推理能力等
	baseTags, err := mp.ToModelTags(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get model base tags, modelId: %v, err: %v", modelInfo.ModelId, err)
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, errMsg)
	}
	allModelTags = append(allModelTags, baseTags...)

	return allModelTags, nil
}
