//go:build !windows

package analyze

import (
	"os"
	"testing"
)

func denyDirectoryReads(t *testing.T, path string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("permission denial is not reliable when tests run as root")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	originalMode := info.Mode().Perm()
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("cannot remove directory permissions: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, originalMode) })
	if _, err := os.ReadDir(path); err == nil {
		t.Skip("mode change did not prevent reading the directory")
	}
}
