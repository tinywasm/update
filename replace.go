package update

import (
	"fmt"
	"os"
)

// Swap and Rollback were MOVED (extracted) from tinywasm/deploy (handler.go:
// HandleUpdate, the "Backup Existing Binary" / "Move New Binary" / restore steps).
// They preserve deploy's proven os.Rename-based logic verbatim.

// Swap backs up targetPath (if it exists) to targetPath+".old" and renames
// newFilePath into its place. It returns the backup path ("" when targetPath did
// not exist). On install failure it restores the backup before returning the error.
func Swap(targetPath, newFilePath string) (string, error) {
	var backupPath string
	if _, err := os.Stat(targetPath); err == nil {
		backupPath = targetPath + ".old"
		if err := os.Rename(targetPath, backupPath); err != nil {
			return "", fmt.Errorf("failed to backup: %w", err)
		}
	}
	if err := os.Rename(newFilePath, targetPath); err != nil {
		if backupPath != "" {
			_ = os.Rename(backupPath, targetPath)
		}
		return "", fmt.Errorf("failed to install: %w", err)
	}
	return backupPath, nil
}

// Rollback restores backupPath over targetPath (used after a post-swap failure,
// e.g. the new process failed its health check).
func Rollback(targetPath, backupPath string) error {
	if backupPath == "" {
		return nil
	}
	return os.Rename(backupPath, targetPath)
}
