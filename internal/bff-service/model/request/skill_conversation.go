package request

type CreateSkillConversationReq struct {
	Title string `json:"title" validate:"required"` // 会话标题
}

type DeleteSkillConversationReq struct {
	ConversationId string `json:"conversationId" validate:"required"` // 会话ID
}

type GetSkillConversationListReq struct {
	PageNo   int `json:"pageNo" form:"pageNo" validate:"required"`     // 页码
	PageSize int `json:"pageSize" form:"pageSize" validate:"required"` // 每页数量
}

type GetSkillConversationDetailReq struct {
	ConversationId string `json:"conversationId" form:"conversationId" validate:"required"` // 会话ID
}

type SkillConversationChatReq struct {
	ConversationId string                 `json:"conversationId" validate:"required"` // 会话ID
	Query          string                 `json:"query" validate:"required"`          // 用户提问
	FileInfo       []ConversionStreamFile `json:"fileInfo"`                           // 上传的文件信息
	ModelConfig    *AppModelConfig        `json:"modelConfig"`                        // 模型
}

type SkillConversationSaveReq struct {
	ConversationId string `json:"conversationId" validate:"required"` // 会话ID
	SkillSaveId    string `json:"skillSaveId" validate:"required"`    // 技能ID
}

func (c *CreateSkillConversationReq) Check() error {
	return nil
}

func (c *DeleteSkillConversationReq) Check() error {
	return nil
}

func (c *GetSkillConversationListReq) Check() error {
	return nil
}

func (c *GetSkillConversationDetailReq) Check() error {
	return nil
}

func (c *SkillConversationChatReq) Check() error {
	return nil
}

func (c *SkillConversationSaveReq) Check() error {
	return nil
}
