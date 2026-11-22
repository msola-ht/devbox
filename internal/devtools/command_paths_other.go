//go:build !windows

package devtools

import (
	"path/filepath"
	"runtime"
)

func platformCommandPaths(name string) []string {
	if runtime.GOOS != "linux" {
		return nil
	}
	return []string{filepath.Join("/usr/bin", name)}
}
