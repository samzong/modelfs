package version

import (
	"runtime/debug"
	"testing"
)

func TestResolve(t *testing.T) {
	fromInfo := func(mainVersion string, ok bool) func() (*debug.BuildInfo, bool) {
		return func() (*debug.BuildInfo, bool) {
			if !ok {
				return nil, false
			}
			return &debug.BuildInfo{
				Main: debug.Module{Version: mainVersion},
			}, true
		}
	}

	cases := []struct {
		name      string
		injected  string
		buildInfo func() (*debug.BuildInfo, bool)
		want      string
	}{
		{
			name:      "linker v-prefixed version wins over build info",
			injected:  "v0.2.0",
			buildInfo: fromInfo("v0.9.9", true),
			want:      "0.2.0",
		},
		{
			name:      "linker prerelease version wins over build info",
			injected:  "0.2.0-rc.1",
			buildInfo: fromInfo("v0.9.9", true),
			want:      "0.2.0-rc.1",
		},
		{
			name:      "dev falls back to read build info module version",
			injected:  "dev",
			buildInfo: fromInfo("v0.3.0", true),
			want:      "0.3.0",
		},
		{
			name:      "empty falls back to read build info module version",
			injected:  "",
			buildInfo: fromInfo("v0.3.1-beta.2", true),
			want:      "0.3.1-beta.2",
		},
		{
			name:      "devel build info keeps dev",
			injected:  "dev",
			buildInfo: fromInfo("(devel)", true),
			want:      "dev",
		},
		{
			name:      "missing build info keeps dev",
			injected:  "dev",
			buildInfo: fromInfo("", false),
			want:      "dev",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(tc.injected, tc.buildInfo)
			if got != tc.want {
				t.Fatalf("Resolve(%q) = %q, want %q", tc.injected, got, tc.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	got := Info{Version: "v1.2.3", Commit: "abc1234", Date: "2026-01-02T03:04:05Z"}.Format(nil)
	want := "1.2.3 (commit abc1234, built at 2026-01-02T03:04:05Z)"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}

	fallback := Info{Version: "dev"}.Format(func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.4.0"}}, true
	})
	wantFallback := "0.4.0 (commit unknown, built at unknown)"
	if fallback != wantFallback {
		t.Fatalf("Format() fallback = %q, want %q", fallback, wantFallback)
	}
}
