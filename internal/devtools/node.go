package devtools

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DetectNode returns cached or newly detected Node.js installations.
func (s *Service) DetectNode(ctx context.Context, refresh bool) ([]ToolDetection, error) {
	return s.node.Read(ctx, refresh, 0, detectNode)
}

func detectNode(ctx context.Context) ([]ToolDetection, error) {
	return detectNodeTool(ctx, "node", "在已检查的路径中没有找到 Node.js。",
		func(ctx context.Context, path string) ToolDetection {
			output, err := runJavaScriptVersion(ctx, path, nil)
			return javascriptVersionDetection(path, output, err)
		})
}

func detectNodeTool(ctx context.Context, name, notFoundMessage string, probe func(context.Context, string) ToolDetection) ([]ToolDetection, error) {
	var stores []nodeVersionStore
	discover := func(ctx context.Context) []toolCandidate {
		stores = nodeVersionStores(ctx)
		paths := append(macOSBinPaths(name), nodeInstallations(ctx, stores, name)...)
		return findToolCandidates(ctx, name, paths)
	}
	return detectCandidates(ctx, notFoundMessage, discover, func(ctx context.Context, path string) ToolDetection {
		result := probe(ctx, path)
		if result.Status == statusAvailable && result.Source == sourceUnchecked {
			result.Source = nodeStoreSource(result.resolvedPath, stores)
			if name == "node" && result.Source == sourceUnchecked {
				result.Source = sourceUnknown
			}
		}
		return result
	})
}

type nodeVersionStore struct {
	directory string
	binDir    string // Relative to a version directory.
	source    ToolSource
}

func nodeVersionStores(ctx context.Context) []nodeVersionStore {
	var stores []nodeVersionStore
	seen := make(map[string]bool)
	add := func(directory, binDir string, source ToolSource) {
		if !filepath.IsAbs(directory) || ctx.Err() != nil {
			return
		}
		dir, err := filepath.EvalSymlinks(directory)
		if err == nil && !seen[dir] {
			seen[dir] = true
			stores = append(stores, nodeVersionStore{directory: dir, binDir: binDir, source: source})
		}
	}

	home, _ := os.UserHomeDir()
	fnmRoots := []string{
		os.Getenv("FNM_DIR"),
		filepath.Join(os.Getenv("XDG_DATA_HOME"), "fnm"),
		filepath.Join(home, ".fnm"),
	}
	fnmBinDir := filepath.Join("installation", "bin")
	if runtime.GOOS == "windows" {
		fnmBinDir = "installation"
		fnmRoots = append(fnmRoots, filepath.Join(os.Getenv("APPDATA"), "fnm"), filepath.Join(os.Getenv("LOCALAPPDATA"), "fnm"))
		add(os.Getenv("NVM_HOME"), "", sourceNVMWindows)
		add(filepath.Join(os.Getenv("APPDATA"), "nvm"), "", sourceNVMWindows)
	} else {
		fnmRoots = append(fnmRoots, filepath.Join(home, ".local", "share", "fnm"))
		for _, root := range []string{os.Getenv("NVM_DIR"), filepath.Join(home, ".nvm"), filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "nvm")} {
			add(filepath.Join(root, "versions", "node"), "bin", sourceNVM)
		}
	}
	if runtime.GOOS == "darwin" {
		fnmRoots = append(fnmRoots, filepath.Join(home, "Library", "Application Support", "fnm"))
	}
	for _, root := range fnmRoots {
		add(filepath.Join(root, "node-versions"), fnmBinDir, sourceFNM)
	}

	for _, cellar := range homebrewCellars() {
		if ctx.Err() != nil {
			break
		}
		formulae, _ := os.ReadDir(cellar)
		for _, formula := range formulae {
			if formula.Name() == "node" || strings.HasPrefix(formula.Name(), "node@") {
				add(filepath.Join(cellar, formula.Name()), "bin", sourceHomebrew)
			}
		}
	}
	return stores
}

func nodeInstallations(ctx context.Context, stores []nodeVersionStore, name string) []string {
	if runtime.GOOS == "windows" {
		switch name {
		case "node":
			name = "node.exe"
		case "npm":
			name = "npm.cmd"
		}
	}
	var paths []string
	for _, store := range stores {
		if ctx.Err() != nil {
			return paths
		}
		versions, _ := os.ReadDir(store.directory)
		for _, version := range versions {
			if ctx.Err() != nil {
				return paths
			}
			if version.IsDir() || version.Type()&os.ModeSymlink != 0 {
				paths = append(paths, filepath.Join(store.directory, version.Name(), store.binDir, name))
			}
		}
	}
	return paths
}

func nodeStoreSource(path string, stores []nodeVersionStore) ToolSource {
	for _, store := range stores {
		if withinDirectory(path, store.directory) {
			return store.source
		}
	}
	return sourceUnchecked
}
