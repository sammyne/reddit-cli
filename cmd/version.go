package cmd

import (
	"runtime/debug"
)

// version is the release tag, injected at build time via ldflags.
var version = "dev"

func gitCommit() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && len(s.Value) >= 7 {
			return s.Value[:7]
		}
	}
	return "unknown"
}

func fullVersion() string {
	return version + ", build " + gitCommit()
}
