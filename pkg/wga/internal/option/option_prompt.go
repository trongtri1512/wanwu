package option

import (
	"context"
	"time"

	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// FormatInstruction 格式化系统提示词。
// 支持 Jinja2 模板语法，注入 agent_name、agent_desc、current_time 变量。
func (options *Options) FormatInstruction(ctx context.Context, cfg *config.Agent) (string, error) {
	// prompt variables
	vs := map[string]any{
		"agent_name":   cfg.Name,
		"agent_desc":   cfg.Description,
		"current_time": time.Now().Format(time.RFC1123),
	}
	// instruction
	rets, err := prompt.FromMessages(schema.Jinja2, schema.SystemMessage(cfg.Prompt)).Format(ctx, vs)
	if err != nil {
		return "", err
	}
	for _, ret := range rets {
		return ret.Content, nil
	}
	return "", nil
}
