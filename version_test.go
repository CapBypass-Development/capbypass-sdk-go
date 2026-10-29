package capbypass

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionFromBuildInfo(t *testing.T) {
	dep := func(path, version string, replace *debug.Module) *debug.Module {
		return &debug.Module{Path: path, Version: version, Replace: replace}
	}

	tests := []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{
			name: "consumer build resolves the tagged module",
			info: &debug.BuildInfo{
				Main: debug.Module{Path: "example.com/app", Version: "(devel)"},
				Deps: []*debug.Module{dep("github.com/other/mod", "v9.9.9", nil), dep(modulePath, "v1.0.10", nil)},
			},
			want: "v1.0.10",
		},
		{
			name: "pseudo-version of an untagged commit",
			info: &debug.BuildInfo{Deps: []*debug.Module{dep(modulePath, "v1.0.10-0.20260101000000-abcdef123456", nil)}},
			want: "v1.0.10-0.20260101000000-abcdef123456",
		},
		{
			name: "replace to a versioned module",
			info: &debug.BuildInfo{Deps: []*debug.Module{dep(modulePath, "v1.0.9", &debug.Module{Path: "example.com/fork", Version: "v1.2.3"})}},
			want: "v1.2.3",
		},
		{
			name: "replace to a local directory",
			info: &debug.BuildInfo{Deps: []*debug.Module{dep(modulePath, "v1.0.9", &debug.Module{Path: "../sdk"})}},
			want: "devel",
		},
		{
			name: "main module in-tree",
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "(devel)"}},
			want: "devel",
		},
		{
			name: "main module built from a tag",
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v1.0.10"}},
			want: "v1.0.10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, versionFromBuildInfo(tt.info))
		})
	}
}
