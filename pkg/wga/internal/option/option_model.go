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
	if options.Model.EndpointUrl == "" {
		return fmt.Errorf("model endpoint url empty")
	}
	return nil
}

// ToChatModel 创建聊天模型实例。
func (options *Options) ToChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	if err := options.checkModel(); err != nil {
		return nil, err
	}
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:            options.Model.Model,
		APIKey:           options.Model.ApiKey,
		BaseURL:          options.Model.EndpointUrl,
		Temperature:      util.IfElse(options.Model.Params.TemperatureEnable, &options.Model.Params.Temperature, nil),
		TopP:             util.IfElse(options.Model.Params.TopPEnable, &options.Model.Params.TopP, nil),
		FrequencyPenalty: util.IfElse(options.Model.Params.FrequencyPenaltyEnable, &options.Model.Params.FrequencyPenalty, nil),
		PresencePenalty:  util.IfElse(options.Model.Params.PresencePenaltyEnable, &options.Model.Params.PresencePenalty, nil),
	})
}
