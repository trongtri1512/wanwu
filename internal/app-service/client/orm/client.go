package orm

import (
	"errors"
	"fmt"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"gorm.io/gorm"
)

const (
	// v0.3.2 增加app版本管理，清除之前发布的记录
	initFlagKey = "v0.3.2_app_table_cleared"
)

type Metadata struct {
	MetaKey   string `gorm:"primaryKey;column:key"`
	MetaValue string `gorm:"column:value"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
}

type Client struct {
	db *gorm.DB
}

type ApiKey struct {
	ApiId     string `json:"apiId"`
	CreatedAt int64  `json:"createdAt"`
	ApiKey    string `json:"apiKey"`
}

type ExplorationAppInfo struct {
	AppId       string
	AppType     string
	CreatedAt   int64
	UpdatedAt   int64
	IsFavorite  bool
	PublishType string
	UserID      string
}

type SensitiveWordTableWithWord struct {
	model.SensitiveWordTable
	SensitiveWords []string
}

type ModelStatistic struct {
	Overview ModelStatisticOverview `json:"overview"`
	Trend    ModelStatisticTrend    `json:"trend"`
}

// ModelStatisticOverviewItem 模型统计概览项
type ModelStatisticOverviewItem struct {
	Value            float32 `json:"value"`
	PeriodOverperiod float32 `json:"periodOverperiod"`
}

type ModelStatisticList struct {
	Items []ModelStatisticItem
	Total int32
}

type ModelStatisticItem struct {
	ModelId              string
	Model                string
	Provider             string
	OrgId                string
	CallCount            int32
	CallFailure          int32
	FailureRate          float32
	PromptTokens         int64
	CompletionTokens     int64
	TotalTokens          int64
	AvgCosts             float32
	AvgFirstTokenLatency float32
}

type ModelStatisticOverview struct {
	CallCount            ModelStatisticOverviewItem `json:"callCount"`
	CallFailure          ModelStatisticOverviewItem `json:"callFailure"`
	TotalTokens          ModelStatisticOverviewItem `json:"totalTokens"`
	CompletionTokens     ModelStatisticOverviewItem `json:"completionTokens"`
	PromptTokens         ModelStatisticOverviewItem `json:"promptTokens"`
	AvgCosts             ModelStatisticOverviewItem `json:"avgCosts"`
	AvgFirstTokenLatency ModelStatisticOverviewItem `json:"avgFirstTokenLatency"`
}

type ModelStatisticTrend struct {
	ModelCalls  StatisticChart `json:"modelCalls"`
	TokensUsage StatisticChart `json:"tokensUsage"`
}

type StatisticChart struct {
	Name  string               `json:"name"`
	Lines []StatisticChartLine `json:"lines"`
}

type StatisticChartLine struct {
	Name  string                   `json:"name"`
	Items []StatisticChartLineItem `json:"items"`
}

type StatisticChartLineItem struct {
	Key   string  `json:"key"`
	Value float32 `json:"value"`
}

func NewClient(db *gorm.DB) (*Client, error) {
	// 先迁移 Metadata 表（用于记录状态）
	if err := db.AutoMigrate(&Metadata{}); err != nil {
		return nil, err
	}
	// 自动迁移表结构
	if err := db.AutoMigrate(
		model.OpenApiKey{},
		model.ApiKey{}, // appKey
		model.App{},
		model.AppHistory{},
		model.AppFavorite{},
		model.AppUrl{},
		model.AppConversation{},
		model.SensitiveWordTable{},
		model.SensitiveWordVocabulary{},
		model.ChatflowApplcation{},
		model.ModelRecord{},
	); err != nil {
		return nil, err
	}

	var meta Metadata
	err := db.Where(&Metadata{MetaKey: initFlagKey}).First(&meta).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 首次运行：清表
			if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.App{}).Error; err != nil {
				return nil, fmt.Errorf("failed to clear app table: %w", err)
			}
			// 写入带版本的初始化标记
			if err := db.Create(&Metadata{
				MetaKey: initFlagKey,
			}).Error; err != nil {
				return nil, fmt.Errorf("failed to set init flag: %w", err)
			}
		} else {
			return nil, fmt.Errorf("query metadata failed: %w", err)
		}
	}
	return &Client{
		db: db,
	}, nil
}

func toErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
}

func canAccessApp(info *model.App, userId, orgId string) bool {
	switch info.PublishType {
	case constant.AppPublishPublic:
		return true
	case constant.AppPublishOrganization:
		return info.OrgID == orgId
	case constant.AppPublishPrivate:
		return info.UserID == userId
	default:
		return false
	}
}
