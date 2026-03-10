package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

const (
	skillConversationESIndexName = "conversation_detail_infos_skill_*"
	skillConversationAuthor      = "wanwu"
)

func getSkillConversationESIndexName() string {
	now := time.Now()
	indexName := fmt.Sprintf("conversation_detail_infos_skill_%d%02d%02d", now.Year(), now.Month(), now.Day())
	return indexName
}

func CreateSkillConversation(ctx *gin.Context, userId, orgId string, req request.CreateSkillConversationReq) (*response.CreateSkillConversationResp, error) {
	rpcResp, err := assistant.CreateSkillConversation(ctx.Request.Context(), &assistant_service.CreateSkillConversationReq{
		Title: req.Title,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateSkillConversationResp{
		ConversationId: rpcResp.ConversationId,
	}, nil
}

func DeleteSkillConversation(ctx *gin.Context, userId, orgId, conversationId string) error {
	// 异步删除 ES 中的历史记录
	go func() {
		defer util.PrintPanicStack()
		_, _ = assistant.DeleteFromES(ctx.Request.Context(), &assistant_service.DeleteFromESReq{
			IndexName: skillConversationESIndexName,
			Conditions: map[string]string{
				"conversationId": conversationId,
			},
		})
	}()

	_, err := assistant.DeleteSkillConversation(ctx.Request.Context(), &assistant_service.DeleteSkillConversationReq{
		ConversationId: conversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func GetSkillConversationList(ctx *gin.Context, userId, orgId string, req request.GetSkillConversationListReq) (*response.PageResult, error) {
	rpcResp, err := assistant.GetSkillConversationList(ctx.Request.Context(), &assistant_service.GetSkillConversationListReq{
		PageNo:   int32(req.PageNo),
		PageSize: int32(req.PageSize),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	list := make([]response.SkillConversationItem, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, response.SkillConversationItem{
			ConversationId: item.ConversationId,
			Title:          item.Title,
			CreatedAt:      util.Time2Str(item.CreatedAt),
		})
	}

	return &response.PageResult{
		List:     list,
		Total:    rpcResp.Total,
		PageNo:   req.PageNo,
		PageSize: req.PageSize,
	}, nil
}

func GetSkillConversationDetail(ctx *gin.Context, userId, orgId string, req request.GetSkillConversationDetailReq) (*response.ListResult, error) {

	// 查询对话记录
	detailList, err := getSkillConversationDetailListFromES(ctx, req.ConversationId)
	if err != nil {
		return nil, err
	}

	// 提取 SaveId
	saveIds := make([]string, 0, len(detailList))
	for _, detail := range detailList {
		for _, rf := range detail.ResponseFiles {
			if skillSaveId, ok := rf.MetaData["skillSaveId"].(string); ok {
				saveIds = append(saveIds, skillSaveId)
			}
		}
	}

	// 是否已发送
	if len(saveIds) > 0 {
		mcpResp, err := mcp.CustomSkillGetBySaveIds(ctx.Request.Context(), &mcp_service.CustomSkillGetBySaveIdsReq{
			SaveIds: saveIds,
		})
		if err != nil {
			return nil, err
		}
		// 有效的 saveIds 集合
		validSaveIds := make(map[string]bool, len(mcpResp.SaveIds))
		for _, saveId := range mcpResp.SaveIds {
			validSaveIds[saveId] = true
		}
		// 标记是否已发送
		for _, detail := range detailList {
			for i := range detail.ResponseFiles {
				if skillSaveId, ok := detail.ResponseFiles[i].MetaData["skillSaveId"].(string); ok {
					detail.ResponseFiles[i].MetaData["inResource"] = validSaveIds[skillSaveId]
				}
			}
		}
	}

	return &response.ListResult{
		List:  detailList,
		Total: int64(len(detailList)),
	}, nil
}

func SkillConversationChat(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq) error {

	if req.ModelConfig == nil || req.ModelConfig.ModelId == "" {
		return fmt.Errorf("modelConfig or modelId is empty")
	}

	// 查询模型信息
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{
		ModelId: req.ModelConfig.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return err
	}
	modelConfig := wga_sandbox_option.ModelConfig{
		Provider:     modelInfo.Provider,
		ProviderName: modelInfo.Provider,
		Model:        modelInfo.Model,
		ModelName:    modelInfo.DisplayName,
	}
	endpoint := mp.ToModelEndpoint(modelInfo.ModelId, modelInfo.Model)
	for k, v := range endpoint {
		if k == "model_url" {
			modelConfig.BaseURL = v.(string)
			break
		}
	}

	// 查询对话记录
	detailList, err := getSkillConversationDetailListFromES(ctx, req.ConversationId)
	if err != nil {
		return err
	}
	messages := make([]adk.Message, 0, len(detailList)*2+1)
	for _, detail := range detailList {
		messages = append(messages, &schema.Message{
			Role:    schema.User,
			Content: detail.Prompt,
		}, &schema.Message{
			Role:    schema.Assistant,
			Content: detail.Response,
		})
	}
	// 当前任务
	messages = append(messages, &schema.Message{
		Role:    schema.User,
		Content: req.Query,
	})

	// 存储路径 /tmp/skills/<uuid>
	messageId := util.GenUUID()
	workspaceDir := filepath.Join("/tmp/skills", messageId)

	if len(req.FileInfo) > 0 {
		if err := os.MkdirAll(workspaceDir, 0755); err != nil {
			return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("create workspace directory err: %v", err))
		}
		for _, fileInfo := range req.FileInfo {
			localPath := filepath.Join(workspaceDir, fileInfo.FileName)
			data, err := minio_util.DownloadFileDirect(ctx.Request.Context(), fileInfo.FileUrl)
			if err != nil {
				return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download file %s err: %v", fileInfo.FileName, err))
			}
			if err := os.WriteFile(localPath, data, 0644); err != nil {
				return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("save file %s into workspace err: %v", fileInfo.FileName, err))
			}
		}
	}

	// 流式问答
	streamCh, err := RunSkillCreator(ctx, modelConfig, messageId, workspaceDir, workspaceDir, messages)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}

	// 处理流式问答
	var responseStr string
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Skill] conversation %v user %v org %v recv", req.ConversationId, userId, orgId), sse_util.DONE_EMPTY).
		WriteStream(streamCh, nil, buildSkillChatRespLineProcessor(&responseStr), buildSkillChatDoneProcessor(ctx, userId, orgId, req, messageId, workspaceDir, &responseStr))
	return nil
}

func SkillConversationSave(ctx *gin.Context, userId, orgId string, req request.SkillConversationSaveReq) (*response.CustomSkillIDResp, error) {

	// 查询对话记录
	detailList, err := getSkillConversationDetailListFromES(ctx, req.ConversationId)
	if err != nil {
		return nil, err
	}
	// 查找zipUrl
	var zipUrl string
	for _, detail := range detailList {
		for _, rf := range detail.ResponseFiles {
			if skillSaveId, ok := rf.MetaData["skillSaveId"].(string); ok && skillSaveId == req.SkillSaveId {
				zipUrl = rf.FileUrl
				break
			}
		}
		if zipUrl != "" {
			break
		}
	}

	if zipUrl == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "skillSaveId not found in conversation files")
	}

	// 保存至资源库自定义Skills
	return CreateCustomSkill(ctx, userId, orgId, request.CreateCustomSkillReq{
		Author:     skillConversationAuthor,
		ZipUrl:     zipUrl,
		SaveId:     req.SkillSaveId,
		SourceType: "skill_conversation",
	})
}

// --- internal ---

func getSkillConversationDetailListFromES(ctx *gin.Context, conversationId string) ([]*response.SkillConversationDetailInfo, error) {
	searchResp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
		IndexName: skillConversationESIndexName,
		Conditions: map[string]string{
			"conversationId": conversationId,
		},
		PageNo:    1,
		PageSize:  1000,
		SortField: "createdAt",
		SortOrder: "asc",
	})
	if err != nil {
		return nil, err
	}
	respList := make([]*response.SkillConversationDetailInfo, 0, len(searchResp.DocJsonList))
	for _, docJSON := range searchResp.DocJsonList {
		var item response.SkillConversationDetailInfo
		if err := json.Unmarshal([]byte(docJSON), &item); err != nil {
			log.Errorf("[Skill] conversation %v parse ES doc json err: %v", conversationId, err)
			continue
		}
		respList = append(respList, &item)
	}
	return respList, nil
}

func buildSkillChatDoneProcessor(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq, messageId, outputDir string, responseStr *string) func(sse_util.SSEWriterClient[string], interface{}) error {
	return func(c sse_util.SSEWriterClient[string], params interface{}) error {

		lastSSE := response.SkillConversationSSEData{
			ConversationSSEData: response.ConversationSSEData{
				Message: "success",
				Finish:  1,
			},
		}

		defer func() {
			// save to es
			createdAt := time.Now().UnixMilli()
			var requestFiles []response.AssistantRequestFile
			for _, fileInfo := range req.FileInfo {
				requestFiles = append(requestFiles, response.AssistantRequestFile{
					FileName: fileInfo.FileName,
					FileSize: fileInfo.FileSize,
					FileUrl:  fileInfo.FileUrl,
				})
			}
			b, _ := json.Marshal(&response.SkillConversationDetailInfo{
				ConversationDetailInfo: response.ConversationDetailInfo{
					Id:             messageId,
					ConversationId: req.ConversationId,
					Prompt:         req.Query,
					Response:       *responseStr,
					CreatedBy:      userId,
					CreatedAt:      createdAt,
					UpdatedAt:      createdAt,
					RequestFiles:   requestFiles,
				},
				ResponseFiles: lastSSE.ResponseFiles,
			})
			if _, err := assistant.SaveToES(ctx.Request.Context(), &assistant_service.SaveToESReq{
				IndexName: getSkillConversationESIndexName(),
				DocJson:   string(b),
			}); err != nil {
				log.Errorf("[Skill] conversation %v user %v org %v save to es err: %v", req.ConversationId, userId, orgId, err)
			}

			// done sse
			marshal, _ := json.Marshal(lastSSE)
			data := "data: " + string(marshal) + "\n\n"
			_ = c.Write(data)
			_ = c.Write(sse_util.DONE_MSG)
			c.Flush()
		}()

		// 目录是否为空
		entries, err := os.ReadDir(outputDir)
		if err != nil || len(entries) == 0 {
			return nil
		}

		// 压缩文件夹
		zipBytes, err := util.ZipDir(outputDir + "/.")
		if err != nil {
			return err
		}
		// skillName, skillDesc
		_, skillName, skillDesc, err := extractSkillMarkdownFromZip(zipBytes)
		if err != nil {
			return err
		}
		// 上传到 minio
		fileName, _, err := minio.UploadFileCommon(ctx.Request.Context(), bytes.NewReader(zipBytes), ".zip", -1, false)
		if err != nil {
			return err
		}
		// 构建文件访问路径并添加到响应中
		lastSSE.ResponseFiles = append(lastSSE.ResponseFiles, &response.AssistantResponseFile{
			FileName: fileName,
			FileSize: int64(len(zipBytes)),
			FileUrl:  buildAccessFilePath(filepath.Join(minio.BucketFileUpload, minio.DirFileExpire, fileName)),
			MIMEType: "application/zip",
			MetaData: map[string]interface{}{
				"name":        skillName,
				"desc":        skillDesc,
				"author":      skillConversationAuthor,
				"avatar":      cacheSkillAvatar(ctx, ""),
				"inResource":  false,
				"expiredAt":   util.Time2Str(time.Now().AddDate(0, 0, 7).UnixMilli()), // 7天后过期
				"skillSaveId": messageId,
			},
		})
		// 删除临时文件
		if err := util.DeleteDir(outputDir); err != nil {
			return err
		}
		return nil
	}
}

func buildSkillChatRespLineProcessor(responeStr *string) func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {

		// 累计流式输出文本
		*responeStr += lineText

		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		resp := response.SkillConversationSSEData{
			ConversationSSEData: response.ConversationSSEData{
				Response: lineText,
			},
		}
		marshal, _ := json.Marshal(resp)
		return "data: " + string(marshal) + "\n\n", false, nil
	}
}
