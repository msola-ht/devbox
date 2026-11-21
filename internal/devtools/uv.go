package devtools

import "context"

// DetectUV returns cached or newly detected uv executables.
func (s *Service) DetectUV(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.uv.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		paths := macOSBinPaths("uv")
		if path := s.uvInstall.binPath(); path != "" {
			paths = append(paths, path)
		}
		return detectTools(ctx, "uv", "在已检查的路径中没有找到 uv。",
			func(ctx context.Context, path string) ToolDetection {
				return detectVersion(ctx, path, "uv ")
			}, paths...)
	})
}
