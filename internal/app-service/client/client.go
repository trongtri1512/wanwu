package client

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
)

type IClient interface {
	// ---api key ---
	CreateApiKey(ctx context.Context, userId, orgId, name, desc string, expiredAt int64, apiKey string) (*model.OpenApiKey, *err_code.Status)
	DeleteApiKey(ctx context.Context, keyId uint32) *err_code.Status
	UpdateApiKey(ctx context.Context, keyId uint32, userId, orgId, name, desc string, expiredAt int64) *err_code.Status
	ListApiKeys(ctx context.Context, userId, orgId string, offset, limit int32) ([]*model.OpenApiKey, int64, *err_code.Status)
	UpdateApiKeyStatus(ctx context.Context, keyId uint32, status bool) *err_code.Status
	GetApiKeyByKey(ctx context.Context, key string) (*model.OpenApiKey, *err_code.Status)

	// --- app key ---
	GetAppKeyList(ctx context.Context, userId, orgId, appId, appType string) ([]*model.ApiKey, *err_code.Status)
	DelAppKey(ctx context.Context, appKeyId uint32) *err_code.Status
	GenAppKey(ctx context.Context, userId, orgId, appId, appType, appKey string) (*model.ApiKey, *err_code.Status)
	GetAppKeyByKey(ctx context.Context, appKey string) (*model.ApiKey, *err_code.Status)

	// --- explore ---
	GetExplorationAppList(ctx context.Context, userId, orgId, name, appType, searchType string) ([]*orm.ExplorationAppInfo, *err_code.Status)
	ChangeExplorationAppFavorite(ctx context.Context, userId, orgId, appId, appType string, isFavorite bool) *err_code.Status

	// --- app ---
	PublishApp(ctx context.Context, userId, orgId, appId, appType, publishType string) *err_code.Status
	UnPublishApp(ctx context.Context, appId, appType, userId string) *err_code.Status
	GetAppList(ctx context.Context, userId, orgId, appType string) ([]*model.App, *err_code.Status)
	DeleteApp(ctx context.Context, appId, appType string) *err_code.Status
	RecordAppHistory(ctx context.Context, userId, appId, appType string) *err_code.Status
	GetAppListByIds(ctx context.Context, ids []string) ([]*model.App, *err_code.Status)
	GetAppInfo(ctx context.Context, appId, appType string) (*model.App, *err_code.Status)
	ConvertAppType(ctx context.Context, appId, oldAppType, newAppType string) *err_code.Status

	// --- safety ---
	CreateSensitiveWordTable(ctx context.Context, userId, orgId, tableName, remark, tableType string) (string, *err_code.Status)
	UpdateSensitiveWordTable(ctx context.Context, tableId uint32, tableName, remark string) *err_code.Status
	UpdateSensitiveWordTableReply(ctx context.Context, tableId uint32, reply string) *err_code.Status
	DeleteSensitiveWordTable(ctx context.Context, tableId uint32) *err_code.Status
	GetSensitiveWordTableList(ctx context.Context, userId, orgId, tableType string) ([]*model.SensitiveWordTable, *err_code.Status)
	GetSensitiveVocabularyList(ctx context.Context, tableId uint32, offset, limit int32) ([]*model.SensitiveWordVocabulary, int64, *err_code.Status)
	UploadSensitiveVocabulary(ctx context.Context, userId, orgId, importType, word, sensitiveType, filePath string, tableId uint32) *err_code.Status
	DeleteSensitiveVocabulary(ctx context.Context, tableId, wordId uint32) *err_code.Status
	GetSensitiveWordTableListWithWordsByIDs(ctx context.Context, tableIds []string) ([]*orm.SensitiveWordTableWithWord, *err_code.Status)
	GetSensitiveWordTableListByIDs(ctx context.Context, tableIds []string) ([]*model.SensitiveWordTable, *err_code.Status)
	GetSensitiveWordTableByID(ctx context.Context, tableId uint32) (*model.SensitiveWordTable, *err_code.Status)
	GetGlobalSensitiveWordTableList(ctx context.Context) ([]*model.SensitiveWordTable, *err_code.Status)

	// --- web_url ---
	CreateAppUrl(ctx context.Context, appUrl *model.AppUrl) *err_code.Status
	DeleteAppUrl(ctx context.Context, urlID uint32) *err_code.Status
	UpdateAppUrl(ctx context.Context, appUrl *model.AppUrl) *err_code.Status
	GetAppUrlList(ctx context.Context, appID, appType string) ([]*model.AppUrl, *err_code.Status)
	GetAppUrlInfoBySuffix(ctx context.Context, suffix string) (*model.AppUrl, *err_code.Status)
	AppUrlStatusSwitch(ctx context.Context, urlID uint32, status bool) *err_code.Status

	// --- conversation ---
	GetConversationByID(ctx context.Context, ConversationId string) (*model.AppConversation, *err_code.Status)
	CreateConversation(ctx context.Context, userId, orgId, appId, appType, conversationId, conversationName string) *err_code.Status
	GetChatflowApplication(ctx context.Context, orgId, userId, workflowId string) (*model.ChatflowApplcation, *err_code.Status)
	GetChatflowApplicationByApplicationID(ctx context.Context, orgId, userId, applicationId string) (*model.ChatflowApplcation, *err_code.Status)
	CreateChatflowApplication(ctx context.Context, orgId, userId, workflowId, applicationId string) *err_code.Status

	// --- statistic ---
	GetModelStatistic(ctx context.Context, userId, orgId, startDate, endDate string, modelIds []string, modelType string) (*orm.ModelStatistic, *err_code.Status)
	GetModelStatisticList(ctx context.Context, userId, orgId, startDate, endDate string, modelIds []string, modelType string, offset, limit int32) (*orm.ModelStatisticList, *err_code.Status)
	RecordModelStatistic(ctx context.Context, userId, orgId, modelId, model, modelType string,
		promptTokens, completionTokens, totalTokens int64, firstTokenLatency, costs int32, isSuccess bool, isStream bool, provider string) *err_code.Status
}
