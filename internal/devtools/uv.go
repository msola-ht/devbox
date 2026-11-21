package devtools

import "context"

// DetectUV returns cached or newly detected uv executables.
func (s *Service) DetectUV(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.uv.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return detectTools(ctx, "uv", "应用的 PATH 中没有找到 uv。",
			func(ctx context.Context, path string) ToolDetection {
				return detectVersion(ctx, path, "uv ")
			})
	})
}
