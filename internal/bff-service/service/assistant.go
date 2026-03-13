package service

import (
	"encoding/json"
	"sort"
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	bff_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func AssistantCreate(ctx *gin.Context, userId, orgId string, req request.AssistantCreateReq) (*response.AssistantCreateResp, error) {
	resp, err := assistant.AssistantCreate(ctx.Request.Context(), &assistant_service.AssistantCreateReq{
		AssistantBrief: assistantBriefConfigModel2Proto(req),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		Category: int32(req.Category),
	})
	if err != nil {
		return nil, err
	}
	return &response.AssistantCreateResp{
		AssistantId: resp.AssistantId,
	}, nil
}

func AssistantUpdate(ctx *gin.Context, userId, orgId string, req request.AssistantBrief) (interface{}, error) {
	_, err := assistant.AssistantUpdate(ctx.Request.Context(), &assistant_service.AssistantUpdateReq{
		AssistantId:    req.AssistantId,
		AssistantBrief: appBriefConfigModel2Proto(req.AppBriefConfig),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return nil, err
}

func AssistantConfigUpdate(ctx *gin.Context, userId, orgId string, req request.AssistantConfig) (interface{}, error) {
	var modelConfig, rerankConfig *common.AppModelConfig
	var err error
	if req.ModelConfig != nil {
		if req.ModelConfig.ModelId == "" {
			// 如果 modelId 为空，则认为用户想要清空模型配置，构造一个空的 AppModelConfig
			modelConfig = &common.AppModelConfig{}
		} else {
			modelConfig, err = appModelConfigModel2Proto(*req.ModelConfig)
			if err != nil {
				return nil, err
			}
		}
	}
	if req.RerankConfig != nil {
		if req.RerankConfig.ModelId == "" {
			// 如果 modelId 为空，则认为用户想要清空模型配置，构造一个空的 AppModelConfig
			rerankConfig = &common.AppModelConfig{}
		} else {
			rerankConfig, err = appModelConfigModel2Proto(*req.RerankConfig)
			if err != nil {
				return nil, err
			}
		}
	}
	var recommendConfig *assistant_service.AssistantRecommendConfig
	if req.RecommendConfig != nil {
		recommendConfig, err = recommendConfigModel2Proto(*req.RecommendConfig)
		if err != nil {
			return nil, err
		}
	}
	var safetyConfig *assistant_service.AssistantSafetyConfig
	if req.SafetyConfig != nil {
		safetyConfig = &assistant_service.AssistantSafetyConfig{
			Enable:         req.SafetyConfig.Enable,
			SensitiveTable: transSafetyConfig2Proto(req.SafetyConfig.Tables),
		}
	}

	var visionConfig *assistant_service.AssistantVisionConfig
	if req.VisionConfig != nil {
		visionConfig = &assistant_service.AssistantVisionConfig{
			PicNum: req.VisionConfig.PicNum,
		}
	}

	var memoryConfig *assistant_service.AssistantMemoryConfig
	if req.MemoryConfig != nil {
		memoryConfig = &assistant_service.AssistantMemoryConfig{
			MaxHistoryLength: req.MemoryConfig.MaxHistoryLength,
		}
	}

	var knowledgeBaseConfig *assistant_service.AssistantKnowledgeBaseConfig
	if req.KnowledgeBaseConfig != nil {
		knowledgeBaseConfig = transKnowledgebases2Proto(*req.KnowledgeBaseConfig)
	}

	_, err = assistant.AssistantConfigUpdate(ctx.Request.Context(), &assistant_service.AssistantConfigUpdateReq{
		AssistantId:         req.AssistantId,
		Prologue:            req.Prologue,
		Instructions:        req.Instructions,
		RecommendQuestion:   req.RecommendQuestion,
		ModelConfig:         modelConfig,
		KnowledgeBaseConfig: knowledgeBaseConfig,
		RerankConfig:        rerankConfig,
		SafetyConfig:        safetyConfig,
		VisionConfig:        visionConfig,
		MemoryConfig:        memoryConfig,
		RecommendConfig:     recommendConfig,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})

	return nil, err
}

func GetAssistantInfo(ctx *gin.Context, userId, orgId string, req request.AssistantIdRequest, needPublished bool) (*response.Assistant, error) {
	var resp *assistant_service.AssistantInfo
	var err error
	if needPublished {
		resp, err = assistant.AssistantSnapshotInfo(ctx.Request.Context(), &assistant_service.AssistantSnapshotInfoReq{
			AssistantId: req.AssistantId,
			Version:     req.Version,
		})
	} else {
		resp, err = assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
			AssistantId: req.AssistantId,
			Identity: &assistant_service.Identity{ //草稿只能看自己的
				UserId: userId,
				OrgId:  orgId,
			},
		})
	}
	if err != nil {
		return nil, err
	}
	return transAssistantResp2Model(ctx, resp)
}

func GetAssistantIdByUuid(ctx *gin.Context, uuid string) (string, error) {
	resp, err := assistant.GetAssistantIdByUuid(ctx.Request.Context(), &assistant_service.GetAssistantIdByUuidReq{
		Uuid: uuid,
	})
	if err != nil {
		return "", err
	}
	return resp.AssistantId, nil
}

func AssistantCopy(ctx *gin.Context, userId, orgId string, req request.AssistantIdRequest) (*response.AssistantCreateResp, error) {
	resp, err := assistant.AssistantCopy(ctx.Request.Context(), &assistant_service.AssistantCopyReq{
		AssistantId: req.AssistantId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.AssistantCreateResp{
		AssistantId: resp.AssistantId,
	}, nil
}

func AssistantWorkFlowCreate(ctx *gin.Context, userId, orgId string, req request.AssistantWorkFlowAddRequest) error {
	_, err := assistant.AssistantWorkFlowCreate(ctx.Request.Context(), &assistant_service.AssistantWorkFlowCreateReq{
		AssistantId: req.AssistantId,
		WorkFlowId:  req.WorkFlowId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantWorkFlowDelete(ctx *gin.Context, userId, orgId string, req request.AssistantWorkFlowDelRequest) error {
	_, err := assistant.AssistantWorkFlowDelete(ctx.Request.Context(), &assistant_service.AssistantWorkFlowDeleteReq{
		AssistantId: req.AssistantId,
		WorkFlowId:  req.WorkFlowId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantWorkFlowEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantWorkFlowToolEnableRequest) error {
	_, err := assistant.AssistantWorkFlowEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantWorkFlowEnableSwitchReq{
		AssistantId: req.AssistantId,
		WorkFlowId:  req.WorkFlowId,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantMCPCreate(ctx *gin.Context, userId, orgId string, req request.AssistantMCPToolAddRequest) error {
	_, err := assistant.AssistantMCPCreate(ctx.Request.Context(), &assistant_service.AssistantMCPCreateReq{
		AssistantId: req.AssistantId,
		McpId:       req.MCPId,
		McpType:     req.MCPType,
		ActionName:  req.ActionName,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantMCPDelete(ctx *gin.Context, userId, orgId string, req request.AssistantMCPToolDelRequest) error {
	_, err := assistant.AssistantMCPDelete(ctx.Request.Context(), &assistant_service.AssistantMCPDeleteReq{
		AssistantId: req.AssistantId,
		McpId:       req.MCPId,
		McpType:     req.MCPType,
		ActionName:  req.ActionName,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantMCPEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantMCPToolEnableRequest) error {
	_, err := assistant.AssistantMCPEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantMCPEnableSwitchReq{
		AssistantId: req.AssistantId,
		McpId:       req.MCPId,
		McpType:     req.MCPType,
		ActionName:  req.ActionName,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantToolCreate(ctx *gin.Context, userId, orgId string, req request.AssistantToolAddRequest) error {
	_, err := assistant.AssistantToolCreate(ctx.Request.Context(), &assistant_service.AssistantToolCreateReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolType:    req.ToolType,
		ActionName:  req.ActionName,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantToolDelete(ctx *gin.Context, userId, orgId string, req request.AssistantToolDelRequest) error {
	_, err := assistant.AssistantToolDelete(ctx.Request.Context(), &assistant_service.AssistantToolDeleteReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolType:    req.ToolType,
		ActionName:  req.ActionName,
	})
	return err
}

func AssistantToolEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantToolEnableRequest) error {
	_, err := assistant.AssistantToolEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantToolEnableSwitchReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolType:    req.ToolType,
		ActionName:  req.ActionName,
		Enable:      req.Enable,
	})
	return err
}

func AssistantToolConfig(ctx *gin.Context, userId, orgId string, req request.AssistantToolConfigRequest) error {
	toolConfigJSON, err := json.Marshal(req.ToolConfig)
	if err != nil {
		return err
	}
	_, err = assistant.AssistantToolConfig(ctx.Request.Context(), &assistant_service.AssistantToolConfigReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolConfig:  string(toolConfigJSON),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantSkillCreate(ctx *gin.Context, userId, orgId string, req request.AssistantSkillAddRequest) error {
	_, err := assistant.AssistantSkillCreate(ctx.Request.Context(), &assistant_service.AssistantSkillCreateReq{
		AssistantId: req.AssistantId,
		SkillId:     req.SkillId,
		SkillType:   req.SkillType,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantSkillDelete(ctx *gin.Context, userId, orgId string, req request.AssistantSkillDelRequest) error {
	_, err := assistant.AssistantSkillDelete(ctx.Request.Context(), &assistant_service.AssistantSkillDeleteReq{
		AssistantId: req.AssistantId,
		SkillId:     req.SkillId,
		SkillType:   req.SkillType,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantSkillEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantSkillEnableSwitchRequest) error {
	_, err := assistant.AssistantSkillEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantSkillEnableSwitchReq{
		AssistantId: req.AssistantId,
		SkillId:     req.SkillId,
		SkillType:   req.SkillType,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentCreate(ctx *gin.Context, userId, orgId string, req request.MultiAgentCreateReq) error {
	_, err := assistant.MultiAgentCreate(ctx.Request.Context(), &assistant_service.MultiAgentCreateReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentDelete(ctx *gin.Context, userId, orgId string, req request.MultiAgentCreateReq) error {
	_, err := assistant.MultiAgentDelete(ctx.Request.Context(), &assistant_service.MultiAgentCreateReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentEnableSwitch(ctx *gin.Context, userId, orgId string, req request.MultiAgentEnableSwitchReq) error {
	_, err := assistant.MultiAgentEnableSwitch(ctx.Request.Context(), &assistant_service.MultiAgentEnableSwitchReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentConfigUpdate(ctx *gin.Context, userId, orgId string, req request.MultiAgentConfigUpdateReq) error {
	_, err := assistant.MultiAgentConfigUpdate(ctx.Request.Context(), &assistant_service.MultiAgentConfigUpdateReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Desc:        req.Desc,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func GetAssistantSelect(ctx *gin.Context, userId, orgId string, req request.GetExplorationAppListRequest) (*response.ListResult, error) {
	wlist, err := GetExplorationAppList(ctx, userId, orgId, request.GetExplorationAppListRequest{
		Name:       req.Name,
		AppType:    constant.AppTypeAgent,
		SearchType: "all",
	})
	if err != nil {
		return nil, err
	}
	var appList []*response.ExplorationAppInfo
	if wlistSlice, ok := wlist.List.([]*response.ExplorationAppInfo); ok {
		for _, w := range wlistSlice {
			if w.User.UserId == userId {
				appList = append(appList, w)
			}
		}
	}
	return &response.ListResult{
		List:  appList,
		Total: int64(len(appList)),
	}, nil
}

func assistantModelConvert(ctx *gin.Context, modelConfigInfo *common.AppModelConfig) (modelConfig request.AppModelConfig, err error) {
	if modelConfigInfo != nil && modelConfigInfo.ModelId != "" {
		log.Debugf("检测到模型配置，模型ID: %s", modelConfigInfo.ModelId)
		modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelConfigInfo.ModelId})
		if err != nil {
			log.Errorf("获取模型信息失败，模型ID: %s, 错误: %v", modelConfigInfo.ModelId, err)
		}
		if modelInfo != nil {
			modelConfig, err = appModelConfigProto2Model(modelConfigInfo, modelInfo.DisplayName)
			if err != nil {
				log.Errorf("模型配置Proto转换到模型失败，模型ID: %s, 错误: %v", modelConfigInfo.ModelId, err)
				return modelConfig, err
			}
			log.Debugf("模型配置转换成功: %+v", modelConfig)
		}
	} else {
		log.Debugf("模型配置为空或模型ID为空")
	}
	return modelConfig, nil
}

func assistantRerankConvert(ctx *gin.Context, rerankConfigInfo *common.AppModelConfig) (request.AppModelConfig, error) {
	var rerankConfig request.AppModelConfig
	if rerankConfigInfo != nil && rerankConfigInfo.ModelId != "" {
		log.Debugf("检测到Rerank配置，模型ID: %s", rerankConfigInfo.ModelId)
		modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: rerankConfigInfo.ModelId})
		if err != nil {
			log.Errorf("获取Rerank模型信息失败，模型ID: %s, 错误: %v", rerankConfigInfo.ModelId, err)
		} else {
			rerankConfig, err = appModelConfigProto2Model(rerankConfigInfo, modelInfo.DisplayName)
			if err != nil {
				log.Errorf("Rerank配置Proto转换到模型失败，模型ID: %s, 错误: %v", rerankConfigInfo.ModelId, err)
				return rerankConfig, err
			}
			log.Debugf("Rerank配置转换成功: %+v", rerankConfig)
		}
	} else {
		log.Debugf("Rerank配置为空或模型ID为空")
	}
	return rerankConfig, nil
}

func assistantWorkFlowConvert(ctx *gin.Context, workFlowInfos []*assistant_service.AssistantWorkFlowInfos) ([]*response.AssistantWorkFlowInfo, error) {
	var assistantWorkFlowInfos []*response.AssistantWorkFlowInfo
	if len(workFlowInfos) > 0 {
		var workflowIds []string
		for _, wf := range workFlowInfos {
			workflowIds = append(workflowIds, wf.WorkFlowId)
		}
		cozeWorkflowList, err := ListWorkflowByIDs(ctx, "", workflowIds)
		if err != nil {
			return nil, err
		}
		for _, wf := range workFlowInfos {
			workFlowInfo := &response.AssistantWorkFlowInfo{
				WorkFlowId: wf.WorkFlowId,
				ApiName:    wf.ApiName,
				Enable:     wf.Enable,
				UniqueId:   bff_util.ConcatAssistantToolUniqueId("workflow", wf.WorkFlowId),
			}

			for _, info := range cozeWorkflowList.Workflows {
				if info.WorkflowId == wf.WorkFlowId {
					// 找到匹配的工作流，设置名称和描述
					workFlowInfo.WorkFlowName = info.Name
					workFlowInfo.WorkFlowDesc = info.Desc
					workFlowInfo.AvatarPath = cacheWorkflowAvatar(info.URL, constant.AppTypeWorkflow)
				}
			}

			// 仅当工作流名称非空时才添加
			if workFlowInfo.WorkFlowName != "" {
				assistantWorkFlowInfos = append(assistantWorkFlowInfos, workFlowInfo)
				log.Debugf("添加工作流信息: WorkFlowId=%s, ApiName=%s", wf.WorkFlowId, wf.ApiName)
			}
		}
		log.Debugf("总共添加 %d 个工作流信息", len(assistantWorkFlowInfos))
	} else {
		log.Debugf("工作流信息为空")
	}
	return assistantWorkFlowInfos, nil
}

func assistantMCPConvert(ctx *gin.Context, assistantMCPInfos []*assistant_service.AssistantMCPInfos) ([]*response.AssistantMCPInfo, error) {
	// 若查询结果为空，返回空列表
	if len(assistantMCPInfos) == 0 {
		return nil, nil
	}

	// 提取MCP ID列表
	var MCPCustomIds, MCPServerIds []string
	for _, m := range assistantMCPInfos {
		switch m.McpType {
		case constant.MCPTypeMCP:
			MCPCustomIds = append(MCPCustomIds, m.McpId)
		case constant.MCPTypeMCPServer:
			MCPServerIds = append(MCPServerIds, m.McpId)
		}
	}

	// 批量查询MCP详情
	mcpResp, err := mcp.GetMCPByMCPIdList(ctx.Request.Context(), &mcp_service.GetMCPByMCPIdListReq{
		McpIdList:       MCPCustomIds,
		McpServerIdList: MCPServerIds,
	})

	// 构建MCP详情映射
	mcpDetailMap := make(map[string]*mcp_service.CustomMCPInfo)
	if err == nil && mcpResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range mcpResp.Infos {
			mcpDetailMap[item.McpId] = item
		}
	}
	// 构建MCPServer详情映射
	mcpserverDetailMap := make(map[string]*mcp_service.MCPServerInfo)
	if err == nil && mcpResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range mcpResp.Servers {
			mcpserverDetailMap[item.McpServerId] = item
		}
	}

	// 构建返回结果
	var retMCPInfos []*response.AssistantMCPInfo
	for _, info := range assistantMCPInfos {
		var exists bool
		var mcpName string
		var avatar request.Avatar

		switch info.McpType {
		case constant.MCPTypeMCP:
			if item, ok := mcpDetailMap[info.McpId]; ok {
				exists = true
				mcpName = item.Info.Name
				avatar = cacheMCPAvatar(ctx, item.Info.AvatarPath, item.AvatarPath)
			}
		case constant.MCPTypeMCPServer:
			if item, ok := mcpserverDetailMap[info.McpId]; ok {
				exists = true
				mcpName = item.Name
				avatar = cacheMCPServerAvatar(ctx, item.AvatarPath)
			}
		}

		if exists {
			retMCPInfos = append(retMCPInfos, &response.AssistantMCPInfo{
				UniqueId:   bff_util.ConcatAssistantToolUniqueId(info.McpType, info.McpId),
				MCPId:      info.McpId,
				MCPType:    info.McpType,
				MCPName:    mcpName,
				ActionName: info.ActionName,
				Enable:     info.Enable,
				Valid:      true,
				Avatar:     avatar,
			})
		}
	}

	return retMCPInfos, nil
}

func assistantToolsConvert(ctx *gin.Context, assistantToolInfos []*assistant_service.AssistantToolInfos) ([]*response.AssistantToolInfo, error) {
	// 若查询为空，返回空列表
	if len(assistantToolInfos) == 0 {
		return nil, nil
	}

	// 提取工具ID列表
	var customToolIds, builtinToolIds []string
	for _, tool := range assistantToolInfos {
		switch tool.ToolType {
		case constant.ToolTypeCustom:
			customToolIds = append(customToolIds, tool.ToolId)
		case constant.ToolTypeBuiltIn:
			builtinToolIds = append(builtinToolIds, tool.ToolId)
		}
	}

	// 批量查询
	toolInfoResp, err := mcp.GetToolByIdList(ctx.Request.Context(), &mcp_service.GetToolByToolIdListReq{
		BuiltInToolIdList: builtinToolIds,
		CustomToolIdList:  customToolIds,
	})

	// 构建ID到工具信息的映射
	customToolMap := make(map[string]*mcp_service.GetCustomToolItem)
	if err == nil && toolInfoResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range toolInfoResp.List {
			customToolMap[item.CustomToolId] = item
		}
	}
	builtinToolMap := make(map[string]*mcp_service.ToolSquareInfo)
	if err == nil && toolInfoResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range toolInfoResp.ToolSquareInfoList {
			builtinToolMap[item.ToolSquareId] = item
		}
	}

	// 组装返回结果
	var retToolInfos []*response.AssistantToolInfo
	for _, info := range assistantToolInfos {
		var exists bool
		var toolName string
		var avatar request.Avatar

		switch info.ToolType {
		case constant.ToolTypeCustom:
			if item, ok := customToolMap[info.ToolId]; ok {
				exists = true
				toolName = item.Name
				avatar = cacheToolAvatar(ctx, constant.ToolTypeCustom, item.AvatarPath)
			}
		case constant.ToolTypeBuiltIn:
			if item, ok := builtinToolMap[info.ToolId]; ok {
				exists = true
				toolName = item.Name
				avatar = cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, item.AvatarPath)
			}
		}

		if exists {
			var toolConfig request.AssistantToolConfig
			if info.ToolConfig != "" {
				if err := json.Unmarshal([]byte(info.ToolConfig), &toolConfig); err != nil {
					log.Warnf("解析ToolConfig失败，使用空配置，error: %v, toolConfig: %s", err, info.ToolConfig)
				}
			}
			retToolInfos = append(retToolInfos, &response.AssistantToolInfo{
				UniqueId:   bff_util.ConcatAssistantToolUniqueId(info.ToolType, info.ToolId),
				ToolId:     info.ToolId,
				ToolType:   info.ToolType,
				ToolName:   toolName,
				ActionName: info.ActionName,
				Enable:     info.Enable,
				Valid:      true,
				ToolConfig: toolConfig,
				Avatar:     avatar,
			})
		}
	}
	return retToolInfos, nil

}

func assistantSkillConvert(ctx *gin.Context, assistantSkillInfos []*assistant_service.AssistantSkillInfo) ([]*response.AssistantSkillInfo, error) {
	if len(assistantSkillInfos) == 0 {
		return nil, nil
	}

	var customSkillIds []string
	for _, skill := range assistantSkillInfos {
		if skill.SkillType == constant.SkillTypeCustom {
			customSkillIds = append(customSkillIds, skill.SkillId)
		}
	}

	// 获取自定义技能
	customSkillResp, err := mcp.GetCustomSkillDetailByIdList(ctx.Request.Context(), &mcp_service.CustomSkillDetailByIdListReq{
		SkillIds: customSkillIds,
	})
	customSkillMap := make(map[string]*mcp_service.CustomSkill)
	if err == nil && customSkillResp != nil {
		for _, item := range customSkillResp.SkillDetails {
			customSkillMap[item.SkillId] = item
		}
	}

	var retSkillInfos []*response.AssistantSkillInfo
	for _, info := range assistantSkillInfos {
		var exists bool
		var skillName, author string
		var avatar request.Avatar

		switch info.SkillType {
		case constant.SkillTypeCustom:
			if item, ok := customSkillMap[info.SkillId]; ok {
				exists = true
				skillName = item.Name
				author = item.Author
				avatar = cacheSkillAvatar(ctx, item.Avatar)
			}
		case constant.SkillTypeBuiltIn:
			skillDetail, err := GetAgentSkillDetail(ctx, info.SkillId)
			if err == nil && skillDetail != nil {
				exists = true
				skillName = skillDetail.Name
				author = skillDetail.Author
				avatar = skillDetail.Avatar
			}
		}

		if exists {
			retSkillInfos = append(retSkillInfos, &response.AssistantSkillInfo{
				SkillId:   info.SkillId,
				SkillType: info.SkillType,
				SkillName: skillName,
				Author:    author,
				Enable:    info.Enable,
				Valid:     true,
				Avatar:    avatar,
			})
		}
	}
	return retSkillInfos, nil
}

func assistantSafetyConvert(ctx *gin.Context, resp *assistant_service.AssistantSafetyConfig) (request.AppSafetyConfig, error) {
	var exists bool
	enable := resp.GetEnable()
	var sensitiveTableList []request.SensitiveTable

	if len(resp.GetSensitiveTable()) != 0 {
		var tableIds []string
		for _, table := range resp.GetSensitiveTable() {
			tableIds = append(tableIds, table.TableId)
		}
		sensitiveWordTable, err := safety.GetSensitiveWordTableListByIDs(ctx, &safety_service.GetSensitiveWordTableListByIDsReq{TableIds: tableIds})

		if err == nil && sensitiveWordTable != nil {
			exists = true
		} else {
			enable = false
		}

		if exists {
			for _, table := range sensitiveWordTable.List {
				sensitiveTableList = append(sensitiveTableList, request.SensitiveTable{
					TableId:   table.TableId,
					TableName: table.TableName,
				})
			}
		}
	}

	if len(sensitiveTableList) == 0 {
		enable = false
	}

	return request.AppSafetyConfig{
		Enable: enable,
		Tables: sensitiveTableList,
	}, nil
}

func ConversationCreate(ctx *gin.Context, userId, orgId string, req request.ConversationCreateRequest, conversationType string) (response.ConversationCreateResp, error) {
	resp, err := assistant.ConversationCreate(ctx.Request.Context(), &assistant_service.ConversationCreateReq{
		AssistantId:      req.AssistantId,
		Prompt:           req.Prompt,
		ConversationType: conversationType,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return response.ConversationCreateResp{}, err
	}
	return response.ConversationCreateResp{
		ConversationId: resp.ConversationId,
	}, nil
}

func ConversationDelete(ctx *gin.Context, userId, orgId string, req request.ConversationIdRequest) (interface{}, error) {
	_, err := assistant.ConversationDelete(ctx.Request.Context(), &assistant_service.ConversationDeleteReq{
		ConversationId: req.ConversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func GetDraftConversationIdByAssistantID(ctx *gin.Context, userId, orgId string, req request.ConversationGetListRequest) (*response.ConversationIdResp, error) {
	resp, err := assistant.GetConversationIdByAssistantId(ctx.Request.Context(), &assistant_service.GetConversationIdByAssistantIdReq{
		AssistantId:      req.AssistantId,
		ConversationType: constant.ConversationTypeDraft,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.ConversationIdResp{
		ConversationId: resp.ConversationId,
	}, nil
}

func DraftConversationDeleteByAssistantID(ctx *gin.Context, userId, orgId string, req request.ConversationDeleteRequest) (interface{}, error) {
	// 获取 conversation_id
	conversationIdResp, err := assistant.GetConversationIdByAssistantId(ctx.Request.Context(), &assistant_service.GetConversationIdByAssistantIdReq{
		AssistantId:      req.AssistantId,
		ConversationType: constant.ConversationTypeDraft,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})

	if conversationIdResp == nil || err != nil {
		return nil, err
	}

	// 删除草稿会话
	_, err = assistant.ConversationDelete(ctx.Request.Context(), &assistant_service.ConversationDeleteReq{
		ConversationId: conversationIdResp.ConversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func GetConversationList(ctx *gin.Context, userId, orgId string, req request.ConversationGetListRequest) (response.PageResult, error) {
	resp, err := assistant.GetConversationList(ctx.Request.Context(), &assistant_service.GetConversationListReq{
		AssistantId:      req.AssistantId,
		ConversationType: constant.ConversationTypePublished,
		PageSize:         int32(req.PageSize),
		PageNo:           int32(req.PageNo),
		Identity: &assistant_service.Identity{
			UserId: userId,
		},
	})
	if err != nil {
		return response.PageResult{}, err
	}
	return response.PageResult{Total: resp.Total, List: resp.Data}, nil
}

func GetConversationDetailList(ctx *gin.Context, userId, orgId string, req request.ConversationGetDetailListRequest) (response.PageResult, error) {
	resp, err := assistant.GetConversationDetailList(ctx.Request.Context(), &assistant_service.GetConversationDetailListReq{
		ConversationId: req.ConversationId,
		PageSize:       int32(req.PageSize),
		PageNo:         int32(req.PageNo),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return response.PageResult{}, err
	}

	// 转换resp.Data为自定义的ConversionDetailInfo结构体数组
	var convertedList []response.ConversationDetailInfo
	for _, item := range resp.Data {
		convertedItem := response.ConversationDetailInfo{
			Id:                  item.Id,
			AssistantId:         item.AssistantId,
			ConversationId:      item.ConversationId,
			Prompt:              item.Prompt,
			SysPrompt:           item.SysPrompt,
			Response:            item.Response,
			ResponseList:        buildResponseList(item.ConversationResponse),
			QaType:              item.QaType,
			CreatedBy:           item.CreatedBy,
			CreatedAt:           item.CreatedAt,
			UpdatedAt:           item.UpdatedAt,
			RequestFiles:        transRequestFiles(item.RequestFiles),
			FileSize:            item.FileSize,
			FileName:            item.FileName,
			SubConversationList: buildSubConversationList(item.SubConversationList),
		}

		// 将SearchList从string转换为interface{}
		convertedItem.SearchList = buildSearchList(item.SearchList)

		convertedList = append(convertedList, convertedItem)

		// 对切片进行排序
		sort.Slice(convertedList, func(i, j int) bool {
			// CreatedAt值小的时间更早，排在前面
			return convertedList[i].CreatedAt < convertedList[j].CreatedAt
		})
	}

	return response.PageResult{Total: resp.Total, List: convertedList, PageNo: req.PageNo, PageSize: req.PageSize}, nil
}

func buildResponseList(conversationResponse []*assistant_service.ConversationResponse) []*response.ConversationResponse {
	if len(conversationResponse) == 0 {
		return make([]*response.ConversationResponse, 0)
	}
	var retList []*response.ConversationResponse
	for _, resp := range conversationResponse {
		retList = append(retList, &response.ConversationResponse{
			Response: resp.Response,
			Order:    resp.Order,
		})
	}
	return retList
}

func buildSearchList(searchListStr string) interface{} {
	// 将SearchList从string转换为interface{}
	if searchListStr != "" {
		var searchList interface{}
		if err := json.Unmarshal([]byte(searchListStr), &searchList); err != nil {
			log.Warnf("解析SearchList失败，使用原始字符串，error: %v, searchList: %s", err, searchListStr)
			return searchList
		} else {
			return searchListStr
		}
	}
	return nil
}

func buildSubConversationList(conversationList []*assistant_service.SubConversation) []*response.SubConversation {
	if len(conversationList) == 0 {
		return make([]*response.SubConversation, 0)
	}
	var subConversationList []*response.SubConversation
	for _, conversation := range conversationList {
		subConversationList = append(subConversationList, &response.SubConversation{
			Response:         conversation.Response,
			SearchList:       buildSearchList(conversation.SearchList),
			Id:               conversation.Id,
			Name:             conversation.Name,
			Profile:          conversation.Profile,
			TimeCost:         conversation.TimeCost,
			Status:           conversation.Status,
			ConversationType: conversation.ConversationType,
			Order:            conversation.Order,
		})
	}
	return subConversationList
}

func transKnowledgebases2Proto(kbConfig request.AppKnowledgebaseConfig) *assistant_service.AssistantKnowledgeBaseConfig {
	var knowIds []string
	if len(kbConfig.Knowledgebases) > 0 {
		for _, v := range kbConfig.Knowledgebases {
			knowIds = append(knowIds, v.ID)
		}
	}
	return &assistant_service.AssistantKnowledgeBaseConfig{
		KnowledgeBaseIds:     knowIds,
		MaxHistory:           kbConfig.Config.MaxHistory,
		Threshold:            kbConfig.Config.Threshold,
		TopK:                 kbConfig.Config.TopK,
		MatchType:            kbConfig.Config.MatchType,
		KeywordPriority:      kbConfig.Config.KeywordPriority,
		PriorityMatch:        kbConfig.Config.PriorityMatch,
		SemanticsPriority:    kbConfig.Config.SemanticsPriority,
		TermWeight:           kbConfig.Config.TermWeight,
		TermWeightEnable:     kbConfig.Config.TermWeightEnable,
		UseGraph:             kbConfig.Config.UseGraph,
		AppKnowledgeBaseList: transKnowledgeParams(kbConfig.Knowledgebases),
	}
}

func transKnowledgeParams(paramsList []request.AppKnowledgeBase) []*assistant_service.AppKnowledgeBase {
	if len(paramsList) == 0 {
		return nil
	}
	var retList []*assistant_service.AppKnowledgeBase
	for _, base := range paramsList {
		retList = append(retList, &assistant_service.AppKnowledgeBase{
			KnowledgeBaseId:      base.ID,
			KnowledgeBaseName:    base.Name,
			GraphSwitch:          base.GraphSwitch,
			Category:             base.Category,
			MetaDataFilterParams: transKnowledgeMetaParams(base.MetaDataFilterParams),
		})
	}
	return retList
}

func transKnowledgeMetaParams(baseInfo *request.MetaDataFilterParams) *assistant_service.MetaDataFilterParams {
	if baseInfo == nil {
		return nil
	}
	return &assistant_service.MetaDataFilterParams{
		FilterEnable:     baseInfo.FilterEnable,
		FilterLogicType:  baseInfo.FilterLogicType,
		MetaFilterParams: transMetaFilterParams(baseInfo.MetaFilterParams),
	}
}

func transMetaFilterParams(metaFilterList []*request.MetaFilterParams) []*assistant_service.MetaFilterParams {
	if metaFilterList == nil {
		return nil
	}
	var metaList []*assistant_service.MetaFilterParams
	for _, m := range metaFilterList {
		metaList = append(metaList, &assistant_service.MetaFilterParams{
			Condition: m.Condition,
			Key:       m.Key,
			Type:      m.Type,
			Value:     m.Value,
		})
	}
	return metaList
}

func transSafetyConfig2Proto(tables []request.SensitiveTable) []*assistant_service.SensitiveTable {
	if tables == nil {
		return nil
	}
	result := make([]*assistant_service.SensitiveTable, 0, len(tables))
	for _, table := range tables {
		result = append(result, &assistant_service.SensitiveTable{
			TableId:   table.TableId,
			TableName: table.TableName,
		})
	}
	return result
}

func transAssistantResp2Model(ctx *gin.Context, resp *assistant_service.AssistantInfo) (*response.Assistant, error) {
	log.Debugf("开始转换Assistant响应到模型，响应内容: %+v", resp)
	if resp == nil {
		log.Debugf("Assistant响应为空，返回空Assistant模型")
		return nil, nil
	}

	// 获取app发布信息，可能没有发布过，不返回错误
	appInfo, _ := app.GetAppInfo(ctx, &app_service.GetAppInfoReq{AppId: resp.AssistantId, AppType: constant.AppTypeAgent})

	// 转换Model配置
	modelConfig, err := assistantModelConvert(ctx, resp.ModelConfig)
	if err != nil {
		return nil, err
	}

	// 转换Workflow配置
	assistantWorkFlowInfos, err := assistantWorkFlowConvert(ctx, resp.WorkFlowInfos)
	if err != nil {
		return nil, err
	}

	// 查询该用户所有权限的所有 MCP
	assistantMCPInfos, err := assistantMCPConvert(ctx, resp.McpInfos)
	if err != nil {
		return nil, err
	}

	// 查询该用户所有权限的 custom、builtin 工具
	assistantToolInfos, err := assistantToolsConvert(ctx, resp.ToolInfos)
	if err != nil {
		return nil, err
	}

	// 转换Skill配置
	assistantSkillInfos, err := assistantSkillConvert(ctx, resp.SkillInfos)
	if err != nil {
		return nil, err
	}

	// 转换Safety配置
	safetyConfig, err := assistantSafetyConvert(ctx, resp.SafetyConfig)
	if err != nil {
		return nil, err
	}

	// 转换KnowledgeBase配置
	knowledgeBaseConfig, err := transKnowledgeBases2Model(ctx, resp.KnowledgeBaseConfig)
	if err != nil {
		return nil, err
	}

	// 转换Rerank配置
	rerankConfig := request.AppModelConfig{}
	if len(knowledgeBaseConfig.Knowledgebases) > 0 {
		rerankConfig, err = assistantRerankConvert(ctx, resp.RerankConfig)
		if err != nil {
			return nil, err
		}
	}

	// 转换Vision配置
	var visionConfig response.VisionConfig
	if resp.VisionConfig != nil {
		visionConfig = response.VisionConfig{
			MaxPicNum: resp.VisionConfig.MaxPicNum,
			PicNum:    resp.VisionConfig.PicNum,
		}
	}

	// 转换Memory配置
	memoryConfig := request.MemoryConfig{
		MaxHistoryLength: resp.MemoryConfig.MaxHistoryLength,
	}

	// 转换Recommend配置
	recommendConfig, err := assistantRecommendConvert(ctx, resp)
	if err != nil {
		return nil, err
	}

	// 转换MultiAgent配置
	assistantMultiAgents := assistantMultiAgentConvert(ctx, resp)

	assistantModel := response.Assistant{
		AssistantId:         resp.AssistantId,
		UUID:                resp.Uuid,
		AppBriefConfig:      appBriefConfigProto2Model(ctx, resp.AssistantBrief, constant.AppTypeAgent),
		Prologue:            resp.Prologue,
		Instructions:        resp.Instructions,
		RecommendQuestion:   resp.RecommendQuestion,
		KnowledgeBaseConfig: knowledgeBaseConfig,
		ModelConfig:         modelConfig,
		RerankConfig:        rerankConfig,
		SafetyConfig:        safetyConfig,
		VisionConfig:        visionConfig,
		MemoryConfig:        memoryConfig,
		RecommendConfig:     recommendConfig,
		Scope:               resp.Scope,
		WorkFlowInfos:       assistantWorkFlowInfos,
		MCPInfos:            assistantMCPInfos,
		ToolInfos:           assistantToolInfos,
		SkillInfos:          assistantSkillInfos,
		MultiAgentInfos:     assistantMultiAgents,
		CreatedAt:           util.Time2Str(resp.CreatTime),
		UpdatedAt:           util.Time2Str(resp.UpdateTime),
		NewAgent:            true,
		PublishType:         appInfo.GetPublishType(),
		Category:            resp.Category,
	}

	log.Debugf("Assistant响应到模型转换完成，结果: %+v", assistantModel)
	return &assistantModel, nil
}

func assistantMultiAgentConvert(ctx *gin.Context, resp *assistant_service.AssistantInfo) []*response.AssistantAgentInfo {
	assistantMultiAgents := make([]*response.AssistantAgentInfo, 0)
	for _, agent := range resp.MultiAgentInfos {
		multiAgent := &response.AssistantAgentInfo{
			AgentId: agent.AgentId,
			Name:    agent.Name,
			Desc:    agent.Desc,
			Avatar:  cacheAppAvatar(ctx, agent.AvatarPath, constant.AppTypeAgent),
			Enable:  agent.Enable,
		}
		assistantMultiAgents = append(assistantMultiAgents, multiAgent)
	}
	return assistantMultiAgents

}

func assistantRecommendConvert(ctx *gin.Context, resp *assistant_service.AssistantInfo) (recommendConfig response.RecommendConfig, err error) {
	if resp.RecommendConfig != nil {
		modelConfig, err := assistantModelConvert(ctx, resp.RecommendConfig.ModelConfig)
		if err != nil {
			recommendConfig.ModelConfig = modelConfig
			return recommendConfig, err
		}
		recommendConfig = response.RecommendConfig{
			ModelConfig:     modelConfig,
			MaxHistory:      resp.RecommendConfig.MaxHistory,
			RecommendEnable: resp.RecommendConfig.RecommendEnable,
			Prompt:          resp.RecommendConfig.SystemPrompt,
			PromptEnable:    resp.RecommendConfig.PromptEnable,
		}
	}
	return recommendConfig, nil
}

func transKnowledgeBases2Model(ctx *gin.Context, kbConfig *assistant_service.AssistantKnowledgeBaseConfig) (request.AppKnowledgebaseConfig, error) {
	if kbConfig == nil {
		log.Debugf("知识库配置为空")
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}, nil
	}
	if len(kbConfig.KnowledgeBaseIds) == 0 {
		log.Debugf("知识库配置为空")
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}, nil
	}

	// 获取知识库详情列表
	kbInfoList, err := knowledgeBase.SelectKnowledgeDetailByIdList(ctx, &knowledgeBase_service.KnowledgeDetailSelectListReq{
		KnowledgeIds: kbConfig.KnowledgeBaseIds,
	})

	if err != nil || kbInfoList == nil || len(kbInfoList.List) == 0 {
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}, err
	}

	knowledgeBases := buildKnowledgeBases(ctx, kbInfoList, kbConfig.KnowledgeBaseIds, kbConfig.AppKnowledgeBaseList)

	return request.AppKnowledgebaseConfig{
		Knowledgebases: knowledgeBases,
		Config: request.AppKnowledgebaseParams{
			MaxHistory:        kbConfig.MaxHistory,
			Threshold:         kbConfig.Threshold,
			TopK:              kbConfig.TopK,
			MatchType:         kbConfig.MatchType,
			PriorityMatch:     kbConfig.PriorityMatch,
			SemanticsPriority: kbConfig.SemanticsPriority,
			KeywordPriority:   kbConfig.KeywordPriority,
			TermWeight:        kbConfig.TermWeight,
			TermWeightEnable:  kbConfig.TermWeightEnable,
			UseGraph:          kbConfig.UseGraph,
		},
	}, nil

}

func buildKnowledgeBases(ctx *gin.Context, kbInfoList *knowledgeBase_service.KnowledgeDetailSelectListResp, kbIdList []string, kbConfigList []*assistant_service.AppKnowledgeBase) []request.AppKnowledgeBase {
	if len(kbInfoList.List) == 0 {
		return make([]request.AppKnowledgeBase, 0)
	}
	var knowledgeMap = make(map[string]*knowledgeBase_service.KnowledgeInfo)
	for _, kbInfo := range kbInfoList.List {
		knowledgeMap[kbInfo.KnowledgeId] = kbInfo
	}
	var knowledgeBases = make([]request.AppKnowledgeBase, 0)
	if len(kbConfigList) > 0 {
		for _, kbConfig := range kbConfigList {
			info := knowledgeMap[kbConfig.KnowledgeBaseId]
			if info == nil {
				continue
			}
			share := info.ShareCount > 1
			var orgName string
			if share {
				orgInfo, err := iam.GetOrgInfo(ctx, &iam_service.GetOrgInfoReq{OrgId: info.CreateOrgId})
				if err != nil {
					log.Errorf("get org info error: %v", err)
				} else {
					orgName = buildShareOrgName(share, orgInfo.Name)
				}
			}
			params := buildAssistantMetaDataFilterParams(kbConfig)
			knowledgeBases = append(knowledgeBases, request.AppKnowledgeBase{
				ID:                   kbConfig.KnowledgeBaseId,
				Name:                 info.Name,
				GraphSwitch:          info.GraphSwitch,
				External:             info.External,
				Category:             info.Category,
				Share:                share,
				OrgName:              orgName,
				MetaDataFilterParams: params,
				Avatar:               cacheKnowledgeAvatar(ctx, info.AvatarPath, info.Category),
			})
		}
	} else {
		for _, kbId := range kbIdList {
			info := knowledgeMap[kbId]
			if info == nil {
				continue
			}
			knowledgeBases = append(knowledgeBases, request.AppKnowledgeBase{
				ID:   kbId,
				Name: info.Name,
			})
		}
	}

	return knowledgeBases
}

func buildAssistantMetaDataFilterParams(kbConfig *assistant_service.AppKnowledgeBase) *request.MetaDataFilterParams {
	params := kbConfig.MetaDataFilterParams
	if params == nil {
		return nil
	}
	return &request.MetaDataFilterParams{
		FilterEnable:     params.FilterEnable,
		FilterLogicType:  params.FilterLogicType,
		MetaFilterParams: buildAssistantMetaFilterParams(params.MetaFilterParams),
	}
}

func buildAssistantMetaFilterParams(metaFilterList []*assistant_service.MetaFilterParams) []*request.MetaFilterParams {
	if metaFilterList == nil {
		return nil
	}
	var metaList []*request.MetaFilterParams
	for _, m := range metaFilterList {
		metaList = append(metaList, &request.MetaFilterParams{
			Condition: m.Condition,
			Key:       m.Key,
			Type:      m.Type,
			Value:     m.Value,
		})
	}
	return metaList
}

func transRequestFiles(files []*assistant_service.RequestFile) []response.AssistantRequestFile {
	if files == nil {
		return nil
	}
	var result []response.AssistantRequestFile
	for _, file := range files {
		result = append(result, response.AssistantRequestFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  file.FileUrl,
		})
	}
	return result
}

func assistantBriefConfigModel2Proto(appBrief request.AssistantCreateReq) *common.AppBriefConfig {
	return &common.AppBriefConfig{
		Name:       appBrief.Name,
		Desc:       appBrief.Desc,
		AvatarPath: appBrief.Avatar.Key,
	}

}

func recommendConfigModel2Proto(recommendConfig request.RecommendConfig) (ret *assistant_service.AssistantRecommendConfig, err error) {
	modelConfig := &common.AppModelConfig{}
	if recommendConfig.ModelConfig.ModelId != "" {
		modelConfig, err = appModelConfigModel2Proto(recommendConfig.ModelConfig)
		if err != nil {
			return nil, err
		}
	}
	ret = &assistant_service.AssistantRecommendConfig{
		RecommendEnable: recommendConfig.RecommendEnable,
		ModelConfig:     modelConfig,
		SystemPrompt:    strings.TrimSpace(recommendConfig.Prompt),
		PromptEnable:    recommendConfig.PromptEnable,
		MaxHistory:      recommendConfig.MaxHistory,
	}

	return ret, nil
}
