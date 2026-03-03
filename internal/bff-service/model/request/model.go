package request

type ImportOrUpdateModelRequest struct {
	ModelConfig
}

func (o *ImportOrUpdateModelRequest) Check() error {
	return o.ModelConfig.Check()
}

type DeleteModelRequest struct {
	BaseModelRequest
}

func (o *DeleteModelRequest) Check() error {
	return nil
}

type GetModelRequest struct {
	BaseModelRequest
}

func (o *GetModelRequest) Check() error {
	return nil
}

type ListModelsRequest struct {
	ModelType   string `json:"modelType" form:"modelType" `    // 模型类型
	Provider    string `json:"provider" form:"provider"`       // 模型供应商
	DisplayName string `json:"displayName" form:"displayName"` // 模型显示名称
	IsActive    bool   `json:"isActive" form:"isActive"`       // 启用状态（true: 启用）
	FilterScope string `json:"filterScope" form:"filterScope"` // 模型作用域类型(public: 公有模型，private: 我的模型)
	ScopeType   string `json:"scopeType" form:"scopeType"`     // 模型公开范围(1-私有 2-公开 3-组织)
}

func (o *ListModelsRequest) Check() error {
	return nil
}

type ListTypeModelsRequest struct {
	ModelType string `json:"modelType" form:"modelType" ` // 模型类型
}

func (o *ListTypeModelsRequest) Check() error {
	return nil
}

type ModelStatusRequest struct {
	BaseModelRequest
	IsActive bool `json:"isActive"` // 启用状态（true: 启用，false: 禁用）
}

func (o *ModelStatusRequest) Check() error {
	return nil
}
