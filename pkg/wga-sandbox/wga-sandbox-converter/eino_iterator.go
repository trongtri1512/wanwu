package wga_sandbox_converter

import (
	"context"

	"github.com/UnicomAI/wanwu/pkg/util"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func ConvertToEinoIterator(
	ctx context.Context,
	runnerType wga_sandbox_option.RunnerType,
	outputCh <-chan string,
) *adk.AsyncIterator[*adk.AgentEvent] {
	iterator, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	conv := NewEinoConverter(runnerType)

	go func() {
		defer util.PrintPanicStack()
		defer generator.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-outputCh:
				if !ok {
					return
				}
				msg, err := conv.Convert(line)
				if err != nil {
					generator.Send(&adk.AgentEvent{Err: err})
					continue
				}
				if msg == nil {
					continue
				}
				generator.Send(&adk.AgentEvent{
					Output: &adk.AgentOutput{
						MessageOutput: &adk.MessageVariant{
							Message: msg,
							Role:    schema.Assistant,
						},
					},
				})
			}
		}
	}()

	return iterator
}

func ConvertToEinoIteratorWithError(
	ctx context.Context,
	runnerType wga_sandbox_option.RunnerType,
	err error,
) *adk.AsyncIterator[*adk.AgentEvent] {
	iterator, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer util.PrintPanicStack()
		defer generator.Close()
		generator.Send(&adk.AgentEvent{Err: err})
	}()

	return iterator
}
