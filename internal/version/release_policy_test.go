package version

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func extractWorkflowStepScript(t *testing.T, workflowText, stepName string) string {
	t.Helper()
	lines := strings.Split(workflowText, "\n")
	target := "- name: " + stepName
	stepIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == target {
			stepIdx = i
			break
		}
	}
	if stepIdx == -1 {
		t.Fatalf("step %q not found in release workflow", stepName)
	}

	runIdx := -1
	baseIndent := 0
	for i := stepIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- name:") || strings.HasPrefix(trimmed, "- uses:") {
			break
		}
		if trimmed == "run: |" {
			runIdx = i
			baseIndent = len(lines[i]) - len(strings.TrimLeft(lines[i], " "))
			break
		}
	}
	if runIdx == -1 {
		t.Fatalf("multiline run block not found for step %q", stepName)
	}

	var scriptLines []string
	contentIndent := -1
	for i := runIdx + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			scriptLines = append(scriptLines, "")
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent <= baseIndent {
			break
		}
		if contentIndent == -1 {
			contentIndent = indent
		}
		if len(line) >= contentIndent {
			scriptLines = append(scriptLines, line[contentIndent:])
		} else {
			scriptLines = append(scriptLines, strings.TrimLeft(line, " "))
		}
	}
	return strings.Join(scriptLines, "\n")
}

func runGitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmdArgs := append([]string{"-c", "core.hooksPath=/dev/null"}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func parseGitHubOutput(t *testing.T, data string) map[string]string {
	t.Helper()
	out := make(map[string]string)
	lines := strings.Split(strings.TrimRight(data, "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}
		if idx := strings.Index(line, "<<"); idx != -1 {
			key := line[:idx]
			delim := line[idx+2:]
			var valLines []string
			for i = i + 1; i < len(lines); i++ {
				if lines[i] == delim {
					break
				}
				valLines = append(valLines, lines[i])
			}
			out[key] = strings.Join(valLines, "\n")
			continue
		}
		if idx := strings.Index(line, "="); idx != -1 {
			out[line[:idx]] = line[idx+1:]
		}
	}
	return out
}

func TestReleaseWorkflowSemVerGateAndDockerGuard(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	wfBytes, err := os.ReadFile(filepath.Join(repoRoot, ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatalf("read release.yml: %v", err)
	}
	wfText := string(wfBytes)

	goreleaserBytes, err := os.ReadFile(filepath.Join(repoRoot, ".goreleaser.yml"))
	if err != nil {
		t.Fatalf("read .goreleaser.yml: %v", err)
	}
	if !strings.Contains(string(goreleaserBytes), "prerelease: auto") {
		t.Fatal(".goreleaser.yml must configure prerelease: auto")
	}

	if !strings.Contains(wfText, "needs: verify") {
		t.Fatal("docker build jobs must depend on verify")
	}

	step1Script := extractWorkflowStepScript(t, wfText, "Verify SemVer release tag")
	metaScript := extractWorkflowStepScript(t, wfText, "Verify tagged commit and resolve release metadata")

	validTags := []string{"v0.0.1", "v0.1.0", "v1.2.3", "v0.1.0-alpha.1", "v0.1.0-beta.2", "v0.1.0-rc.10"}
	for _, tag := range validTags {
		cmd := exec.Command("bash", "-c", step1Script)
		cmd.Env = append(os.Environ(), "RELEASE_TAG="+tag)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("expected valid tag %q to pass step-1 gate, got %v: %s", tag, err, out)
		}
	}

	invalidTags := []string{"v0.1.0-fix.1", "v01.2.0", "v0.1.0-rc.0", "v0.1.0-alpha.01", "1.2.3", "v1.2", "main", ""}
	for _, tag := range invalidTags {
		cmd := exec.Command("bash", "-c", step1Script)
		cmd.Env = append(os.Environ(), "RELEASE_TAG="+tag)
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("expected invalid tag %q to fail step-1 gate, output: %s", tag, out)
		}
	}

	tmpRepo := t.TempDir()
	runGitCmd(t, tmpRepo, "init")
	runGitCmd(t, tmpRepo, "config", "user.email", "test@example.com")
	runGitCmd(t, tmpRepo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(tmpRepo, "README.md"), []byte("init\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitCmd(t, tmpRepo, "add", "README.md")
	runGitCmd(t, tmpRepo, "commit", "-m", "init")
	runGitCmd(t, tmpRepo, "tag", "v0.2.0")
	runGitCmd(t, tmpRepo, "tag", "v0.2.0-rc.1")

	runMeta := func(tag string) (map[string]string, error, string) {
		ghOutFile := filepath.Join(t.TempDir(), "github_output")
		cmd := exec.Command("bash", "-c", metaScript)
		cmd.Dir = tmpRepo
		cmd.Env = append(os.Environ(),
			"RELEASE_TAG="+tag,
			"REPOSITORY_OWNER=Samzong",
			"GITHUB_OUTPUT="+ghOutFile,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err, string(out)
		}
		raw, readErr := os.ReadFile(ghOutFile)
		if readErr != nil {
			t.Fatalf("read GITHUB_OUTPUT: %v", readErr)
		}
		return parseGitHubOutput(t, string(raw)), nil, string(out)
	}

	stableOut, err, rawOut := runMeta("v0.2.0")
	if err != nil {
		t.Fatalf("stable tag metadata resolution failed: %v\n%s", err, rawOut)
	}
	if stableOut["is_prerelease"] != "false" || stableOut["version"] != "0.2.0" {
		t.Fatalf("unexpected stable outputs: %#v", stableOut)
	}
	wantStableModelfs := "ghcr.io/samzong/modelfs:v0.2.0\nghcr.io/samzong/modelfs:latest"
	if stableOut["modelfs_tags"] != wantStableModelfs {
		t.Fatalf("modelfs_tags = %q, want %q", stableOut["modelfs_tags"], wantStableModelfs)
	}
	wantStableUI := "ghcr.io/samzong/modelfs-ui-server:v0.2.0\nghcr.io/samzong/modelfs-ui-server:latest"
	if stableOut["modelfs_ui_server_tags"] != wantStableUI {
		t.Fatalf("modelfs_ui_server_tags = %q, want %q", stableOut["modelfs_ui_server_tags"], wantStableUI)
	}

	preOut, err, rawOut := runMeta("v0.2.0-rc.1")
	if err != nil {
		t.Fatalf("prerelease tag metadata resolution failed: %v\n%s", err, rawOut)
	}
	if preOut["is_prerelease"] != "true" || preOut["version"] != "0.2.0-rc.1" {
		t.Fatalf("unexpected prerelease outputs: %#v", preOut)
	}
	if strings.Contains(preOut["modelfs_tags"], ":latest") || preOut["modelfs_tags"] != "ghcr.io/samzong/modelfs:v0.2.0-rc.1" {
		t.Fatalf("prerelease modelfs_tags leaked latest or mismatched: %q", preOut["modelfs_tags"])
	}
	if strings.Contains(preOut["modelfs_ui_server_tags"], ":latest") || preOut["modelfs_ui_server_tags"] != "ghcr.io/samzong/modelfs-ui-server:v0.2.0-rc.1" {
		t.Fatalf("prerelease modelfs_ui_server_tags leaked latest or mismatched: %q", preOut["modelfs_ui_server_tags"])
	}

	if err := os.WriteFile(filepath.Join(tmpRepo, "README.md"), []byte("moved\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitCmd(t, tmpRepo, "add", "README.md")
	runGitCmd(t, tmpRepo, "commit", "-m", "second")
	if _, err, out := runMeta("v0.2.0"); err == nil {
		t.Fatalf("expected commit mismatch to fail, got output: %s", out)
	}
}
