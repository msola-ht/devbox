//go:build !windows

package devtools

import "context"

func runJavaScriptVersion(ctx context.Context, path string, environment []string) ([]byte, error) {
	return runCommand(ctx, path, []string{"--version"}, environment...)
}
