package model

type ModelRecord struct {
	ID        uint32 `gorm:"primary_key"`
	CreatedAt int64  `gorm:"autoCreateTime:milli;index:idx_model_record_created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	// 组织ID
	OrgID string `gorm:"index:idx_model_record_org_id"`
	// 用户ID
	UserID string `gorm:"index:idx_model_record_user_id"`
	// 模型 ID
	ModelID string `gorm:"index:idx_model_record_model_id"`
	// 模型名称
	Model string `gorm:"index:idx_model_record_model"`
	// 模型类型（大语言、视觉、ocr、gui）
	ModelType string `gorm:"index:idx_model_record_model_type"`
	// 模型供应商
	Provider string `gorm:"index:idx_model_record_provider"`
	// prompt tokens
	PromptTokens int64 `gorm:"index:idx_model_record_prompt_tokens"`
	// completion tokens
	CompletionTokens int64 `gorm:"index:idx_model_record_completion_tokens"`
	// total tokens
	TotalTokens int64 `gorm:"index:idx_model_record_total_tokens"`
	// 首token时延 (流式ms)
	FirstTokenLatency int32 `gorm:"index:idx_model_record_first_token_latency"`
	// 耗时 (非流式ms)
	Costs int32 `gorm:"index:idx_model_record_costs"`
	// call count
	CallCount int32 `gorm:"index:idx_model_record_call_count"`
	// 流式调用次数
	StreamCount int32 `gorm:"index:idx_model_record_stream_count"`
	// 非流式调用次数
	NonStreamCount int32 `gorm:"index:idx_model_record_non_stream_count"`
	// call failure
	CallFailure int32 `gorm:"index:idx_model_record_call_failure"`
	// 流式调用次数失败
	StreamFailure int32 `gorm:"index:idx_model_record_stream_failure"`
	// 非流式调用次数失败
	NonStreamFailure int32 `gorm:"index:idx_model_record_non_stream_failure"`
	// 日期
	Date string `gorm:"index:idx_model_record_date"`
}
