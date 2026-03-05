package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerAgentSkill(apiV1 *gin.RouterGroup) {
	// skills 模板（只读）
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/list", http.MethodGet, v1.GetAgentSkillList, "获取skill模板列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/detail", http.MethodGet, v1.GetAgentSkillDetail, "获取skill模板详情")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/download", http.MethodGet, v1.DownloadAgentSkill, "下载skill模板")

	// Skill Conversation 相关接口
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/conversation", http.MethodPost, v1.CreateSkillConversation, "创建Skill生成会话")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/conversation", http.MethodDelete, v1.DeleteSkillConversation, "删除Skill生成会话")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/conversation/detail", http.MethodGet, v1.GetSkillConversationDetail, "获取Skill生成会话详情")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/conversation/list", http.MethodGet, v1.GetSkillConversationList, "获取Skill生成会话列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/conversation/chat", http.MethodPost, v1.SkillConversationChat, "Skill生成流式对话")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/conversation/save", http.MethodPost, v1.SkillConversationSave, "将Skill发送到资源库")
	// 自定义 skill（CRUD）
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/list", http.MethodGet, v1.GetCustomSkillList, "获取自定义skill列表")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom/detail", http.MethodGet, v1.GetCustomSkillDetail, "获取自定义skill详情")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom", http.MethodPost, v1.CreateCustomSkill, "创建自定义skill")
	mid.Sub("resource.skill").Reg(apiV1, "/agent/skill/custom", http.MethodDelete, v1.DeleteCustomSkill, "删除自定义skill")
}
