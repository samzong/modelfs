package main

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestUIServerVersionFlag(t *testing.T) {
	origVersion, origCommit, origDate := version, commit, date
	t.Cleanup(func() {
		version, commit, date = origVersion, origCommit, origDate
	})

	version = "dev"
	commit = "unknown"
	date = "unknown"

	var stdout bytes.Buffer
	started := false
	err := runServer(
		[]string{"--version"},
		&stdout,
		flag.NewFlagSet("ui-server", flag.ContinueOnError),
		func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: "v0.4.0"}}, true
		},
		func(serverConfig) error {
			started = true
			return nil
		},
	)
	if err != nil {
		t.Fatalf("runServer(--version) error: %v", err)
	}
	if started {
		t.Fatal("server started when --version was passed")
	}
	if got, want := stdout.String(), "0.4.0 (commit unknown, built at unknown)\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}

	version = "v1.2.3"
	commit = "abc1234"
	date = "2026-01-02T03:04:05Z"
	stdout.Reset()
	err = runServer(
		[]string{"--version"},
		&stdout,
		flag.NewFlagSet("ui-server", flag.ContinueOnError),
		func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Version: "v0.9.9"}}, true
		},
		func(serverConfig) error {
			started = true
			return nil
		},
	)
	if err != nil {
		t.Fatalf("runServer(--version) injected error: %v", err)
	}
	if got, want := stdout.String(), "1.2.3 (commit abc1234, built at 2026-01-02T03:04:05Z)\n"; got != want {
		t.Fatalf("injected stdout = %q, want %q", got, want)
	}
}

func TestUIServerBinaryVersionInjection(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "ui-server")
	build := exec.Command("go", "build", "-ldflags", "-X main.version=v1.2.3 -X main.commit=deadbeef -X main.date=2026-01-02T03:04:05Z", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("ui-server --version failed: %v\n%s", err, out)
	}
	if got, want := string(out), "1.2.3 (commit deadbeef, built at 2026-01-02T03:04:05Z)\n"; got != want {
		t.Fatalf("binary output = %q, want %q", got, want)
	}

	noVCSBin := filepath.Join(t.TempDir(), "ui-server-novcs")
	buildNoVCS := exec.Command("go", "build", "-buildvcs=false", "-o", noVCSBin, ".")
	if out, err := buildNoVCS.CombinedOutput(); err != nil {
		t.Fatalf("go build -buildvcs=false failed: %v\n%s", err, out)
	}
	noVCSOut, err := exec.Command(noVCSBin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("novcs ui-server --version failed: %v\n%s", err, noVCSOut)
	}
	if got, want := string(noVCSOut), "dev (commit unknown, built at unknown)\n"; got != want {
		t.Fatalf("novcs binary output = %q, want %q", got, want)
	}

	if _, err := os.Stat(filepath.Join("..", "..", ".git")); err == nil {
		vcsBin := filepath.Join(t.TempDir(), "ui-server-vcs")
		buildVCS := exec.Command("go", "build", "-o", vcsBin, ".")
		if out, err := buildVCS.CombinedOutput(); err != nil {
			t.Fatalf("go build vcs failed: %v\n%s", err, out)
		}
		vcsOut, err := exec.Command(vcsBin, "--version").CombinedOutput()
		if err != nil {
			t.Fatalf("vcs ui-server --version failed: %v\n%s", err, vcsOut)
		}
		got := string(vcsOut)
		if strings.HasPrefix(got, "dev ") || strings.HasPrefix(got, "v") || !strings.HasSuffix(got, " (commit unknown, built at unknown)\n") {
			t.Fatalf("unexpected vcs ReadBuildInfo output = %q", got)
		}
	}
}
