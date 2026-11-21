//go:build !darwin

package devtools

import "context"

func (s *Service) manageProjectUV(context.Context, string, *uvInstallJob) (UVInstallStatus, string) {
	return uvInstallFailed, "当前仅在 macOS 上支持管理项目 uv。"
}
