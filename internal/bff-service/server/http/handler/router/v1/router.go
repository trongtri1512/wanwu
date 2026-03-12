package v1

import (
	"github.com/gin-gonic/gin"
)

func Register(apiV1 *gin.RouterGroup) {
	// guest
	registerGuest(apiV1)

	// common
	registerCommon(apiV1)

	// callback
	registerV1Callback(apiV1)

	// model
	registerModel(apiV1)

	// knowledge
	registerKnowledge(apiV1)

	// mcp square
	registerMCPSquare(apiV1)

	// tool
	registerTool(apiV1)

	// safety
	registerSafety(apiV1)

	// skill
	registerAgentSkill(apiV1)

	// rag
	registerRag(apiV1)

	// workflow
	registerWorkflow(apiV1)

	// assistant
	registerAssistant(apiV1)

	// exploration
	registerExploration(apiV1)

	// permission
	registerPermission(apiV1)

	// setting
	registerSetting(apiV1)

	// statistic_client
	// registerStatisticClient(apiV1)

	// oauth
	registerOauth(apiV1)

	// api_key
	registerAPIKey(apiV1)

	// model_statistic
	registerModelStatistic(apiV1)
}
