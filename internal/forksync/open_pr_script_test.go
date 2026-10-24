package forksync

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenUpstreamSyncPROpensDraftWhenMergeConflicts(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required")
	}

	root := t.TempDir()
	upstream := filepath.Join(root, "upstream.git")
	origin := filepath.Join(root, "origin.git")
	fork := filepath.Join(root, "fork")
	initGitRepo(t, upstream)
	writeGitFile(t, upstream, "shared.txt", "common base\n")
	runGit(t, upstream, "add", "-A")
	runGit(t, upstream, "commit", "-m", "common base")

	runGit(t, filepath.Dir(fork), "clone", upstream, fork)
	runGit(t, fork, "config", "user.email", "test@example.com")
	runGit(t, fork, "config", "user.name", "test")
	writeGitFile(t, fork, "shared.txt", "fork kept this\n")
	writeGitFile(t, fork, "only-fork.txt", "from fork\n")
	runGit(t, fork, "add", "-A")
	runGit(t, fork, "commit", "-m", "fork main")

	writeGitFile(t, upstream, "shared.txt", "upstream latest\n")
	writeGitFile(t, upstream, "only-upstream.txt", "from upstream\n")
	runGit(t, upstream, "add", "-A")
	runGit(t, upstream, "commit", "-m", "upstream v9.9.9")
	runGit(t, upstream, "tag", "v9.9.9")

	if err := os.MkdirAll(origin, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, origin, "init", "--bare", "-b", "main")
	runGit(t, fork, "remote", "set-url", "origin", origin)

	scriptDir := filepath.Join(fork, "scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(findRepoRoot(t), "scripts", "open-upstream-sync-pr.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptDir, "open-upstream-sync-pr.sh"), src, 0o755); err != nil {
		t.Fatal(err)
	}

	ghLog := filepath.Join(root, "gh.log")
	path := fakeBinPath(t, root)
	cmd := exec.Command("bash", filepath.Join(scriptDir, "open-upstream-sync-pr.sh"), "v9.9.9")
	cmd.Dir = fork
	cmd.Env = isolatedGitEnv(
		"PATH="+path,
		"GH_LOG="+ghLog,
		"UPSTREAM_REPO=MHSanaei/3x-ui",
		"UPSTREAM_GIT_URL="+upstream,
		"BASE_BRANCH=main",
		"GH_TOKEN=test",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("script failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "conflict") || !strings.Contains(string(out), "shared.txt") {
		t.Fatalf("expected the script to name the conflicted file, got:\n%s", out)
	}

	body, err := os.ReadFile(ghLog)
	if err != nil {
		t.Fatalf("gh was not invoked: %v", err)
	}
	logged := string(body)
	if !strings.Contains(logged, "pr create") || !strings.Contains(logged, "--draft") {
		t.Fatalf("gh pr create --draft was not invoked:\n%s", logged)
	}
	if !strings.Contains(logged, "sync: merge upstream v9.9.9 (conflicts)") {
		t.Fatalf("PR title missing the conflicted release tag:\n%s", logged)
	}

	want := gitOutput(t, upstream, "rev-parse", "v9.9.9^{commit}")
	got := gitOutput(t, origin, "rev-parse", "refs/heads/sync/upstream-v9.9.9")
	if got != want {
		t.Fatalf("pushed sync branch %s, want the upstream tag commit %s\n%s", got, want, out)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
}

func writeGitFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	_ = gitOutput(t, dir, args...)
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = isolatedGitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func isolatedGitEnv(extra ...string) []string {
	vars := map[string]string{
		"GIT_CONFIG_NOSYSTEM": "1",
		"GIT_CONFIG_GLOBAL":   "/dev/null",
		"GIT_CONFIG_SYSTEM":   "/dev/null",
		"GIT_AUTHOR_NAME":     "test",
		"GIT_AUTHOR_EMAIL":    "test@example.com",
		"GIT_COMMITTER_NAME":  "test",
		"GIT_COMMITTER_EMAIL": "test@example.com",
		"HOME":                "/dev/null",
	}
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		switch k {
		case "PATH", "TMPDIR", "XDG_CONFIG_HOME":
			vars[k] = v
		}
	}
	for _, kv := range extra {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		vars[k] = v
	}
	env := make([]string, 0, len(vars))
	for k, v := range vars {
		env = append(env, k+"="+v)
	}
	return env
}

func fakeBinPath(t *testing.T, root string) string {
	t.Helper()
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GH_LOG\"\necho https://example.test/pr/1\n"
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin + string(os.PathListSeparator) + os.Getenv("PATH")
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "scripts", "open-upstream-sync-pr.sh")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
