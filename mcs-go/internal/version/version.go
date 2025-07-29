package version

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Build information. These variables are set at build time using ldflags.
var (
	// Version is the semantic version (or "dev" for development builds)
	Version = "dev"
	
	// GitCommit is the git commit hash
	GitCommit = "unknown"
	
	// BuildTime is the time of the build
	BuildTime = "unknown"
	
	// GitTag is the git tag if this commit has one
	GitTag = ""
	
	// GitDirty indicates if there were uncommitted changes
	GitDirty = ""
)

// Info returns version information
func Info() string {
	v := Version
	
	// For dev builds, include commit info
	if v == "dev" && GitCommit != "unknown" && GitCommit != "" {
		commit := GitCommit
		if len(commit) > 8 {
			commit = commit[:8]
		}
		
		// Format: dev-{commit}[-dirty]
		v = fmt.Sprintf("dev-%s", commit)
		if GitDirty == "true" {
			v += "-dirty"
		}
	}
	
	return v
}

// DetailedInfo returns detailed version information
func DetailedInfo() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("Version:    %s", Info()))
	
	if GitCommit != "unknown" && GitCommit != "" {
		parts = append(parts, fmt.Sprintf("Git commit: %s", GitCommit))
	}
	
	if GitTag != "" {
		parts = append(parts, fmt.Sprintf("Git tag:    %s", GitTag))
	}
	
	if BuildTime != "unknown" {
		parts = append(parts, fmt.Sprintf("Built:      %s", BuildTime))
	}
	
	parts = append(parts, fmt.Sprintf("Go version: %s", runtime.Version()))
	parts = append(parts, fmt.Sprintf("OS/Arch:    %s/%s", runtime.GOOS, runtime.GOARCH))
	
	return strings.Join(parts, "\n")
}

// IsDevBuild returns true if this is a development build
func IsDevBuild() bool {
	return Version == "dev" || strings.HasPrefix(Version, "dev-")
}

// IsPreRelease returns true if this is a pre-release version
func IsPreRelease() bool {
	return strings.Contains(Version, "-beta") || 
	       strings.Contains(Version, "-rc") || 
	       strings.Contains(Version, "-alpha")
}

// BuildDate returns the build time as a parsed time.Time
func BuildDate() time.Time {
	if BuildTime == "unknown" {
		return time.Time{}
	}
	
	t, err := time.Parse(time.RFC3339, BuildTime)
	if err != nil {
		return time.Time{}
	}
	
	return t
}