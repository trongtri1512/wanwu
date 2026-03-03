package sandbox

import (
	"context"
)

// 一次性容器模式沙箱实现（未完成）。
//
// 一次性模式为每次执行创建新的容器，执行完成后销毁容器。
//
// 优点：
//   - 完全隔离
//   - 安全性高
//
// 适用场景：
//   - 生产环境
//   - 多租户场景
//
// TODO: 实现以下功能：
//   - Prepare: 创建新容器
//   - Cleanup: 销毁容器
//   - Execute/ExecuteSync: 在容器内执行命令
//   - CopyToSandbox/CopyFromSandbox: 文件复制

var _ Sandbox = (*oneshotSandbox)(nil)

type oneshotSandbox struct {
	imageName   string
	imageTag    string
	apiEndpoint string
	uuid        string
}

func newOneshotSandbox(imageName, apiEndpoint, uuid string) Sandbox {
	imageName, imageTag := parseImageName(imageName)
	return &oneshotSandbox{
		imageName:   imageName,
		imageTag:    imageTag,
		apiEndpoint: apiEndpoint,
		uuid:        uuid,
	}
}

func (s *oneshotSandbox) Prepare(ctx context.Context) error {
	panic("not implemented: oneshotSandbox.Prepare")
}

func (s *oneshotSandbox) Cleanup(ctx context.Context) error {
	panic("not implemented: oneshotSandbox.Cleanup")
}

func (s *oneshotSandbox) Execute(ctx context.Context, args ...string) (<-chan string, error) {
	panic("not implemented: oneshotSandbox.Execute")
}

func (s *oneshotSandbox) ExecuteSync(ctx context.Context, args ...string) (string, error) {
	panic("not implemented: oneshotSandbox.ExecuteSync")
}

func (s *oneshotSandbox) CopyToSandbox(ctx context.Context, localPath string, destPath ...string) error {
	panic("not implemented: oneshotSandbox.CopyToSandbox")
}

func (s *oneshotSandbox) CopyFromSandbox(ctx context.Context, localPath string) error {
	panic("not implemented: oneshotSandbox.CopyFromSandbox")
}

func (s *oneshotSandbox) WorkDir() string {
	panic("not implemented: oneshotSandbox.WorkDir")
}

func (s *oneshotSandbox) UUID() string {
	panic("not implemented: oneshotSandbox.UUID")
}

func parseImageName(name string) (imageName, imageTag string) {
	parts := splitImageName(name)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return name, "latest"
}

func splitImageName(name string) []string {
	var result []string
	var current string
	for _, c := range name {
		if c == ':' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}
