package analyze

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunContinuesAfterUnreadableDirectoryAtDepthOne(t *testing.T) {
	testUnreadableDirectory(t, "a_denied")
}

func TestRunContinuesAfterUnreadableDirectoryAtDepthTwo(t *testing.T) {
	testUnreadableDirectory(t, filepath.Join("top", "mid", "a_denied"))
}

func testUnreadableDirectory(t *testing.T, deniedRelative string) {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, root, "root.go", "package root\n")
	writeTestFile(t, root, filepath.ToSlash(filepath.Join(deniedRelative, "hidden.go")), "package hidden\n")

	siblingRelative := "z_after"
	if filepath.Dir(deniedRelative) != "." {
		siblingRelative = filepath.Join(filepath.Dir(deniedRelative), "z_after")
	}
	writeTestFile(t, root, filepath.ToSlash(filepath.Join(siblingRelative, "readable.go")), "package readable\n")

	deniedPath := filepath.Join(root, deniedRelative)
	denyDirectoryReads(t, deniedPath)

	_, records, warnings, err := Run([]string{root}, byteCounter{}, Options{Workers: 2})
	if err != nil {
		t.Fatalf("unreadable directory became fatal: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records = %#v, want root.go and readable sibling", records)
	}
	joinedRecords := records[0].Path + "\n" + records[1].Path
	if !strings.Contains(joinedRecords, "root.go") || !strings.Contains(joinedRecords, filepath.ToSlash(filepath.Join(siblingRelative, "readable.go"))) {
		t.Fatalf("readable sibling was lost: %#v", records)
	}
	if len(warnings) == 0 {
		t.Fatalf("unreadable directory produced no warning")
	}
	deniedSlash := filepath.ToSlash(deniedPath)
	if !containsWarningPath(warnings, deniedSlash) {
		t.Fatalf("warnings %#v do not identify denied directory %q", warnings, deniedSlash)
	}
}

func containsWarningPath(warnings []ScanWarning, path string) bool {
	for _, warning := range warnings {
		if warning.Stage == "walk" && warning.Path == path {
			return true
		}
	}
	return false
}
