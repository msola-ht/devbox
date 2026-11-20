//go:build !darwin

package proxy

import "context"

func readSystemSettings(context.Context) ([]SystemEntry, bool, error) {
	return []SystemEntry{}, false, nil
}
