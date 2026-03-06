package v1

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetAgentSkillList
//
//	@Tags			resource.skill
//	@Summary		获取skill模板列表
//	@Description	获取skill模板列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill模板名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.SkillDetail}}
//	@Router			/agent/skill/list [get]
func GetAgentSkillList(ctx *gin.Context) {
	resp, err := service.GetAgentSkillList(ctx, ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetAgentSkillDetail
//
//	@Tags			resource.skill
//	@Summary		获取skill模板详情
//	@Description	获取skill模板详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill模板ID"
//	@Success		200		{object}	response.Response{data=response.SkillDetail}
//	@Router			/agent/skill/detail [get]
func GetAgentSkillDetail(ctx *gin.Context) {
	resp, err := service.GetAgentSkillDetail(ctx, ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// DownloadAgentSkill
//
//	@Tags			resource.skill
//	@Summary		下载skill模板
//	@Description	下载skill模板
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			skillId	query		string	true	"skill模板ID"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/download [get]
func DownloadAgentSkill(ctx *gin.Context) {
	fileName := fmt.Sprintf("%s.zip", ctx.Query("skillId"))
	resp, err := service.DownloadAgentSkill(ctx, ctx.Query("skillId"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 设置响应头
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	// 直接写入字节数据
	ctx.Data(http.StatusOK, "application/octet-stream", resp)
}

// GetCustomSkillList
//
//	@Tags			resource.skill
//	@Summary		获取自定义skill列表
//	@Description	获取自定义skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.CustomSkillDetail}}
//	@Router			/agent/skill/custom/list [get]
func GetCustomSkillList(ctx *gin.Context) {
	resp, err := service.GetCustomSkillList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetCustomSkillDetail
//
//	@Tags			resource.skill
//	@Summary		获取自定义skill详情
//	@Description	获取自定义skill详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	query		string	true	"skill ID"
//	@Success		200		{object}	response.Response{data=response.CustomSkillDetail}
//	@Router			/agent/skill/custom/detail [get]
func GetCustomSkillDetail(ctx *gin.Context) {
	resp, err := service.GetCustomSkill(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("skillId"))
	gin_util.Response(ctx, resp, err)
}

// CreateCustomSkill
//
//	@Tags			resource.skill
//	@Summary		创建自定义skill
//	@Description	创建自定义skill
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateCustomSkillReq	true	"自定义skill信息"
//	@Success		200		{object}	response.Response{data=response.CustomSkillIDResp}
//	@Router			/agent/skill/custom [post]
func CreateCustomSkill(ctx *gin.Context) {
	var req request.CreateCustomSkillReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateCustomSkill(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DeleteCustomSkill
//
//	@Tags			resource.skill
//	@Summary		删除自定义skill
//	@Description	删除自定义skill
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			skillId	body		request.DeleteCustomSkillReq	true	"skill ID"
//	@Success		200		{object}	response.Response
//	@Router			/agent/skill/custom [delete]
func DeleteCustomSkill(ctx *gin.Context) {
	var req request.DeleteCustomSkillReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteCustomSkill(ctx, req.SkillId)
	gin_util.Response(ctx, nil, err)
}
