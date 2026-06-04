package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// Moved from tinywasm/installer (app_binary_test.go: TestVerifyChecksum_RejectsMismatch).
func TestVerifyChecksum(t *testing.T) {
	data := []byte("binary-bytes")
	sums := "deadbeef  tinywasm-linux-amd64\n" // intentional mismatch
	if err := VerifyChecksum("tinywasm-linux-amd64", data, []byte(sums)); err == nil {
		t.Fatal("expected error on checksum mismatch")
	}

	// Correct checksum
	sum := sha256.Sum256(data)
	correctSums := fmt.Sprintf("%s  tinywasm-linux-amd64\n", hex.EncodeToString(sum[:]))
	if err := VerifyChecksum("tinywasm-linux-amd64", data, []byte(correctSums)); err != nil {
		t.Errorf("expected success on correct checksum, got: %v", err)
	}
}
