package capbypass

import "runtime/debug"

const modulePath = "github.com/CapBypass-Development/capbypass-sdk-go"

// sdkVersion is the version of this module as the consumer's build resolved it.
// Release versions are git tags created by the publish workflow, so the Go
// toolchain already records them in the consuming binary's build info: nothing
// in the source is edited per release and nothing can drift.
var sdkVersion = moduleVersion()

var userAgent = "capbypass-sdk-go/" + sdkVersion

func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return versionFromBuildInfo(info)
}

func versionFromBuildInfo(info *debug.BuildInfo) string {
	for _, dep := range info.Deps {
		if dep.Path != modulePath {
			continue
		}
		if dep.Replace != nil {
			// A replace to a local directory carries no version: unreleased code.
			if dep.Replace.Version == "" {
				return "devel"
			}
			return dep.Replace.Version
		}
		if dep.Version != "" {
			return dep.Version
		}
	}
	// This module is the main module (tests, examples built in-tree).
	if info.Main.Path == modulePath && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "devel"
}
