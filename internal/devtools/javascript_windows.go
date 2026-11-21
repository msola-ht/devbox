package devtools

import (
	"context"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func runJavaScriptVersion(ctx context.Context, path string, environment []string) ([]byte, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".cmd", ".bat":
		systemDir, err := windows.GetSystemDirectory()
		if err != nil {
			return nil, err
		}
		// Pass the quoted entry through the environment so cmd does not expand
		// percent signs in the path or apply Go's argument quoting to the script.
		environment = append(environment, `DEVBOX_VERSION_ENTRY="`+path+`"`, "NoDefaultCurrentDirectoryInExePath=1")
		args := []string{"/d", "/e:on", "/v:off", "/c", "%DEVBOX_VERSION_ENTRY%", "--version"}
		return runCommand(ctx, filepath.Join(systemDir, "cmd.exe"), args, environment...)
	default:
		return runCommand(ctx, path, []string{"--version"}, environment...)
	}
}
