// Package sandbox 提供沙箱环境接口。
package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/util"
)

// 复用容器模式沙箱实现。
//
// 复用模式使用预运行的容器，通过 HTTP API 进行交互，
// 每次执行在容器内创建独立的工作目录，执行完成后清理工作目录。
//
// 优点：
//   - 启动快（无需创建新容器）
//   - 资源占用少
//
// 适用场景：
//   - 开发环境
//   - 频繁执行的场景

var _ Sandbox = (*reuseSandbox)(nil)

type reuseSandbox struct {
	client  *client
	uuid    string
	workDir string
}

func newReuseSandbox(apiEndpoint, uuid string) Sandbox {
	return &reuseSandbox{
		client: newClient(apiEndpoint),
		uuid:   uuid,
	}
}

func (s *reuseSandbox) Prepare(ctx context.Context) error {
	s.workDir = filepath.Join(workspaceBase, s.uuid, "workspace")
	_, err := s.client.exec(ctx, "mkdir -p "+s.workDir, "/")
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	return nil
}

func (s *reuseSandbox) Cleanup(ctx context.Context) error {
	workspacePath := filepath.Join(workspaceBase, s.uuid)
	_, err := s.client.exec(ctx, "rm -rf "+workspacePath, "/")
	if err != nil {
		return fmt.Errorf("failed to cleanup workspace: %w", err)
	}
	return nil
}

func (s *reuseSandbox) Execute(ctx context.Context, args ...string) (<-chan string, error) {
	cmd := strings.Join(args, " ")
	outputCh := make(chan string, 1024)

	go func() {
		defer util.PrintPanicStack()
		defer close(outputCh)

		err := s.client.execWithOutput(ctx, cmd, s.workDir, outputCh)
		if err != nil {
			select {
			case outputCh <- fmt.Sprintf("[ERROR] command execution failed: %v", err):
			case <-ctx.Done():
			}
		}
	}()

	return outputCh, nil
}

func (s *reuseSandbox) ExecuteSync(ctx context.Context, args ...string) (string, error) {
	cmd := strings.Join(args, " ")
	result, err := s.client.exec(ctx, cmd, s.workDir)
	if err != nil {
		return "", err
	}
	return result.Output, nil
}

// CopyToSandbox 将本地文件或目录复制到沙箱。
// localPath: 本地文件或目录路径
// relativePath: 可选，相对于 workDir 的目标路径，为空则使用 workDir
func (s *reuseSandbox) CopyToSandbox(ctx context.Context, localPath string, relativePath ...string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local path: %w", err)
	}

	sandboxPath := s.workDir
	if len(relativePath) > 0 {
		sandboxPath = filepath.Join(s.workDir, relativePath[0])
	}

	if info.IsDir() {
		return s.copyDirToSandbox(ctx, localPath, sandboxPath)
	}
	return s.copyFileToSandbox(ctx, localPath, sandboxPath)
}

// CopyFromSandbox 将沙箱的 workDir 复制到本地。
// localPath: 本地目标路径
func (s *reuseSandbox) CopyFromSandbox(ctx context.Context, localPath string) error {
	info, err := s.getSandboxInfo(ctx, s.workDir)
	if err != nil {
		return fmt.Errorf("failed to get sandbox info: %w", err)
	}

	if info.isDir {
		return s.copyDirFromSandbox(ctx, localPath)
	}
	return s.copyFileFromSandbox(ctx, s.workDir, localPath)
}

func (s *reuseSandbox) WorkDir() string {
	return s.workDir
}

func (s *reuseSandbox) UUID() string {
	return s.uuid
}

// copyDirToSandbox 将本地目录复制到沙箱。
// 如果目录为空，跳过复制。
func (s *reuseSandbox) copyDirToSandbox(ctx context.Context, localDir, sandboxPath string) error {
	if isDirEmpty(localDir) {
		return nil
	}

	tarData, err := util.TarDir(localDir)
	if err != nil {
		return fmt.Errorf("failed to tar directory: %w", err)
	}

	tarName := filepath.Base(localDir) + ".tar"
	tarPath := filepath.Join(s.workDir, tarName)

	if err := s.client.uploadData(ctx, tarData, tarPath); err != nil {
		return fmt.Errorf("failed to upload tar: %w", err)
	}
	defer func() { _ = s.client.delete(ctx, tarPath) }()

	if _, err := s.client.exec(ctx, "mkdir -p "+sandboxPath, "/"); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	cmd := fmt.Sprintf("cd %s && tar -xzf %s && rm %s", sandboxPath, tarPath, tarPath)
	if _, err := s.client.exec(ctx, cmd, s.workDir); err != nil {
		return fmt.Errorf("failed to extract tar: %w", err)
	}

	return nil
}

func (s *reuseSandbox) copyFileToSandbox(ctx context.Context, localFile, sandboxPath string) error {
	return s.client.upload(ctx, localFile, sandboxPath)
}

// copyDirFromSandbox 将沙箱的 workDir 复制到本地目录。
// 如果目录为空或只包含隐藏文件，跳过复制。
func (s *reuseSandbox) copyDirFromSandbox(ctx context.Context, localPath string) error {
	onlyHidden, err := s.hasOnlyHiddenFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to check sandbox dir: %w", err)
	}
	if onlyHidden {
		return nil
	}

	tarName := filepath.Base(s.workDir) + ".tar"
	tarPath := filepath.Join(filepath.Dir(s.workDir), tarName)

	cmd := fmt.Sprintf("cd %s && tar -czf %s %s", filepath.Dir(s.workDir), tarName, filepath.Base(s.workDir))
	if _, err := s.client.exec(ctx, cmd, "/"); err != nil {
		return fmt.Errorf("failed to create tar: %w", err)
	}
	defer func() { _ = s.client.delete(ctx, tarPath) }()

	tarData, err := s.client.downloadData(ctx, tarPath)
	if err != nil {
		return fmt.Errorf("failed to download tar: %w", err)
	}

	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	if err := util.Untar(tarData, localPath); err != nil {
		return fmt.Errorf("failed to extract tar: %w", err)
	}

	return nil
}

func (s *reuseSandbox) copyFileFromSandbox(ctx context.Context, sandboxPath, localPath string) error {
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}
	return s.client.download(ctx, sandboxPath, localPath)
}

type sandboxInfo struct {
	isDir bool
	name  string
}

func (s *reuseSandbox) getSandboxInfo(ctx context.Context, sandboxPath string) (*sandboxInfo, error) {
	cmd := fmt.Sprintf("if [ -d %s ]; then echo DIR; elif [ -f %s ]; then echo FILE; else echo NOTFOUND; fi", sandboxPath, sandboxPath)
	result, err := s.client.exec(ctx, cmd, "/")
	if err != nil {
		return nil, err
	}
	output := strings.TrimSpace(result.Output)
	if output == "NOTFOUND" {
		return nil, fmt.Errorf("sandbox path not found: %s", sandboxPath)
	}
	return &sandboxInfo{
		isDir: output == "DIR",
		name:  filepath.Base(sandboxPath),
	}, nil
}

// hasOnlyHiddenFiles 检查沙箱目录是否为空或只包含隐藏文件/目录。
func (s *reuseSandbox) hasOnlyHiddenFiles(ctx context.Context) (bool, error) {
	cmd := fmt.Sprintf("ls -A %s | grep -v '^[.]' || true", s.workDir)
	result, err := s.client.exec(ctx, cmd, "/")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(result.Output) == "", nil
}

// isDirEmpty 检查本地目录是否为空。
func isDirEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return true
	}
	return len(entries) == 0
}
