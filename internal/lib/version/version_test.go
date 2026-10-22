package version

import (
	"runtime/debug"
	"testing"
)

func TestInfoReadsBuildSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings []debug.BuildSetting
		goVer    string
		want     BuildInfo
	}{
		{
			name: "revision and modified",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "123456789abcdef0123456789abcdef012345678"},
				{Key: "vcs.modified", Value: "true"},
			},
			goVer: "go1.27.2-custom",
			want: BuildInfo{
				Version:  "dev",
				Commit:   "123456789abc",
				Modified: true,
				Go:       "go1.27.2-custom",
			},
		},
		{
			name:  "missing revision",
			goVer: "go1.27.2-custom",
			want: BuildInfo{
				Version: "dev",
				Commit:  "unknown",
				Go:      "go1.27.2-custom",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := info(func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{GoVersion: tt.goVer, Settings: tt.settings}, true
			})
			if got != tt.want {
				t.Fatalf("info() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestInfoVersionRule(t *testing.T) {
	tests := []struct {
		name        string
		ldflags     string
		mainVersion string
		want        string
	}{
		{name: "ldflags value wins", ldflags: "v9.9.9", mainVersion: "v1.2.3", want: "v9.9.9"},
		{name: "release version", ldflags: "dev", mainVersion: "v1.2.3", want: "v1.2.3"},
		{name: "release candidate", ldflags: "dev", mainVersion: "v1.2.3-rc.4", want: "v1.2.3-rc.4"},
		{name: "pseudo-version", ldflags: "dev", mainVersion: "v0.0.3-0.20261022041405-6f14e3bd53b9", want: "dev"},
		{name: "dirty", ldflags: "dev", mainVersion: "v1.2.3+dirty", want: "dev"},
		{name: "devel", ldflags: "dev", mainVersion: "(devel)", want: "dev"},
		{name: "empty", ldflags: "dev", mainVersion: "", want: "dev"},
		{name: "other suffix", ldflags: "dev", mainVersion: "v1.2.3-beta.1", want: "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved := Version
			Version = tt.ldflags
			t.Cleanup(func() { Version = saved })

			got := info(func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{Main: debug.Module{Version: tt.mainVersion}}, true
			})
			if got.Version != tt.want {
				t.Fatalf("info().Version = %q, want %q", got.Version, tt.want)
			}
		})
	}
}
