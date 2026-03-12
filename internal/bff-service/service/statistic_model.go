package service

import (
	"context"
	"math"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetModelStatistic(ctx *gin.Context, userId, orgId, startDate, endDate string, modelIds []string, modelType string) (*response.ModelStatistic, error) {
	if modelType == "" {
		modelType = mp.ModelTypeLLM
	}
	resp, err := app.GetModelStatistic(ctx.Request.Context(), &app_service.GetModelStatisticReq{
		UserId:    userId,
		OrgId:     orgId,
		StartDate: startDate,
		EndDate:   endDate,
		ModelIds:  modelIds,
		ModelType: modelType,
	})
	if err != nil {
		return nil, err
	}
	return &response.ModelStatistic{
		Overview: response.ModelStatisticOverview{
			CallCountTotal:        convertModelStatisticOverviewItem(resp.Overview.GetCallCount()),
			CallFailureTotal:      convertModelStatisticOverviewItem(resp.Overview.GetCallFailure()),
			TotalTokensTotal:      convertModelStatisticOverviewItem(resp.Overview.GetTotalTokens()),
			PromptTokensTotal:     convertModelStatisticOverviewItem(resp.Overview.GetPromptTokens()),
			CompletionTokensTotal: convertModelStatisticOverviewItem(resp.Overview.GetCompletionTokens()),
			AvgCosts:              convertModelStatisticOverviewItem(resp.Overview.GetAvgCosts()),
			AvgFirstTokenLatency:  convertModelStatisticOverviewItem(resp.Overview.GetAvgFirstTokenLatency()),
		},
		Trend: response.ModelStatisticTrend{
			ModelCalls:  convertStatisticChart(ctx, resp.Trend.GetModelCalls()),
			TokensUsage: convertStatisticChart(ctx, resp.Trend.GetTokensUsage()),
		},
	}, nil
}

func GetModelStatisticList(ctx *gin.Context, userId, orgId, startDate, endDate string, modelIds []string, modelType string, page, pageSize int32) (*response.PageResult, error) {
	if modelType == "" {
		modelType = mp.ModelTypeLLM
	}
	resp, err := app.GetModelStatisticList(ctx.Request.Context(), &app_service.GetModelStatisticListReq{
		UserId:    userId,
		OrgId:     orgId,
		StartDate: startDate,
		EndDate:   endDate,
		ModelIds:  modelIds,
		ModelType: modelType,
		PageNo:    page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]response.ModelStatisticItem, 0, len(resp.Items))
	// 收集items中的orgId然后获取OrgIds对应的OrgNames
	var orgIds []string
	for _, item := range resp.Items {
		orgIds = append(orgIds, item.OrgId)
	}
	// 调用IAM服务获取组织信息
	orgResp, err := iam.GetOrgByOrgIDs(ctx, &iam_service.GetOrgByOrgIDsReq{
		OrgIds: orgIds,
	})
	if err != nil {
		return nil, err
	}
	// 创建orgId到orgName的映射
	orgNameMap := make(map[string]string)
	if orgResp != nil && orgResp.Orgs != nil {
		for _, org := range orgResp.Orgs {
			orgNameMap[org.Id] = org.Name
		}
	}
	for _, item := range resp.Items {
		roundedFailureRate := float32(math.Round(float64(item.FailureRate)*100) / 100)
		roundedAvgCosts := float32(math.Round(float64(item.AvgCosts)*100) / 100)
		roundedAvgFirstTokenLatency := float32(math.Round(float64(item.AvgFirstTokenLatency)*100) / 100)
		items = append(items, response.ModelStatisticItem{
			ModelId:              item.ModelId,
			Model:                item.Model,
			Provider:             item.Provider,
			OrgName:              orgNameMap[item.OrgId],
			CallCount:            item.CallCount,
			CallFailure:          item.CallFailure,
			FailureRate:          roundedFailureRate,
			PromptTokens:         item.PromptTokens,
			CompletionTokens:     item.CompletionTokens,
			TotalTokens:          item.TotalTokens,
			AvgCosts:             roundedAvgCosts,
			AvgFirstTokenLatency: roundedAvgFirstTokenLatency,
		})
	}
	return &response.PageResult{
		List:     items,
		Total:    int64(resp.Total),
		PageNo:   int(page),
		PageSize: int(pageSize),
	}, nil
}

func ExportModelStatisticList(ctx *gin.Context, userId, orgId, startDate, endDate string, modelIds []string, modelType string) (*excelize.File, error) {
	resp, err := GetModelStatisticList(ctx, userId, orgId, startDate, endDate, modelIds, modelType, -1, -1)
	if err != nil {
		return nil, err
	}
	return writeModelListExcel(resp.List.([]response.ModelStatisticItem))
}

func recordModelStatistic(_ *gin.Context, modelInfo *model_service.ModelInfo, isSuccess bool, promptTokens, completionTokens, totalTokens, costs, firstTokenLatency int, isStream bool) {
	go func() {
		defer util.PrintPanicStack()
		_, err := app.RecordModelStatistic(context.Background(), &app_service.RecordModelStatisticReq{
			UserId:            modelInfo.UserId,
			OrgId:             modelInfo.OrgId,
			ModelId:           modelInfo.ModelId,
			Model:             modelInfo.Model,
			Provider:          modelInfo.Provider,
			ModelType:         modelInfo.ModelType,
			PromptTokens:      int64(promptTokens),
			CompletionTokens:  int64(completionTokens),
			TotalTokens:       int64(totalTokens),
			FirstTokenLatency: int32(firstTokenLatency),
			Costs:             int32(costs),
			IsSuccess:         isSuccess,
			IsStream:          isStream,
		})
		if err != nil {
			log.Errorf("record modelId %v modelName %v modelType %v statistic err:%v", modelInfo.ModelId, modelInfo.Model, modelInfo.ModelType, err)
		}
	}()
}

func convertModelStatisticOverviewItem(item *app_service.ModelStatisticOverviewItem) response.StatisticOverviewItem {
	return response.StatisticOverviewItem{
		Value:            item.Value,
		PeriodOverPeriod: item.PeriodOverPeriod,
	}
}

func writeModelListExcel(items []response.ModelStatisticItem) (*excelize.File, error) {
	sheet := "模型统计列表"
	title := []any{"模型ID", "模型", "模型供应商", "组织", "调用次数", "调用失败次数", "失败率", "Prompt Tokens", "Completion Tokens", "总Tokens", "平均耗时(非流式)", "平均首Token时延(流式)"}
	var rows [][]any
	for _, item := range items {
		rows = append(rows, []any{
			item.ModelId,
			item.Model,
			item.Provider,
			item.OrgName,
			item.CallCount,
			item.CallFailure,
			item.FailureRate,
			item.PromptTokens,
			item.CompletionTokens,
			item.TotalTokens,
			item.AvgCosts,
			item.AvgFirstTokenLatency,
		})
	}
	return writeExcel(sheet, title, rows)
}

func writeExcel(sheet string, title []any, rows [][]any) (*excelize.File, error) {
	f := excelize.NewFile()
	index, err := f.NewSheet(sheet)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	if err := writeExcelRow(f, sheet, 1, title); err != nil {
		return nil, err
	}
	for i, row := range rows {
		if err := writeExcelRow(f, sheet, i+2, row); err != nil {
			return nil, err
		}
	}
	return f, nil
}

func writeExcelRow(f *excelize.File, sheet string, row int, values []any) error {
	for col, value := range values {
		cell, err := excelize.CoordinatesToCellName(col+1, row)
		if err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell, value); err != nil {
			return err
		}
	}
	return nil
}
