// Package version provides hive version and build information.
package version

import (
	"regexp"
	"runtime"
	"runtime/debug"
)

// Version is "dev" unless a release build sets it with -ldflags -X.
var Version = "dev"

// releaseVersion matches vX.Y.Z and vX.Y.Z-rc.N with no other suffix.
var releaseVersion = regexp.MustCompile(`^v\d+\.\d+\.\d+(-rc\.\d+)?$`)

// BuildInfo contains the version and build data for hive.
type BuildInfo struct {
	Version  string
	Commit   string
	Modified bool
	Go       string
}

// Info returns the version and build data for hive.
func Info() BuildInfo {
	return info(debug.ReadBuildInfo)
}

func info(readBuildInfo func() (*debug.BuildInfo, bool)) BuildInfo {
	result := BuildInfo{
		Version: Version,
		Commit:  "unknown",
		Go:      runtime.Version(),
	}
	buildInfo, ok := readBuildInfo()
	if !ok {
		return result
	}

	result.Go = buildInfo.GoVersion
	if result.Version == "dev" && releaseVersion.MatchString(buildInfo.Main.Version) {
		result.Version = buildInfo.Main.Version
	}
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			result.Commit = setting.Value
			if len(result.Commit) > 12 {
				result.Commit = result.Commit[:12]
			}
		case "vcs.modified":
			result.Modified = setting.Value == "true"
		}
	}
	return result
}
