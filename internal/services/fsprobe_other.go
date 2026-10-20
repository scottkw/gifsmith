//go:build !darwin

package services

import "errors"

// Statfs is unsupported off macOS, so ValidateDownloadDir refuses (folder_unwritable), never passes.
func Statfs(string) (string, bool, error) {
	return "", false, errors.New("statfs: unsupported on this OS")
}
