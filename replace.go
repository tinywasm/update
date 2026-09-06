package update

import (
	"fmt"
	"io"
	"os"
)

// Swap and Rollback were MOVED (extracted) from webtyp/deploy (handler.go:
// HandleUpdate, the "Backup Existing Binary" / "Move New Binary" / restore steps).
// They preserve deploy's proven os.Rename-based logic verbatim.

// Swap backs up targetPath (if it exists) to targetPath+".old" and renames
// newFilePath into its place. It returns the backup path ("" when targetPath did
// not exist). On install failure it restores the backup before returning the error.
func Swap(targetPath, newFilePath string) (string, error) {
	var backupPath string
	if _, err := os.Stat(targetPath); err == nil {
		backupPath = targetPath + ".old"
		if err := moveFile(targetPath, backupPath); err != nil {
			return "", fmt.Errorf("failed to backup: %w", err)
		}
	}
	if err := moveFile(newFilePath, targetPath); err != nil {
		if backupPath != "" {
			_ = moveFile(backupPath, targetPath)
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
	return moveFile(backupPath, targetPath)
}

// moveFile renames src to dst, falling back to copy+remove across filesystems (EXDEV).
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}
