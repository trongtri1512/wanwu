// Package sandbox 提供沙箱环境接口。
package sandbox

import "context"

const workspaceBase = "/home/root/workspace"

// Sandbox 沙箱环境接口。
type Sandbox interface {
	// Prepare 准备沙箱环境，创建 workDir。
	Prepare(ctx context.Context) error
	// Cleanup 清理沙箱环境，删除 workDir。
	Cleanup(ctx context.Context) error
	// Execute 异步执行命令，返回输出通道。
	Execute(ctx context.Context, args ...string) (<-chan string, error)
	// ExecuteSync 同步执行命令，返回完整输出。
	ExecuteSync(ctx context.Context, args ...string) (string, error)
	// CopyToSandbox 将本地文件或目录复制到沙箱。
	// localPath: 本地文件或目录路径
	// relativePath: 可选，相对于 workDir 的目标路径，为空则使用 workDir
	CopyToSandbox(ctx context.Context, localPath string, relativePath ...string) error
	// CopyFromSandbox 将沙箱的 workDir 复制到本地。
	// localPath: 本地目标路径
	CopyFromSandbox(ctx context.Context, localPath string) error
	// WorkDir 返回沙箱工作目录的绝对路径。
	WorkDir() string
	// UUID 返回沙箱唯一标识。
	UUID() string
}
