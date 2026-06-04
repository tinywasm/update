package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSwapAndRollback(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tool")
	src := filepath.Join(dir, "tool.new")
	os.WriteFile(target, []byte("old"), 0755)
	os.WriteFile(src, []byte("new"), 0755)

	backup, err := Swap(target, src)
	if err != nil || backup == "" {
		t.Fatalf("swap: backup=%q err=%v", backup, err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Errorf("target=%q want new", got)
	}
	if err := Rollback(target, backup); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if got, _ := os.ReadFile(target); string(got) != "old" {
		t.Errorf("after rollback target=%q want old", got)
	}
}

func TestSwapFreshInstall(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "tool") // inexistente
	src := filepath.Join(dir, "tool.new")
	os.WriteFile(src, []byte("new"), 0755)
	backup, err := Swap(target, src)
	if err != nil || backup != "" {
		t.Fatalf("fresh: backup=%q err=%v", backup, err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Errorf("target=%q want new", got)
	}
}
