// Package runner 提供智能体运行器接口。
package runner

import (
	"context"
)

// Runner 智能体运行器接口。
type Runner interface {
	BeforeRun(ctx context.Context) error
	Run(ctx context.Context) (<-chan string, error)
	AfterRun(ctx context.Context) error
}
