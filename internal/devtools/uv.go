package devtools

import "context"

// DetectUV checks the uv executables found in the app's PATH.
func (Service) DetectUV(ctx context.Context) ([]ToolDetection, error) {
	return detectTools(ctx, "uv", "应用的 PATH 中没有找到 uv。",
		func(ctx context.Context, path string) ToolDetection {
			return detectVersion(ctx, path, "uv ")
		})
}
