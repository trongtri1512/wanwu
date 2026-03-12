package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetModelStatistic
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型统计数据
//	@Description	获取模型统计数据
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			startDate	query		string	true	"开始时间（格式yyyy-mm-dd）"
//	@Param			endDate		query		string	true	"结束时间（格式yyyy-mm-dd）"
//	@Param			models		query		string	false	"模型ID列表"
//	@Param			modelType	query		string	true	"模型类型"
//	@Success		200			{object}	response.Response{data=response.ModelStatistic}
//	@Router			/statistic/model [get]
func GetModelStatistic(ctx *gin.Context) {
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	modelIds := ctx.Query("models")
	modelType := ctx.Query("modelType")
	var modelIdList []string
	if modelIds != "" {
		modelIdList = strings.Split(modelIds, ",")
	}
	resp, err := service.GetModelStatistic(ctx, getUserID(ctx), getOrgID(ctx), startDate, endDate, modelIdList, modelType)
	gin_util.Response(ctx, resp, err)
}

// GetModelStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		获取模型统计列表
//	@Description	获取模型统计列表（分页）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			startDate	query		string	true	"开始时间（格式yyyy-mm-dd）"
//	@Param			endDate		query		string	true	"结束时间（格式yyyy-mm-dd）"
//	@Param			models		query		string	false	"模型ID列表"
//	@Param			modelType	query		string	true	"模型类型"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.ModelStatisticItem}}
//	@Router			/statistic/model/list [get]
func GetModelStatisticList(ctx *gin.Context) {
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	modelIds := ctx.Query("models")
	modelType := ctx.Query("modelType")
	var modelIdList []string
	if modelIds != "" {
		modelIdList = strings.Split(modelIds, ",")
	}
	resp, err := service.GetModelStatisticList(ctx, getUserID(ctx), getOrgID(ctx), startDate, endDate, modelIdList, modelType, getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// ExportModelStatisticList
//
//	@Tags			app_observability.statistic
//	@Summary		导出模型统计列表
//	@Description	导出模型统计列表数据
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			startDate	query		string	true	"开始时间（格式yyyy-mm-dd）"
//	@Param			endDate		query		string	true	"结束时间（格式yyyy-mm-dd）"
//	@Param			models		query		string	false	"模型ID列表"
//	@Param			modelType	query		string	true	"模型类型"
//	@Success		200			{object}	response.Response
//	@Router			/statistic/model/export [get]
func ExportModelStatisticList(ctx *gin.Context) {
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	modelIds := ctx.Query("models")
	modelType := ctx.Query("modelType")
	var modelIdList []string
	if modelIds != "" {
		modelIdList = strings.Split(modelIds, ",")
	}
	file, err := service.ExportModelStatisticList(ctx, getUserID(ctx), getOrgID(ctx), startDate, endDate, modelIdList, modelType)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	fileName := fmt.Sprintf("模型统计列表_%v-%v.xlsx", startDate, endDate)
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if _, err := file.WriteTo(ctx.Writer); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}
