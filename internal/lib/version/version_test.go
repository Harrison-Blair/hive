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
