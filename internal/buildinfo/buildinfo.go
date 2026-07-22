// Package buildinfo reports the version of the tloc binary.
package buildinfo

import (
	"runtime/debug"
	"strings"
)

// version is populated by GoReleaser through -ldflags. Keep it unexported so
// callers use Version and get the module-build fallback used by go install.
var version string

// Version returns the release version without a leading "v". Development
// builds for which neither an ldflag nor module version is available report
// "devel".
func Version() string {
	if normalized := normalize(version); normalized != "" {
		return normalized
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	if normalized := normalize(info.Main.Version); normalized != "" {
		return normalized
	}
	return "devel"
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "(devel)" {
		return ""
	}
	return strings.TrimPrefix(value, "v")
}
