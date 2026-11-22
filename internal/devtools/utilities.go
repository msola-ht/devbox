package devtools

import "context"

// DetectJQ returns cached or newly detected jq executables.
func (s *Service) DetectJQ(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.jq.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return s.detectVersionedTool(ctx, toolJQ, "jq", "jq-", nil)
	})
}

// DetectRG returns cached or newly detected ripgrep executables.
func (s *Service) DetectRG(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.rg.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return s.detectVersionedTool(ctx, toolRG, "rg", "ripgrep ", nil)
	})
}

// DetectGH returns cached or newly detected GitHub CLI executables.
func (s *Service) DetectGH(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.gh.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return s.detectVersionedTool(ctx, toolGH, "gh", "gh version ", nil, "GH_NO_UPDATE_NOTIFIER=1")
	})
}
