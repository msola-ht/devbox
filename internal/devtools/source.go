package devtools

import (
	"path/filepath"
	"runtime"
	"strings"
)

// ToolSource identifies the recognized provider of a detected executable.
type ToolSource string

const (
	sourceUnchecked           ToolSource = ""
	sourceUnknown             ToolSource = "unknown"
	sourceSystem              ToolSource = "system"
	sourceAppleDeveloperTools ToolSource = "apple_developer_tools"
	sourceHomebrew            ToolSource = "homebrew"
	sourceFNM                 ToolSource = "fnm"
	sourceNVM                 ToolSource = "nvm"
	sourceNVMWindows          ToolSource = "nvm_windows"
	sourceUV                  ToolSource = "uv"
	sourceCorepack            ToolSource = "corepack"
)

func sourceForExecutable(path string) ToolSource {
	if runtime.GOOS == "darwin" {
		switch filepath.Dir(path) {
		case "/bin", "/usr/bin":
			return sourceSystem
		}
		if withinDirectory(path, "/Library/Developer/CommandLineTools") {
			return sourceAppleDeveloperTools
		}
		if bundle, _, found := strings.Cut(path, "/Contents/Developer/"); found {
			name := filepath.Base(bundle)
			if name == "Xcode.app" || (strings.HasPrefix(name, "Xcode-") && strings.HasSuffix(name, ".app")) {
				return sourceAppleDeveloperTools
			}
		}
	}
	for _, cellar := range homebrewCellars() {
		if withinDirectory(path, cellar) {
			return sourceHomebrew
		}
	}
	return sourceUnchecked
}

func homebrewCellars() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/opt/homebrew/Cellar", "/usr/local/Cellar"}
	case "linux":
		return []string{"/home/linuxbrew/.linuxbrew/Cellar"}
	default:
		return nil
	}
}

func withinDirectory(path, dir string) bool {
	relative, err := filepath.Rel(dir, path)
	return err == nil && relative != "." && filepath.IsLocal(relative)
}
