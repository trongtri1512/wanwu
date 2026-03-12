package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerModelStatistic(apiV1 *gin.RouterGroup) {
	mid.Sub("app_observability.statistic").Reg(apiV1, "/statistic/model", http.MethodGet, v1.GetModelStatistic, "获取模型统计数据")
	mid.Sub("app_observability.statistic").Reg(apiV1, "/statistic/model/list", http.MethodGet, v1.GetModelStatisticList, "获取模型统计列表")
	mid.Sub("app_observability.statistic").Reg(apiV1, "/statistic/model/export", http.MethodGet, v1.ExportModelStatisticList, "导出模型统计列表")
}
