// Package converter 提供沙箱事件转换器。
package converter

import (
	"github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/schema"
)

// EventConverter 沙箱事件转换器接口，将 runner 输出转换为 eino Message。
type EventConverter interface {
	Convert(line string) (*schema.Message, error)
}

// NewEventConverter 根据运行器类型创建事件转换器。
func NewEventConverter(runnerType wga_sandbox_option.RunnerType) EventConverter {
	switch runnerType {
	default:
		return newOpencodeConverter()
	}
}
