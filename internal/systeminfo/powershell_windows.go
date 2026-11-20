package systeminfo

import (
	"context"

	"devbox/internal/localcmd"
)

func powershellJSON(ctx context.Context, query string) ([]byte, error) {
	const setup = `$ErrorActionPreference = 'Stop'; [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false); `
	script := setup + `ConvertTo-Json -Compress -InputObject @(` + query + `)`
	args := []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-Command", script}
	return localcmd.Output(ctx, "powershell.exe", args, nil)
}
