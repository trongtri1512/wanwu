package wga_sandbox_converter

import (
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/schema"
)

type EinoConverter interface {
	Convert(line string) (*schema.Message, error)
}

func NewEinoConverter(runnerType wga_sandbox_option.RunnerType) EinoConverter {
	switch runnerType {
	default:
		return newOpencodeConverter()
	}
}
