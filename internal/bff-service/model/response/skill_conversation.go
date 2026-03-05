package response

type CreateSkillConversationResp struct {
	ConversationId string `json:"conversationId"` // 会话ID
}

type SkillConversationItem struct {
	ConversationId string `json:"conversationId"` // 会话ID
	Title          string `json:"title"`          // 会话标题
	CreatedAt      string `json:"createdAt"`      // 创建时间
}

type SkillConversationDetailInfo struct {
	ConversationDetailInfo
	ResponseFiles []AssistantResponseFile `json:"responseFiles"`
}

type AssistantResponseFile struct {
	FileName string `json:"name"`
	FileSize int64  `json:"size"`
	FileUrl  string `json:"fileUrl"`
	MIMEType string `json:"mimeType"`

	// 扩展信息：
	// skill => {"name":"技能名称", "desc":"技能描述", "author":"技能作者", "avatar":{"path":"技能图标"}, "inResource": bool, "expiredAt": "过期时间7天", "skillSaveId": "保存的技能ID"}
	MetaData map[string]interface{} `json:"metadata"`
}

type SkillConversationChatResp struct {
	Code           int                    `json:"code"`
	Message        string                 `json:"message"`
	Response       string                 `json:"response"`
	Order          int                    `json:"order"`
	EventType      int                    `json:"eventType"`
	EventData      interface{}            `json:"eventData"`
	GenFileUrlList []interface{}          `json:"gen_file_url_list"`
	History        []interface{}          `json:"history"`
	Finish         int                    `json:"finish"`
	Usage          SkillConversationUsage `json:"usage"`
	SearchList     []interface{}          `json:"search_list"`
	QaType         int                    `json:"qa_type"`
	SkillSaveId    string                 `json:"skillSaveId"`
}

type SkillConversationUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
