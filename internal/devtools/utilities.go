package devtools

import (
	"context"
	"fmt"

	"devbox/internal/readcache"
)

// DetectJQ returns cached or newly detected jq executables.
func (s *Service) DetectJQ(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return detectUtility(ctx, &s.jq, refresh, "jq", "jq-")
}

// DetectRG returns cached or newly detected ripgrep executables.
func (s *Service) DetectRG(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return detectUtility(ctx, &s.rg, refresh, "rg", "ripgrep ")
}

// DetectGH returns cached or newly detected GitHub CLI executables.
func (s *Service) DetectGH(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return detectUtility(ctx, &s.gh, refresh, "gh", "gh version ", "GH_NO_UPDATE_NOTIFIER=1")
}

func detectUtility(ctx context.Context, cache *readcache.Cache[[]ToolDetection], refresh bool, name, prefix string, environment ...string) ([]ToolDetection, error) {
	return cache.Read(ctx, refresh, 0, func(ctx context.Context) ([]ToolDetection, error) {
		return detectTools(ctx, name, fmt.Sprintf("在已检查的路径中没有找到 %s。", name),
			func(ctx context.Context, path string) ToolDetection {
				return detectVersion(ctx, path, prefix, environment...)
			}, macOSBinPaths(name)...)
	})
}
