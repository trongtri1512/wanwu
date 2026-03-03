package sandbox

import (
	"context"
	"fmt"
	"sync"

	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
)

// Manager 管理沙箱实例的生命周期。
//
// 提供以下功能：
//   - Create: 创建并初始化沙箱
//   - Get: 获取沙箱实例
//   - Cleanup: 清理沙箱
//
// Manager 内部使用 map 按 runID 管理沙箱实例，
// 支持 SkipCleanup 场景下的延迟清理。
type Manager struct {
	mu        sync.RWMutex
	sandboxes map[string]Sandbox
}

// NewManager 创建沙箱管理器。
func NewManager() *Manager {
	return &Manager{
		sandboxes: make(map[string]Sandbox),
	}
}

// Create 创建并初始化沙箱。
//
// 根据 sandboxConfig 创建对应的沙箱实例（reuse 或 oneshot），
// 调用 Prepare 初始化工作目录，并注册到管理器中。
func (m *Manager) Create(ctx context.Context, runID string, cfg wga_sandbox_option.SandboxConfig) error {
	var sb Sandbox
	if cfg.Type() == wga_sandbox_option.SandboxTypeOneshot {
		sb = newOneshotSandbox(cfg.ImageName(), cfg.APIEndpoint(), runID)
	} else {
		sb = newReuseSandbox(cfg.APIEndpoint(), runID)
	}

	if err := sb.Prepare(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sandboxes[runID] = sb
	return nil
}

// Get 获取沙箱实例。
// 如果沙箱不存在，返回 error。
func (m *Manager) Get(runID string) (Sandbox, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sb, ok := m.sandboxes[runID]
	if !ok {
		return nil, fmt.Errorf("sandbox not found: %s", runID)
	}
	return sb, nil
}

// Cleanup 清理沙箱。
//
// 从管理器中移除沙箱实例，并调用 Sandbox.Cleanup 清理工作目录。
// 如果沙箱不存在，直接返回 nil。
func (m *Manager) Cleanup(ctx context.Context, runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sb, ok := m.sandboxes[runID]
	if !ok {
		return nil
	}
	delete(m.sandboxes, runID)
	return sb.Cleanup(ctx)
}
