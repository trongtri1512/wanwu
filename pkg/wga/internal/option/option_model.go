package option

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func (options *Options) checkModel() error {
	if options.Model.Model == "" {
		return fmt.Errorf("model required")
	}
	if options.Model.BaseURL == "" {
		return fmt.Errorf("model base url empty")
	}
	return nil
}

// ToChatModel 创建聊天模型实例。
func (options *Options) ToChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	if err := options.checkModel(); err != nil {
		return nil, err
	}
	cfg := &openai.ChatModelConfig{
		Model:   options.Model.Model,
		APIKey:  options.Model.APIKey,
		BaseURL: options.Model.BaseURL,
	}
	if options.Model.Params != nil {
		cfg.Temperature = util.IfElse(options.Model.Params.TemperatureEnable, &options.Model.Params.Temperature, nil)
		cfg.TopP = util.IfElse(options.Model.Params.TopPEnable, &options.Model.Params.TopP, nil)
		cfg.FrequencyPenalty = util.IfElse(options.Model.Params.FrequencyPenaltyEnable, &options.Model.Params.FrequencyPenalty, nil)
		cfg.PresencePenalty = util.IfElse(options.Model.Params.PresencePenaltyEnable, &options.Model.Params.PresencePenalty, nil)
	}
	return openai.NewChatModel(ctx, cfg)
}
