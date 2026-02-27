package middleware

import (
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/UnicomAI/wanwu/pkg/gin-util/route"
)

func Init() {

	mid.InitWrapper(Record)

	// --- openapi ---
	mid.NewSub("openapi", "对外提供原子能力", route.PermNone, false, false)

	// --- callback ---
	mid.NewSub("callback", "系统内部调用", route.PermNone, false, false)

	// --- openurl ---
	mid.NewSub("openurl", "智能体Url", route.PermNone, false, false)

	// --- guest ---
	mid.NewSub("guest", "", route.PermNone, false, false)

	// --- common ---
	mid.NewSub("common", "", route.PermNeedEnable, false, false, JWTUser, CheckUserEnable)

	// --- model ---
	mid.NewSub("model", "模型服务", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// model.model_management
	mid.Sub("model").NewSub("model_management", "模型管理", route.PermNeedCheck, true, true)

	// --- resource ---
	mid.NewSub("resource", "资源库", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// resource.knowledge
	mid.Sub("resource").NewSub("knowledge", "知识库", route.PermNeedCheck, true, true)

	// resource.mcp
	mid.Sub("resource").NewSub("mcp", "MCP服务", route.PermNeedCheck, true, true)

	// resource.tool
	mid.Sub("resource").NewSub("tool", "工具", route.PermNeedCheck, true, true)

	// resource.prompt
	mid.Sub("resource").NewSub("prompt", "提示词", route.PermNeedCheck, true, true)

	// resource.safety
	mid.Sub("resource").NewSub("safety", "安全护栏", route.PermNeedCheck, true, true)

	// resource.skill
	mid.Sub("resource").NewSub("skill", "Skills", route.PermNeedCheck, true, true)

	// --- app ---
	mid.NewSub("app", "应用开发", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// app.rag
	mid.Sub("app").NewSub("rag", "知识问答", route.PermNeedCheck, true, true)

	// app.workflow
	mid.Sub("app").NewSub("workflow", "工作流", route.PermNeedCheck, true, true)

	// app.agent
	mid.Sub("app").NewSub("agent", "智能体", route.PermNeedCheck, true, true)

	// --- exploration ---
	mid.NewSub("exploration", "探索广场", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// exploration.mcp
	mid.Sub("exploration").NewSub("mcp", "MCP广场", route.PermNeedCheck, true, true)

	// exploration.app
	mid.Sub("exploration").NewSub("app", "应用广场", route.PermNeedCheck, true, true)

	// exploration.template
	mid.Sub("exploration").NewSub("template", "模板广场", route.PermNeedCheck, true, true)

	// --- operation ---
	mid.NewSub("operation", "运营管理", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// operation.statistic_client
	//mid.Sub("operation").NewSub("statistic_client", "用户统计", route.PermNeedCheck, true, true)

	// operation.oauth
	mid.Sub("operation").NewSub("oauth", "OAuth密钥管理", route.PermNeedCheck, true, true)

	// --- api_key ---
	mid.NewSub("api_key", "API Key管理", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// api_key.api_key_management
	mid.Sub("api_key").NewSub("api_key_management", "API Key管理", route.PermNeedCheck, true, true)

	// --- permission ---
	mid.NewSub("permission", "组织管理", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// permission.user
	mid.Sub("permission").NewSub("user", "用户", route.PermNeedCheck, true, true)

	// permission.org
	mid.Sub("permission").NewSub("org", "组织", route.PermNeedCheck, true, true)

	// permission.role
	mid.Sub("permission").NewSub("role", "角色", route.PermNeedCheck, true, true)

	// --- setting ---
	mid.NewSub("setting", "平台配置", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

}
