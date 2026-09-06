package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// VerifyChecksum was MOVED verbatim from webtyp/installer (mode_binary.go:
// verifyChecksum) — only the name was exported. It verifies data's SHA256 against
// the entry for asset in a checksums.txt body ("<hex>  <asset>" per line).
func VerifyChecksum(asset string, data []byte, sums []byte) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])

	lines := strings.Split(string(sums), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		if parts[1] == asset {
			if parts[0] == got {
				return nil
			}
			return fmt.Errorf("checksum mismatch for %s: want %s, got %s", asset, parts[0], got)
		}
	}

	return fmt.Errorf("checksum for %s not found in checksums.txt", asset)
}
