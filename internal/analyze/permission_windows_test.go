//go:build windows

package analyze

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func denyDirectoryReads(t *testing.T, path string) {
	t.Helper()
	principalOutput, err := exec.Command("whoami.exe").Output()
	if err != nil {
		t.Skipf("cannot determine Windows principal: %v", err)
	}
	principal := strings.TrimSpace(string(principalOutput))
	if principal == "" {
		t.Skip("whoami returned an empty Windows principal")
	}

	denied := false
	t.Cleanup(func() {
		if denied {
			_, _ = exec.Command("icacls.exe", path, "/remove:d", principal).CombinedOutput()
		}
	})
	ace := fmt.Sprintf("%s:(OI)(CI)(RX)", principal)
	if output, err := exec.Command("icacls.exe", path, "/deny", ace).CombinedOutput(); err != nil {
		t.Skipf("cannot deny directory reads with icacls: %v: %s", err, output)
	}
	denied = true
	if _, err := os.ReadDir(path); err == nil {
		t.Skip("Windows ACL denial did not prevent reading the directory")
	}
}
