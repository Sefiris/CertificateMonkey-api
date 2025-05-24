package version

import (
	"os"
	"strings"
	"time"
)

// Build information (can be set via ldflags)
var (
	// Version is the current version, read from VERSION file
	Version = getVersionFromFile()

	// BuildTime can be set at build time via ldflags
	BuildTime = "unknown"

	// GitCommit can be set at build time via ldflags
	GitCommit = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = "unknown"
)

// Info represents version and build information
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
}

// Get returns the current version information
func Get() Info {
	return Info{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
	}
}

// GetVersion returns just the version string
func GetVersion() string {
	return Version
}

// getVersionFromFile reads the version from the VERSION file
func getVersionFromFile() string {
	// Try project root first (normal case)
	if content, err := os.ReadFile("VERSION"); err == nil {
		if version := strings.TrimSpace(string(content)); version != "" {
			return version
		}
	}

	// Try one level up (for tests running from internal/version/)
	if content, err := os.ReadFile("../../VERSION"); err == nil {
		if version := strings.TrimSpace(string(content)); version != "" {
			return version
		}
	}

	// Fallback version if file can't be read
	return "0.1.0-dev"
}

// GetBuildInfo returns formatted build information
func GetBuildInfo() map[string]interface{} {
	info := Get()

	return map[string]interface{}{
		"service":    "certificate-monkey",
		"version":    info.Version,
		"build_time": info.BuildTime,
		"git_commit": info.GitCommit,
		"go_version": info.GoVersion,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
}
