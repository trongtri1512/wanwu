package app

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
)

func (s *Service) GetModelStatistic(ctx context.Context, req *app_service.GetModelStatisticReq) (*app_service.ModelStatistic, error) {
	stats, err := s.cli.GetModelStatistic(ctx, req.UserId, req.OrgId, req.StartDate, req.EndDate, req.ModelIds, req.ModelType)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatistic(stats), nil
}

func (s *Service) GetModelStatisticList(ctx context.Context, req *app_service.GetModelStatisticListReq) (*app_service.GetModelStatisticListResp, error) {
	stats, err := s.cli.GetModelStatisticList(ctx, req.UserId, req.OrgId, req.StartDate, req.EndDate, req.ModelIds, req.ModelType, toOffset(req), req.PageSize)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return convertModelStatisticList(stats), nil
}

func (s *Service) RecordModelStatistic(ctx context.Context, req *app_service.RecordModelStatisticReq) (*emptypb.Empty, error) {
	err := s.cli.RecordModelStatistic(ctx, req.UserId, req.OrgId, req.ModelId, req.Model, req.ModelType,
		req.PromptTokens, req.CompletionTokens, req.TotalTokens, req.FirstTokenLatency, req.Costs, req.IsSuccess, req.IsStream, req.Provider)
	if err != nil {
		return nil, errStatus(errs.Code_AppModelRecord, err)
	}
	return &emptypb.Empty{}, nil
}

func convertModelStatistic(stats *orm.ModelStatistic) *app_service.ModelStatistic {
	return &app_service.ModelStatistic{
		Overview: convertModelStatisticOverview(stats.Overview),
		Trend:    convertModelStatisticTrend(stats.Trend),
	}
}

func convertModelStatisticOverview(overview orm.ModelStatisticOverview) *app_service.ModelStatisticOverview {
	return &app_service.ModelStatisticOverview{
		CallCount:            convertModelStatisticOverviewItem(overview.CallCount),
		CallFailure:          convertModelStatisticOverviewItem(overview.CallFailure),
		TotalTokens:          convertModelStatisticOverviewItem(overview.TotalTokens),
		CompletionTokens:     convertModelStatisticOverviewItem(overview.CompletionTokens),
		PromptTokens:         convertModelStatisticOverviewItem(overview.PromptTokens),
		AvgCosts:             convertModelStatisticOverviewItem(overview.AvgCosts),
		AvgFirstTokenLatency: convertModelStatisticOverviewItem(overview.AvgFirstTokenLatency),
	}
}

func convertModelStatisticOverviewItem(item orm.ModelStatisticOverviewItem) *app_service.ModelStatisticOverviewItem {
	return &app_service.ModelStatisticOverviewItem{
		Value:            item.Value,
		PeriodOverPeriod: item.PeriodOverperiod,
	}
}

func convertModelStatisticTrend(trend orm.ModelStatisticTrend) *app_service.ModelStatisticTrend {
	return &app_service.ModelStatisticTrend{
		ModelCalls:  convertStatisticChart(trend.ModelCalls),
		TokensUsage: convertStatisticChart(trend.TokensUsage),
	}
}

func convertStatisticChart(chart orm.StatisticChart) *common.StatisticChart {
	pbChart := &common.StatisticChart{
		TableName:  chart.Name,
		ChartLines: make([]*common.StatisticChartLine, 0, len(chart.Lines)),
	}
	for _, line := range chart.Lines {
		pbLine := &common.StatisticChartLine{
			LineName: line.Name,
			Items:    make([]*common.StatisticChartLineItem, 0, len(line.Items)),
		}
		for _, item := range line.Items {
			pbLine.Items = append(pbLine.Items, &common.StatisticChartLineItem{
				Key:   item.Key,
				Value: item.Value,
			})
		}
		pbChart.ChartLines = append(pbChart.ChartLines, pbLine)
	}
	return pbChart
}

func convertModelStatisticList(list *orm.ModelStatisticList) *app_service.GetModelStatisticListResp {
	items := make([]*app_service.ModelStatisticItem, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, &app_service.ModelStatisticItem{
			ModelId:              item.ModelId,
			Model:                item.Model,
			Provider:             item.Provider,
			OrgId:                item.OrgId,
			CallCount:            item.CallCount,
			CallFailure:          item.CallFailure,
			FailureRate:          item.FailureRate,
			PromptTokens:         item.PromptTokens,
			CompletionTokens:     item.CompletionTokens,
			TotalTokens:          item.TotalTokens,
			AvgCosts:             item.AvgCosts,
			AvgFirstTokenLatency: item.AvgFirstTokenLatency,
		})
	}
	return &app_service.GetModelStatisticListResp{
		Items: items,
		Total: list.Total,
	}
}
