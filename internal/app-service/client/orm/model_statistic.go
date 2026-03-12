package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

const luaUpdateModelStats = `
local key = KEYS[1]
local field = KEYS[2]
local delta = cjson.decode(ARGV[1])
local expire = tonumber(ARGV[2])

local current = redis.call('HGET', key, field)

if current then
	local record = cjson.decode(current)
	record.promptTokens = record.promptTokens + delta.promptTokens
	record.completionTokens = record.completionTokens + delta.completionTokens
	record.totalTokens = record.totalTokens + delta.totalTokens
	record.firstTokenLatency = record.firstTokenLatency + delta.firstTokenLatency
	record.costs = record.costs + delta.costs
	record.callCount = record.callCount + 1
	if delta.isFailure == 1 then
		record.callFailure = record.callFailure + 1
	end
	if delta.isStream == 1 then
		record.streamCount = record.streamCount + 1
		if delta.isFailure == 1 then
			record.streamFailure = record.streamFailure + 1
		end
	else
		record.nonStreamCount = record.nonStreamCount + 1
		if delta.isFailure == 1 then
			record.nonStreamFailure = record.nonStreamFailure + 1
		end
	end
	redis.call('HSET', key, field, cjson.encode(record))
else
	redis.call('HSET', key, field, ARGV[1])
end

redis.call('EXPIRE', key, expire)
return 1
`

// ModelRecordStats 模型记录统计结构体
type ModelRecordStats struct {
	Model             string `json:"model"`             // 模型名称
	Provider          string `json:"provider"`          // 模型供应商
	ModelType         string `json:"modelType"`         // 模型类型
	PromptTokens      int64  `json:"promptTokens"`      // 提示词token
	CompletionTokens  int64  `json:"completionTokens"`  // 问答token
	TotalTokens       int64  `json:"totalTokens"`       // 总token
	CallCount         int32  `json:"callCount"`         // 调用次数
	CallFailure       int32  `json:"callFailure"`       // 调用失败次数
	StreamCount       int32  `json:"streamCount"`       // 流式调用次数
	NonStreamCount    int32  `json:"nonStreamCount"`    // 非流失调用次数
	StreamFailure     int32  `json:"streamFailure"`     // 流式调用失败次数
	NonStreamFailure  int32  `json:"nonStreamFailure"`  // 非流式调用失败次数
	FirstTokenLatency int32  `json:"firstTokenLatency"` // 首token时延(总)
	Costs             int32  `json:"costs"`             // 耗时(总)
}

// GetModelStatistic 获取模型统计（概览+趋势）
func (c *Client) GetModelStatistic(ctx context.Context, userId, orgId, startDate, endDate string, modelIds []string, modelType string) (*ModelStatistic, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_model_statistic_get", fmt.Errorf("startDate %v greater than endDate %v", startDate, endDate).Error())
	}

	today := util.Time2Date(time.Now().UnixMilli())
	if err := updateModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync model stats for today %v err: %v", today, err)
	}

	overview, err := statisticModelStatsOverview(ctx, c.db, userId, orgId, startDate, endDate, modelIds, modelType)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_get", err.Error())
	}

	trend, err := statisticModelStatsTrend(ctx, c.db, userId, orgId, startDate, endDate, modelIds, modelType)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_get", err.Error())
	}

	return &ModelStatistic{
		Overview: *overview,
		Trend:    *trend,
	}, nil
}

func (c *Client) GetModelStatisticList(ctx context.Context, userId, orgId, startDate, endDate string, modelIds []string, modelType string, offset, limit int32) (*ModelStatisticList, *errs.Status) {
	if startDate > endDate {
		return nil, toErrStatus("app_model_statistic_list_get", fmt.Errorf("startDate %v greater than endDate %v", startDate, endDate).Error())
	}

	today := util.Time2Date(time.Now().UnixMilli())
	if err := updateModelStats(ctx, today, c.db); err != nil {
		log.Errorf("sync model stats for today %v err: %v", today, err)
	}

	items, total, err := getModelStatisticList(ctx, c.db, userId, orgId, startDate, endDate, modelIds, modelType, offset, limit)
	if err != nil {
		return nil, toErrStatus("app_model_statistic_list_get", err.Error())
	}

	return &ModelStatisticList{
		Items: items,
		Total: total,
	}, nil
}

// RecordModelStatistic 记录模型统计数据
func (c *Client) RecordModelStatistic(ctx context.Context, userId, orgId, modelId, model, modelType string,
	promptTokens, completionTokens, totalTokens int64, firstTokenLatency, costs int32, isSuccess bool, isStream bool, provider string) *errs.Status {
	err := recordModelStatistic(ctx, userId, orgId, modelId, model, modelType,
		promptTokens, completionTokens, totalTokens, firstTokenLatency, costs, isSuccess, isStream, provider)
	if err != nil {
		return toErrStatus("app_model_record_statistic", err.Error())
	}
	return nil
}

// --- internal ---

// getRedisModelStatsKey 获取Redis模型统计key
// e.g. modelrecord|2006-01-02
func getRedisModelStatsKey(date string) string {
	return fmt.Sprintf("modelrecord|%s", date)
}

// getRedisModelStatsItemField 获取Redis模型统计字段
// e.g. modelId|userId|orgId|provider
func getRedisModelStatsItemField(modelId, userId, orgId, provider string) string {
	return fmt.Sprintf("%s|%s|%s|%s", modelId, userId, orgId, provider)
}

// parseRedisModelStatsItem 解析Redis模型统计字段
func parseRedisModelStatsItem(field string) (string, string, string, string, bool) {
	parts := strings.Split(field, "|")
	if len(parts) != 4 {
		return "", "", "", "", false
	}
	return parts[0], parts[1], parts[2], parts[3], true
}

// getRedisModelStatsItemValue 获取Redis模型统计值
func getRedisModelStatsItemValue(value string) (*ModelRecordStats, error) {
	record := &ModelRecordStats{}
	if err := json.Unmarshal([]byte(value), record); err != nil {
		return nil, fmt.Errorf("unmarshal value %v err: %v", value, err)
	}
	return record, nil
}

func recordModelStatistic(ctx context.Context, userId, orgId, modelId, model, modelType string,
	promptTokens, completionTokens, totalTokens int64, firstTokenLatency, costs int32, isSuccess bool, isStream bool, provider string) error {
	today := util.Time2Date(time.Now().UnixMilli())
	key := getRedisModelStatsKey(today)
	field := getRedisModelStatsItemField(modelId, userId, orgId, provider)

	callFailure := 0
	streamCount := 0
	streamFailure := 0
	nonStreamCount := 0
	nonStreamFailure := 0

	if isStream {
		streamCount = 1
		if !isSuccess {
			streamFailure = 1
		}
	} else {
		nonStreamCount = 1
		if !isSuccess {
			nonStreamFailure = 1
		}
	}
	if !isSuccess {
		callFailure = 1
	}

	delta := map[string]any{
		"model":             model,
		"provider":          provider,
		"modelType":         modelType,
		"promptTokens":      promptTokens,
		"completionTokens":  completionTokens,
		"totalTokens":       totalTokens,
		"callCount":         1,
		"callFailure":       callFailure,
		"streamCount":       streamCount,
		"streamFailure":     streamFailure,
		"nonStreamCount":    nonStreamCount,
		"nonStreamFailure":  nonStreamFailure,
		"firstTokenLatency": firstTokenLatency,
		"costs":             costs,
	}

	deltaJSON, _ := json.Marshal(delta)

	_, err := redis.App().Eval(ctx, luaUpdateModelStats, []string{key, field}, string(deltaJSON), 30*24*3600)
	if err != nil {
		return fmt.Errorf("redis eval err: %v", err)
	}
	return nil
}

// updateModelStats 从Redis更新模型统计数据到数据库
func updateModelStats(ctx context.Context, date string, db *gorm.DB) error {
	key := getRedisModelStatsKey(date)
	resultMap, err := redis.App().HGetAll(ctx, key)
	if err != nil {
		return fmt.Errorf("redis HGetAll key %v failed: %v", key, err)
	}
	if len(resultMap) == 0 {
		return nil
	}

	for _, item := range resultMap {
		if modelId, userId, orgId, provider, ok := parseRedisModelStatsItem(item.K); ok {
			record, err := getRedisModelStatsItemValue(item.V)
			if err != nil {
				log.Errorf("get model stat item %v err: %v", item.K, err)
				continue
			}
			if err := updateModelStatsByRecord(ctx, db, modelId, userId, orgId, provider, date, record); err != nil {
				log.Errorf("update date %v model failed for modelId %v userId %v orgId %v err: %v",
					date, modelId, userId, orgId, err)
			}
		}
	}
	return nil
}

// updateModelStatsByRecord 根据记录更新模型统计数据到数据库
func updateModelStatsByRecord(ctx context.Context, db *gorm.DB, modelId, userId, orgId, provider, date string, record *ModelRecordStats) error {
	var modelStat *model.ModelRecord
	if err := sqlopt.SQLOptions(
		sqlopt.WithModelID(modelId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithProvider(provider),
		sqlopt.WithDate(date),
	).Apply(db.WithContext(ctx)).First(&modelStat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.WithContext(ctx).Create(&model.ModelRecord{
				OrgID:             orgId,
				UserID:            userId,
				ModelID:           modelId,
				Model:             record.Model,
				Provider:          record.Provider,
				ModelType:         record.ModelType,
				PromptTokens:      record.PromptTokens,
				CompletionTokens:  record.CompletionTokens,
				TotalTokens:       record.TotalTokens,
				FirstTokenLatency: record.FirstTokenLatency,
				Costs:             record.Costs,
				StreamCount:       record.StreamCount,
				NonStreamCount:    record.NonStreamCount,
				CallCount:         record.CallCount,
				CallFailure:       record.CallFailure,
				NonStreamFailure:  record.NonStreamFailure,
				StreamFailure:     record.StreamFailure,
				Date:              date,
			}).Error
		}
		return err
	}

	// 检查是否需要更新
	if modelStat.PromptTokens == record.PromptTokens &&
		modelStat.CompletionTokens == record.CompletionTokens &&
		modelStat.TotalTokens == record.TotalTokens &&
		modelStat.CallCount == record.CallCount &&
		modelStat.NonStreamFailure == record.NonStreamFailure &&
		modelStat.StreamFailure == record.StreamFailure &&
		modelStat.CallFailure == record.CallFailure &&
		modelStat.StreamCount == record.StreamCount &&
		modelStat.NonStreamCount == record.NonStreamCount &&
		modelStat.FirstTokenLatency == record.FirstTokenLatency &&
		modelStat.Costs == record.Costs {
		return nil
	}

	return db.WithContext(ctx).Model(&modelStat).Updates(map[string]any{
		"prompt_tokens":       record.PromptTokens,
		"completion_tokens":   record.CompletionTokens,
		"total_tokens":        record.TotalTokens,
		"first_token_latency": record.FirstTokenLatency,
		"costs":               record.Costs,
		"stream_count":        record.StreamCount,
		"non_stream_count":    record.NonStreamCount,
		"call_count":          record.CallCount,
		"call_failure":        record.CallFailure,
		"stream_failure":      record.StreamFailure,
		"non_stream_failure":  record.NonStreamFailure,
	}).Error
}

func getModelStatisticList(ctx context.Context, db *gorm.DB, userId, orgId, startDate, endDate string, modelIds []string, modelType string, offset, limit int32) ([]ModelStatisticItem, int32, error) {
	opts := []sqlopt.SQLOption{
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithModelType(modelType),
		sqlopt.WithModelIds(modelIds),
	}
	var total int64
	countQuery := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).
		Model(&model.ModelRecord{}).
		Select("COUNT(DISTINCT model_id, user_id,org_id)")
	countQuery.Count(&total)
	var stats []model.ModelRecord
	query := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).
		Select("model_id,model, user_id,org_id,provider, SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(prompt_tokens) as prompt_tokens, SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens, SUM(costs) as costs, " +
			"SUM(first_token_latency) as first_token_latency, " +
			"SUM(stream_count) as stream_count, SUM(non_stream_count) as non_stream_count").
		Group("model_id, model,user_id,org_id,provider").Offset(int(offset)).Limit(int(limit))

	if err := query.Find(&stats).Error; err != nil {
		return nil, 0, fmt.Errorf("get model stat list err: %v", err)
	}

	items := make([]ModelStatisticItem, 0, len(stats))
	for _, stat := range stats {
		failureRate := float32(0)
		if stat.CallCount > 0 {
			failureRate = float32(stat.CallFailure) / float32(stat.CallCount) * 100
		}
		avgCosts := float32(0)
		if stat.NonStreamCount-stat.NonStreamFailure > 0 {
			avgCosts = float32(stat.Costs) / float32(stat.NonStreamCount-stat.NonStreamFailure)
		}
		avgFirstTokenLatency := float32(0)
		if stat.StreamCount-stat.StreamFailure > 0 {
			avgFirstTokenLatency = float32(stat.FirstTokenLatency) / float32(stat.StreamCount-stat.StreamFailure)
		}
		items = append(items, ModelStatisticItem{
			ModelId:              stat.ModelID,
			Model:                stat.Model,
			Provider:             stat.Provider,
			OrgId:                stat.OrgID,
			CallCount:            stat.CallCount,
			CallFailure:          stat.CallFailure,
			FailureRate:          failureRate,
			PromptTokens:         stat.PromptTokens,
			CompletionTokens:     stat.CompletionTokens,
			TotalTokens:          stat.TotalTokens,
			AvgCosts:             avgCosts,
			AvgFirstTokenLatency: avgFirstTokenLatency,
		})
	}

	return items, int32(total), nil
}

// statisticModelStatsOverview 统计模型概览数据（新接口）
func statisticModelStatsOverview(ctx context.Context, db *gorm.DB, userID, orgID, startDate, endDate string, modelIds []string, modelType string) (*ModelStatisticOverview, error) {
	prevPeriod, currPeriod, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	current, err := modelStatsByDateRange(ctx, db, userID, orgID, currPeriod, modelIds, modelType)
	if err != nil {
		return nil, err
	}

	previous, err := modelStatsByDateRange(ctx, db, userID, orgID, prevPeriod, modelIds, modelType)
	if err != nil {
		return nil, err
	}

	current.CallCount.PeriodOverperiod = calculatePoP(current.CallCount.Value, previous.CallCount.Value)
	current.CallFailure.PeriodOverperiod = calculatePoP(current.CallFailure.Value, previous.CallFailure.Value)
	current.TotalTokens.PeriodOverperiod = calculatePoP(current.TotalTokens.Value, previous.TotalTokens.Value)
	current.CompletionTokens.PeriodOverperiod = calculatePoP(current.CompletionTokens.Value, previous.CompletionTokens.Value)
	current.PromptTokens.PeriodOverperiod = calculatePoP(current.PromptTokens.Value, previous.PromptTokens.Value)
	current.AvgCosts.PeriodOverperiod = calculatePoP(current.AvgCosts.Value, previous.AvgCosts.Value)
	current.AvgFirstTokenLatency.PeriodOverperiod = calculatePoP(current.AvgFirstTokenLatency.Value, previous.AvgFirstTokenLatency.Value)

	return current, nil
}

// modelStatsByDateRange 按日期范围获取模型统计数据（新接口）
func modelStatsByDateRange(ctx context.Context, db *gorm.DB, userID, orgID string, dates []string, modelIds []string, modelType string) (*ModelStatisticOverview, error) {
	startDate, endDate := dates[0], dates[len(dates)-1]
	var stat model.ModelRecord
	opts := []sqlopt.SQLOption{
		sqlopt.WithOrgID(orgID),
		sqlopt.WithUserID(userID),
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	if err := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).
		Select("SUM(prompt_tokens) as prompt_tokens, " +
			"SUM(completion_tokens) as completion_tokens, " +
			"SUM(total_tokens) as total_tokens, " +
			"SUM(call_count) as call_count, " +
			"SUM(call_failure) as call_failure, " +
			"SUM(stream_count) as stream_count, " +
			"SUM(non_stream_count) as non_stream_count, " +
			"SUM(first_token_latency) as first_token_latency, " +
			"SUM(costs) as costs").
		First(&stat).Error; err != nil {
		return nil, fmt.Errorf("model stat [%v, %v] err: %v", startDate, endDate, err)
	}

	avgCosts := float32(0)
	if stat.NonStreamCount-stat.NonStreamFailure > 0 {
		avgCosts = float32(stat.Costs) / float32(stat.NonStreamCount-stat.NonStreamFailure)
	}
	avgFirstTokenLatency := float32(0)
	if stat.StreamCount-stat.StreamFailure > 0 {
		avgFirstTokenLatency = float32(stat.FirstTokenLatency) / float32(stat.StreamCount-stat.StreamFailure)
	}

	return &ModelStatisticOverview{
		CallCount:            ModelStatisticOverviewItem{Value: float32(stat.CallCount)},
		CallFailure:          ModelStatisticOverviewItem{Value: float32(stat.CallFailure)},
		TotalTokens:          ModelStatisticOverviewItem{Value: float32(stat.TotalTokens)},
		CompletionTokens:     ModelStatisticOverviewItem{Value: float32(stat.CompletionTokens)},
		PromptTokens:         ModelStatisticOverviewItem{Value: float32(stat.PromptTokens)},
		AvgCosts:             ModelStatisticOverviewItem{Value: avgCosts},
		AvgFirstTokenLatency: ModelStatisticOverviewItem{Value: avgFirstTokenLatency},
	}, nil
}

// statisticModelStatsTrend 统计模型趋势数据
func statisticModelStatsTrend(ctx context.Context, db *gorm.DB, userID, orgID, startDate, endDate string, modelIds []string, modelType string) (*ModelStatisticTrend, error) {
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return nil, err
	}
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return nil, err
	}

	var stats []model.ModelRecord
	opts := []sqlopt.SQLOption{
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
		sqlopt.WithModelIds(modelIds),
		sqlopt.WithModelType(modelType),
	}
	if err := sqlopt.SQLOptions(opts...).Apply(db).WithContext(ctx).
		Select("date, SUM(call_count) as call_count, SUM(call_failure) as call_failure, " +
			"SUM(total_tokens) as total_tokens, SUM(completion_tokens) as completion_tokens, " +
			"SUM(prompt_tokens) as prompt_tokens").
		Group("date").
		Order("date").
		Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("model stat trend err: %v", err)
	}

	dateMap := make(map[string]model.ModelRecord)
	for _, stat := range stats {
		dateMap[stat.Date] = stat
	}

	var callCountTotalLine, callSuccessLine, callFailureLine []StatisticChartLineItem
	var totalTokensLine, completionTokensLine, promptTokensLine []StatisticChartLineItem

	for _, date := range util.DateRange(startTs, endTs) {
		if stat, ok := dateMap[date]; ok {
			callSuccess := stat.CallCount - stat.CallFailure
			callCountTotalLine = append(callCountTotalLine, StatisticChartLineItem{
				Key:   date,
				Value: float32(stat.CallCount),
			})
			callSuccessLine = append(callSuccessLine, StatisticChartLineItem{
				Key:   date,
				Value: float32(callSuccess),
			})
			callFailureLine = append(callFailureLine, StatisticChartLineItem{
				Key:   date,
				Value: float32(stat.CallFailure),
			})
			totalTokensLine = append(totalTokensLine, StatisticChartLineItem{
				Key:   date,
				Value: float32(stat.TotalTokens),
			})
			completionTokensLine = append(completionTokensLine, StatisticChartLineItem{
				Key:   date,
				Value: float32(stat.CompletionTokens),
			})
			promptTokensLine = append(promptTokensLine, StatisticChartLineItem{
				Key:   date,
				Value: float32(stat.PromptTokens),
			})
		} else {
			callCountTotalLine = append(callCountTotalLine, StatisticChartLineItem{
				Key:   date,
				Value: 0,
			})
			callSuccessLine = append(callSuccessLine, StatisticChartLineItem{
				Key:   date,
				Value: 0,
			})
			callFailureLine = append(callFailureLine, StatisticChartLineItem{
				Key:   date,
				Value: 0,
			})
			totalTokensLine = append(totalTokensLine, StatisticChartLineItem{
				Key:   date,
				Value: 0,
			})
			completionTokensLine = append(completionTokensLine, StatisticChartLineItem{
				Key:   date,
				Value: 0,
			})
			promptTokensLine = append(promptTokensLine, StatisticChartLineItem{
				Key:   date,
				Value: 0,
			})
		}
	}

	return &ModelStatisticTrend{
		ModelCalls: StatisticChart{
			Name: "app_statistic_model_calls",
			Lines: []StatisticChartLine{
				{
					Name:  "app_statistic_call_count_total",
					Items: callCountTotalLine,
				},
				{
					Name:  "app_statistic_call_success",
					Items: callSuccessLine,
				},
				{
					Name:  "app_statistic_call_failure",
					Items: callFailureLine,
				},
			},
		},
		TokensUsage: StatisticChart{
			Name: "app_statistic_tokens_usage",
			Lines: []StatisticChartLine{
				{
					Name:  "app_statistic_total_tokens",
					Items: totalTokensLine,
				},
				{
					Name:  "app_statistic_completion_tokens",
					Items: completionTokensLine,
				},
				{
					Name:  "app_statistic_prompt_tokens",
					Items: promptTokensLine,
				},
			},
		},
	}, nil
}

// calculatePoP 计算环比
func calculatePoP(current, previous float32) float32 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100 // 避免除以零的错误
	}
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", ((current-previous)/previous)*100), 32)
	return float32(value)
}
