package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
)

func GetAgentSkillList(ctx *gin.Context, name string) (*response.ListResult, error) {
	var skillsTemplateList []*response.SkillDetail
	for _, skillsCfg := range config.Cfg().AgentSkills {
		if name != "" && !strings.Contains(skillsCfg.Name, name) {
			continue
		}
		skillsTemplateList = append(skillsTemplateList, buildSkillTempDetail(*skillsCfg, false))
	}
	return &response.ListResult{
		List:  skillsTemplateList,
		Total: int64(len(skillsTemplateList)),
	}, nil
}

func GetAgentSkillDetail(ctx *gin.Context, skillId string) (*response.SkillDetail, error) {
	skillsCfg, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_agent_skill_detail", "get skill detail empty")
	}
	return buildSkillTempDetail(skillsCfg, true), nil
}

func DownloadAgentSkill(ctx *gin.Context, skillId string) ([]byte, error) {
	// 需要把SkConfigDir+templateId路径下的所有文件在内存打成一个压缩包
	sf, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_agent_skill_download", "get skill detail empty")
	}
	return sf.AgentSkillZipToBytes(skillId)
}

// --- Skill Conversation ---

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
		// 索引格式为 skill_creation_conversation_detail_*
		_, _ = assistant.DeleteFromES(ctx.Request.Context(), &assistant_service.DeleteFromESReq{
			IndexName: "skill_creation_conversation_detail_*",
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
	if req.ConversationId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_skill_conversation_detail", "conversationId is empty")
	}

	// 从 ES 中读取
	indexName := "skill_creation_conversation_detail_*"
	searchResp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
		IndexName: indexName,
		Conditions: map[string]string{
			"conversationId": req.ConversationId,
		},
		PageNo:    1,
		PageSize:  1000,
		SortField: "createdAt",
		SortOrder: "desc",
	})
	if err != nil {
		return nil, err
	}

	// 提取 SaveId
	saveIds := make([]string, 0, len(searchResp.DocJsonList))

	// 会话记录
	respList := make([]*response.SkillConversationDetailInfo, 0, len(searchResp.DocJsonList))
	for _, docJSON := range searchResp.DocJsonList {
		var item response.SkillConversationDetailInfo
		if err := json.Unmarshal([]byte(docJSON), &item); err != nil {
			continue
		}
		respList = append(respList, &item)
		// 提取 SaveId
		for _, rf := range item.ResponseFiles {
			if skillSaveId, ok := rf.MetaData["skillSaveId"].(string); ok {
				saveIds = append(saveIds, skillSaveId)
			}
		}
	}

	// 是否已发送
	mcpResp, err := mcp.CustomSkillGetBySaveIds(ctx.Request.Context(), &mcp_service.CustomSkillGetBySaveIdsReq{
		SaveIds: saveIds,
	})
	if err != nil {
		return nil, err
	}

	// 有效的 saveIds 集合
	validSaveIds := make(map[string]bool, len(mcpResp.SaveIds))
	for _, sid := range mcpResp.SaveIds {
		validSaveIds[sid] = true
	}

	// 标记是否已发送
	for _, item := range respList {
		for i := range item.ResponseFiles {
			if skillSaveId, ok := item.ResponseFiles[i].MetaData["skillSaveId"].(string); ok {
				item.ResponseFiles[i].MetaData["isSend"] = validSaveIds[skillSaveId]
			}
		}
	}

	return &response.ListResult{
		List:  respList,
		Total: searchResp.Total,
	}, nil
}

func SkillConversationChat(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq) error {
	// 查询模型信息
	_, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{
		ModelId: req.ModelConfig.ModelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return err
	}

	// 存储路径 /tmp/skills/<uuid>
	outputDir := fmt.Sprintf("/tmp/skills/%v", util.GenUUID())

	// 流式问答
	streamCh, err := skillConversationChatTemp(ctx, userId, orgId, req)
	if err != nil {
		return err
	}

	// 处理流式问答
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Skill] conversation %v user %v org %v recv", req.ConversationId, userId, orgId), sse_util.DONE_MSG).
		WriteStream(streamCh, nil, buildSkillChatRespLineProcessor(), buildSkillChatDoneProcessor(ctx, userId, orgId, req, outputDir))
	return nil
}

func skillConversationChatTemp(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq) (<-chan string, error) {
	url := "http://192.168.0.21:8081/service/api/openapi/v1/agent/chat"
	token := "ww-e1fa6296573c4a94a599596630528388"

	payload := map[string]interface{}{
		"uuid":            "2012090886208360448",
		"conversation_id": req.ConversationId,
		"query":           req.Query,
		"stream":          true,
	}
	payloadBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx.Request.Context(), "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawCh := make(chan string, 128)
	go func() {
		defer util.PrintPanicStack()
		defer close(rawCh)

		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				rawCh <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	return rawCh, nil
}

func buildSkillChatDoneProcessor(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq, OutputDir string) func(sse_util.SSEWriterClient[string], interface{}) error {
	return func(c sse_util.SSEWriterClient[string], params interface{}) error {
		skillSaveId := ""

		if OutputDir != "" {
			if _, err := os.Stat(OutputDir); err == nil {
				zipBytes, err := util.DirToBytes(OutputDir)
				if err != nil {
					log.Errorf("DirToBytes error: %v", err)
				} else {
					fileName := util.GenUUID() + ".zip"
					_, _, err := minio.UploadFileCommonWithNotExpire(ctx.Request.Context(), bytes.NewReader(zipBytes), ".zip", int64(len(zipBytes)))
					if err != nil {
						log.Errorf("UploadFileCommonWithNotExpire error: %v", err)
					} else {
						minioUrl, err := minio.GetUploadFileCommon(ctx.Request.Context(), fileName, false)
						if err != nil {
							log.Errorf("GetUploadFileCommon error: %v", err)
						} else {
							log.Infof("Uploaded skill zip to minio: %s", minioUrl)
							skillSaveId = "mock-skill-save-id-" + util.GenUUID()

							_, err = mcp.CustomSkillCreate(ctx.Request.Context(), &mcp_service.CustomSkillCreateReq{
								Name:       "skill-" + util.GenUUID(),
								ObjectPath: minioUrl,
								SourceType: "generated",
								Identity: &mcp_service.Identity{
									UserId: userId,
									OrgId:  orgId,
								},
							})
							if err != nil {
								log.Errorf("CustomSkillCreate error: %v", err)
							}
						}
					}
				}
			}
		}

		resp := response.SkillConversationChatResp{
			Code:    0,
			Message: "success",
			Finish:  1,
		}
		if skillSaveId != "" {
			resp.SkillSaveId = skillSaveId
		}

		marshal, _ := json.Marshal(resp)
		data := "data: " + string(marshal) + "\n\n"
		_ = c.Write(data)
		c.Flush()
		return nil
	}
}

func buildSkillChatRespLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		resp := response.SkillConversationChatResp{
			Response: lineText,
		}
		marshal, _ := json.Marshal(resp)
		return "data: " + string(marshal) + "\n\n", false, nil
	}
}

func SkillConversationSave(ctx *gin.Context, userId, orgId string, req request.SkillConversationSaveReq) (*response.CustomSkillIDResp, error) {
	//searchResp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
	//	IndexName: "skill_creation_conversation_detail_*",
	//	Conditions: map[string]string{
	//		"conversationId": req.ConversationId,
	//	},
	//	PageNo:    1,
	//	PageSize:  1000,
	//	SortField: "createdAt",
	//	SortOrder: "desc",
	//})
	//if err != nil {
	//	return err
	//}
	//
	//saveIds := make([]string, 0)
	//for _, docJSON := range searchResp.DocJsonList {
	//	var detail response.SkillConversationDetailInfo
	//	if err := json.Unmarshal([]byte(docJSON), &detail); err != nil {
	//		continue
	//	}
	//	for _, cardInfo := range detail.ResponseFiles {
	//		if cardInfo.MetaData["skillSaveId"] != "" {
	//			saveIds = append(saveIds, cardInfo.MetaData["skillSaveId"].(string))
	//		}
	//	}
	//}
	//
	//if len(saveIds) == 0 {
	//	return nil
	//}

	skillId, err := CreateCustomSkill(ctx, userId, orgId, request.CreateCustomSkillReq{
		Author:     "wanwu",
		ZipUrl:     "数据库获取",
		SaveId:     req.SkillSaveId,
		SourceType: "skill_conversation",
	})
	if err != nil {
		return nil, err
	}

	return skillId, nil
}

// --- internal ---
func buildSkillTempDetail(skillsCfg config.SkillsConfig, needMd bool) *response.SkillDetail {
	iconUrl := config.Cfg().DefaultIcon.SkillIcon
	if skillsCfg.Avatar != "" {
		iconUrl = skillsCfg.Avatar
	}
	ret := &response.SkillDetail{
		SkillId: skillsCfg.SkillId,
		Author:  skillsCfg.Author,
		Avatar:  request.Avatar{Path: iconUrl},
		Name:    skillsCfg.Name,
		Desc:    skillsCfg.Desc,
	}
	if needMd {
		ret.SkillMarkdown = string(skillsCfg.SkillMarkdown)
	}
	return ret
}
