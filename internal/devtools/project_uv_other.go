//go:build !darwin

package devtools

import "context"

func (s *Service) installProjectUV(context.Context, string, *uvInstallJob) (UVInstallStatus, string) {
	return uvInstallFailed, "当前仅在 macOS 上支持安装项目 uv。"
}
