package openapi

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/openapi"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func Register(openAPI *gin.RouterGroup) {
	// agent
	mid.Sub("openapi").Reg(openAPI, "/agent/conversation", http.MethodPost, openapi.CreateAgentConversation, "智能体创建对话OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeAgent))
	mid.Sub("openapi").Reg(openAPI, "/agent/chat", http.MethodPost, openapi.ChatAgent, "智能体问答OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeAgent))
	// rag
	mid.Sub("openapi").Reg(openAPI, "/rag/chat", http.MethodPost, openapi.ChatRag, "文本问答OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeRag))
	// workflow
	mid.Sub("openapi").Reg(openAPI, "/workflow/run", http.MethodPost, openapi.WorkflowRun, "工作流OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeWorkflow))
	mid.Sub("openapi").Reg(openAPI, "/workflow/file/upload", http.MethodPost, openapi.WorkflowFileUpload, "工作流OpenAPI文件上传", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))
	// chatflow
	mid.Sub("openapi").Reg(openAPI, "/chatflow/conversation", http.MethodPost, openapi.CreateChatflowConversation, "对话流创建对话OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))
	mid.Sub("openapi").Reg(openAPI, "/chatflow/conversation/message/list", http.MethodPost, openapi.GetConversationMessageList, "对话流根据conversationId获取历史对话", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))
	mid.Sub("openapi").Reg(openAPI, "/chatflow/chat", http.MethodPost, openapi.ChatflowChat, "对话流OpenAPI", middleware.AuthOpenAPIKey(constant.OpenAPITypeChatflow))
	// knowledge
	mid.Sub("openapi").Reg(openAPI, "/file/upload/direct", http.MethodPost, openapi.DirectUploadFiles, "直接上传文件", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge))
	mid.Sub("openapi").Reg(openAPI, "/knowledge", http.MethodPost, openapi.CreateKnowledge, "新建知识库", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthModel([]string{"embeddingModelInfo.modelId", "knowledgeGraph.llmModelId"}))
	mid.Sub("openapi").Reg(openAPI, "/knowledge", http.MethodPut, openapi.UpdateKnowledge, "更新知识库", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))
	mid.Sub("openapi").Reg(openAPI, "/knowledge", http.MethodDelete, openapi.DeleteKnowledge, "删除知识库", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeSystem))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/select", http.MethodPost, openapi.GetKnowledgeSelect, "查询知识库列表", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/doc/config", http.MethodGet, openapi.GetDocConfig, "获取文档配置信息", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/doc/list", http.MethodPost, openapi.GetDocList, "获取文档列表", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/doc/import", http.MethodPost, openapi.ImportDoc, "上传文档", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/doc/update/config", http.MethodPost, openapi.UpdateDocConfig, "更新文档配置", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/doc/import/tip", http.MethodGet, openapi.GetDocImportTip, "获取知识库文档上传状态", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/doc/export", http.MethodPost, openapi.ExportKnowledgeDoc, "知识库文档导出", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/export/record/list", http.MethodGet, openapi.GetKnowledgeExportRecordList, "获取知识库导出记录列表", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeView))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/export/record", http.MethodDelete, openapi.DeleteKnowledgeExportRecord, "删除知识库库导出记录", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/doc", http.MethodDelete, openapi.DeleteDoc, "删除文档", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthKnowledge("knowledgeId", middleware.KnowledgeEdit))
	mid.Sub("openapi").Reg(openAPI, "/knowledge/hit", http.MethodPost, openapi.KnowledgeHit, "知识库命中测试", middleware.AuthOpenAPIKey(constant.OpenAPITypeKnowledge), middleware.AuthModel([]string{"knowledgeMatchParams.rerankModelId"}))

	// mcp server
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/sse", http.MethodGet, openapi.GetMCPServerSSE, "新建MCP服务sse连接", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/message", http.MethodPost, openapi.GetMCPServerMessage, "获取MCP服务sse消息", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/streamable", http.MethodGet, openapi.GetMCPServerStreamable, "获取MCP服务streamable消息(GET)", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))
	mid.Sub("openapi").Reg(openAPI, "/mcp/server/streamable", http.MethodPost, openapi.GetMCPServerStreamable, "获取MCP服务streamable消息(POST)", middleware.AuthAppKeyByQuery(constant.AppTypeMCPServer))

	// oauth
	mid.Sub("openapi").Reg(openAPI, "/oauth/jwks", http.MethodGet, openapi.OAuthJWKS, "JWT公钥")
	mid.Sub("openapi").Reg(openAPI, "/oauth/login", http.MethodGet, openapi.OAuthLogin, "OAuth登录授权")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/authorize", http.MethodGet, openapi.OAuthAuthorize, "获取授权码")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/token", http.MethodPost, openapi.OAuthToken, "授权码获取token")
	mid.Sub("openapi").Reg(openAPI, "/oauth/code/token/refresh", http.MethodPost, openapi.OAuthRefresh, "刷新Access Token")
	mid.Sub("openapi").Reg(openAPI, "/.well-known/openid-configuration", http.MethodGet, openapi.OAuthConfig, "返回Endpoint配置")
	// oauth user
	mid.Sub("openapi").Reg(openAPI, "/oauth/userinfo", http.MethodGet, openapi.OAuthGetUserInfo, "OAuth获取用户信息", middleware.JWTOAuthAccess)
}
