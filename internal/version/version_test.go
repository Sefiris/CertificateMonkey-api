package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Version should not be empty")
	}

	// Should follow semver format (basic check)
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		t.Errorf("Version should be in semver format, got: %s", version)
	}
}

func TestGet(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	// Build info might be "unknown" in tests
	if info.BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}
}

func TestGetBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo()

	requiredFields := []string{"service", "version", "build_time", "git_commit", "go_version", "timestamp"}

	for _, field := range requiredFields {
		if _, exists := buildInfo[field]; !exists {
			t.Errorf("Build info should contain field: %s", field)
		}
	}

	if buildInfo["service"] != "certificate-monkey" {
		t.Errorf("Expected service name 'certificate-monkey', got: %s", buildInfo["service"])
	}
}
