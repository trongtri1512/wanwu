package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateSkillConversation
//
//	@Tags			resource.skill
//	@Summary		创建Skill生成会话
//	@Description	创建Skill生成会话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateSkillConversationReq	true	"创建会话参数"
//	@Success		200		{object}	response.Response{data=response.CreateSkillConversationResp}
//	@Router			/agent/skill/conversation [post]
func CreateSkillConversation(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateSkillConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateSkillConversation(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// DeleteSkillConversation
//
//	@Tags			resource.skill
//	@Summary		删除Skill生成会话
//	@Description	删除Skill生成会话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			conversationId	body		request.DeleteSkillConversationReq	true	"会话ID"
//	@Success		200				{object}	response.Response
//	@Router			/agent/skill/conversation [delete]
func DeleteSkillConversation(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteSkillConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteSkillConversation(ctx, userId, orgId, req.ConversationId)
	gin_util.Response(ctx, nil, err)
}

// ClearSkillConversation
//
//	@Tags			resource.skill
//	@Summary		清除Skill生成会话对话记录
//	@Description	清除Skill生成会话对话记录（保留会话ID）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			conversationId	body		request.ClearSkillConversationReq	true	"会话ID"
//	@Success		200				{object}	response.Response
//	@Router			/agent/skill/conversation/clear [delete]
func ClearSkillConversation(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ClearSkillConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ClearSkillConversation(ctx, userId, orgId, req.ConversationId)
	gin_util.Response(ctx, nil, err)
}

// GetSkillConversationList
//
//	@Tags			resource.skill
//	@Summary		获取Skill生成会话列表
//	@Description	获取Skill生成会话列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			pageNo		query		int	false	"页码"
//	@Param			pageSize	query		int	false	"每页数量"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.SkillConversationItem}}
//	@Router			/agent/skill/conversation/list [get]
func GetSkillConversationList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetSkillConversationListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetSkillConversationList(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// GetSkillConversationDetail
//
//	@Tags			resource.skill
//	@Summary		获取Skill生成会话详情
//	@Description	获取Skill生成会话详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			conversationId	query		string	true	"会话ID"
//	@Success		200				{object}	response.Response{data=response.ListResult{list=[]response.SkillConversationDetailInfo}}
//
//	@Router			/agent/skill/conversation/detail [get]
func GetSkillConversationDetail(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetSkillConversationDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetSkillConversationDetail(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// SkillConversationChat
//
//	@Tags			resource.skill
//	@Summary		Skill生成流式对话
//	@Description	Skill生成流式对话
//	@Security		JWT
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.SkillConversationChatReq	true	"对话参数"
//	@Success		200		{string}	string								"stream"
//	@Router			/agent/skill/conversation/chat [post]
func SkillConversationChat(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.SkillConversationChatReq
	if !gin_util.Bind(ctx, &req) {
		return
	}

	// service层直接处理流式响应写入
	err := service.SkillConversationChat(ctx, userId, orgId, req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
}

// SkillConversationSave
//
//	@Tags			resource.skill
//	@Summary		Skill发送到资源库
//	@Description	Skill发送到资源库
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SkillConversationSaveReq	true	"会话ID和保存的技能ID"
//	@Success		200		{object}	response.Response{}
//	@Router			/agent/skill/conversation/save [post]
func SkillConversationSave(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.SkillConversationSaveReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	_, err := service.SkillConversationSave(ctx, userId, orgId, req)
	gin_util.Response(ctx, nil, err)
}
