package service

import (
	"strings"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
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
	//rpcResp, err := mcp.CreateSkillConversation(ctx.Request.Context(), &mcp_service.CreateSkillConversationReq{
	//	Title: req.Title,
	//	Identity: &mcp_service.Identity{
	//		UserId: userId,
	//		OrgId:  orgId,
	//	},
	//})
	//if err != nil {
	//	return nil, err
	//}
	//return &response.CreateSkillConversationResp{
	//	ConversationId: rpcResp.ConversationId,
	//}, nil
	return nil, nil
}

func DeleteSkillConversation(ctx *gin.Context, userId, orgId, conversationId string) error {
	//_, err := mcp.DeleteSkillConversation(ctx.Request.Context(), &mcp_service.DeleteSkillConversationReq{
	//	ConversationId: conversationId,
	//	Identity: &mcp_service.Identity{
	//		UserId: userId,
	//		OrgId:  orgId,
	//	},
	//})
	return nil
}

func GetSkillConversationList(ctx *gin.Context, userId, orgId string, req request.GetSkillConversationListReq) (*response.PageResult, error) {
	//rpcResp, err := mcp.GetSkillConversationList(ctx.Request.Context(), &mcp_service.GetSkillConversationListReq{
	//	PageNo:   int32(req.PageNo),
	//	PageSize: int32(req.PageSize),
	//	Identity: &mcp_service.Identity{
	//		UserId: userId,
	//		OrgId:  orgId,
	//	},
	//})
	//if err != nil {
	//	return nil, err
	//}
	//
	//list := make([]response.SkillConversationItem, 0, len(rpcResp.List))
	//for _, item := range rpcResp.List {
	//	list = append(list, response.SkillConversationItem{
	//		ConversationId: item.ConversationId,
	//		Title:          item.Title,
	//		CreatedAt:      item.CreatedAt,
	//	})
	//}
	//
	//return &response.PageResult{
	//	List:     list,
	//	Total:    rpcResp.Total,
	//	PageNo:   req.PageNo,
	//	PageSize: req.PageSize,
	//}, nil
	return nil, nil
}

func GetSkillConversationDetail(ctx *gin.Context, userId, orgId string, req request.GetSkillConversationDetailReq) (*response.ListResult, error) {
	//rpcResp, err := mcp.GetSkillConversationDetail(ctx.Request.Context(), &mcp_service.GetSkillConversationDetailReq{
	//	ConversationId: req.ConversationId,
	//	Identity: &mcp_service.Identity{
	//		UserId: userId,
	//		OrgId:  orgId,
	//	},
	//})
	//if err != nil {
	//	return nil, err
	//}
	//
	//list := make([]response.SkillConversationDetailInfo, 0, len(rpcResp.List))
	//for _, item := range rpcResp.List {
	//	if item == nil {
	//		continue
	//	}
	//	// Map ResponseList
	//	var responseList []*response.SkillConversationDetailInfo
	//	responseList = append(responseList, &response.SkillConversationDetailInfo{
	//		ConversationDetailInfo: response.ConversationDetailInfo{
	//			Id:        item.ConversationId,
	//			Prompt:    item.Prompt,
	//			Response:  item.Response,
	//			CreatedAt: item.CreatedAt,
	//		},
	//	})
	//
	//	// Map RequestFiles
	//	var requestFiles []response.AssistantRequestFile
	//	if item.RequestFiles != nil {
	//		for _, f := range item.RequestFiles {
	//			requestFiles = append(requestFiles, response.AssistantRequestFile{
	//				FileName: f.Name,
	//				FileSize: f.Size,
	//				FileUrl:  f.Url,
	//			})
	//		}
	//	}
	//
	//	respItem := response.SkillConversationDetailInfo{
	//		ConversationDetailInfo: response.ConversationDetailInfo{
	//			Id:        item.ConversationId,
	//			Prompt:    item.Prompt,
	//			Response:  item.Response,
	//			CreatedAt: item.CreatedAt,
	//		},
	//	}
	//	list = append(list, respItem)
	//}
	//
	//return &response.ListResult{
	//	List:  list,
	//	Total: rpcResp.Total,
	//}, nil
	return nil, nil
}

func SkillConversationChat(ctx *gin.Context, userId, orgId string, req request.SkillConversationChatReq) (chan string, error) {
	//rpcReqFiles := make([]*mcp_service.SkillRequestFile, 0, len(req.FileInfo))
	//for _, f := range req.FileInfo {
	//	rpcReqFiles = append(rpcReqFiles, &mcp_service.SkillRequestFile{
	//		Name: f.FileName,
	//		Size: f.FileSize,
	//		Url:  f.FileUrl,
	//	})
	//}
	//
	//rpcReq := &mcp_service.SkillConversationChatReq{
	//	ConversationId: req.ConversationId,
	//	Query:          req.Query,
	//	Files:          rpcReqFiles,
	//	Identity: &mcp_service.Identity{
	//		UserId: userId,
	//		OrgId:  orgId,
	//	},
	//}
	//
	//stream, err := mcp.SkillConversationChat(ctx.Request.Context(), rpcReq)
	//if err != nil {
	//	return nil, err
	//}
	//
	//msgChan := make(chan string)
	//
	//go func() {
	//	defer close(msgChan)
	//	for {
	//		resp, err := stream.Recv()
	//		if err == io.EOF {
	//			return
	//		}
	//		if err != nil {
	//			// Log error?
	//			return
	//		}
	//		msgChan <- resp.Content
	//	}
	//}()

	return nil, nil
}

func SkillConversationSave(ctx *gin.Context, userId, orgId string, req request.SkillConversationSaveReq) error {
	//_, err := mcp.SkillConversationSave(ctx.Request.Context(), &mcp_service.SkillConversationSaveReq{
	//	ConversationId: req.ConversationId,
	//	SkillSaveId:    req.SkillSaveId,
	//	Identity: &mcp_service.Identity{
	//		UserId: userId,
	//		OrgId:  orgId,
	//	},
	//})
	return nil
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
