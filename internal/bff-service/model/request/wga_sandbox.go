package request

type WgaSandboxRunReq struct {
	ThreadID       string              `json:"threadId"`
	RunID          string              `json:"runId"`
	Model          AppModelConfig      `json:"model" validate:"required"`
	CurrentTask    string              `json:"currentTask" validate:"required"`
	Instruction    string              `json:"instruction"`
	OverallTask    string              `json:"overallTask"`
	Messages       []WgaSandboxMessage `json:"messages"`
	Tools          []WgaSandboxTool    `json:"tools"`
	Skills         []WgaSandboxSkill   `json:"skills"`
	InputDir       string              `json:"inputDir"`
	OutputDir      string              `json:"outputDir"`
	EnableThinking bool                `json:"enableThinking"`
	SkipCleanup    bool                `json:"skipCleanup"`
	AgentName      string              `json:"agentName"`
}

type WgaSandboxMessage struct {
	Role    string `json:"role" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type WgaSandboxTool struct {
	// TODO: 待定义
}

type WgaSandboxSkill struct {
	// TODO: 待定义
}

func (r *WgaSandboxRunReq) Check() error {
	return nil
}

type WgaSandboxCleanupReq struct {
	RunID string `json:"runId" validate:"required"`
}

func (r *WgaSandboxCleanupReq) Check() error {
	return nil
}
