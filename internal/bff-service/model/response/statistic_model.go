package response

type ModelStatistic struct {
	Overview ModelStatisticOverview `json:"overview"`
	Trend    ModelStatisticTrend    `json:"trend"`
}

type ModelStatisticOverview struct {
	CallCountTotal        StatisticOverviewItem `json:"callCountTotal"`
	CallFailureTotal      StatisticOverviewItem `json:"callFailureTotal"`
	TotalTokensTotal      StatisticOverviewItem `json:"totalTokensTotal"`
	PromptTokensTotal     StatisticOverviewItem `json:"promptTokensTotal"`
	CompletionTokensTotal StatisticOverviewItem `json:"completionTokensTotal"`
	AvgCosts              StatisticOverviewItem `json:"avgCosts"`
	AvgFirstTokenLatency  StatisticOverviewItem `json:"avgFirstTokenLatency"`
}

type ModelStatisticTrend struct {
	ModelCalls  StatisticChart `json:"modelCalls"`
	TokensUsage StatisticChart `json:"tokensUsage"`
}

type ModelStatisticItem struct {
	ModelId              string  `json:"modelId"`
	Model                string  `json:"model"`
	Provider             string  `json:"provider"`
	OrgName              string  `json:"orgName"`
	CallCount            int32   `json:"callCount"`
	CallFailure          int32   `json:"callFailure"`
	FailureRate          float32 `json:"failureRate"`
	PromptTokens         int64   `json:"promptTokens"`
	CompletionTokens     int64   `json:"completionTokens"`
	TotalTokens          int64   `json:"totalTokens"`
	AvgCosts             float32 `json:"avgCosts"`
	AvgFirstTokenLatency float32 `json:"avgFirstTokenLatency"`
}
