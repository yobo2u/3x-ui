package panel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPanelUpdateURLsUseThisFork(t *testing.T) {
	if got, want := panelReleaseAPI(""), "https://api.github.com/repos/yobo2u/3x-ui/releases/latest"; got != want {
		t.Fatalf("latest release URL = %s, want %s", got, want)
	}
	if got, want := panelReleaseAPI("v3.10.2"), "https://api.github.com/repos/yobo2u/3x-ui/releases/tags/v3.10.2"; got != want {
		t.Fatalf("tagged release URL = %s, want %s", got, want)
	}
	if strings.Contains(panelUpdaterURL, "MHSanaei") || strings.Contains(panelUpdaterURL, "mhsanaei") {
		t.Fatalf("panel updater still downloads upstream update.sh: %s", panelUpdaterURL)
	}
	if panelUpdaterURL != "https://raw.githubusercontent.com/yobo2u/3x-ui/main/update.sh" {
		t.Fatalf("updater URL = %s", panelUpdaterURL)
	}
}

func TestDockerImageIsBuiltFromThisFork(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	const want = "ghcr.io/yobo2u/3x-ui"
	const forbidden = "ghcr.io/mhsanaei/3x-ui"
	mustName := []string{
		".github/workflows/docker.yml",
		"docker-compose.yml",
	}
	for _, name := range mustName {
		text := readRepoFile(t, root, name)
		if !strings.Contains(text, want) {
			t.Fatalf("%s does not publish %s", name, want)
		}
		if strings.Contains(text, forbidden) || strings.Contains(text, "hsanaeii/3x-ui") {
			t.Fatalf("%s still points at an upstream image", name)
		}
	}
	installTS := readRepoFile(t, root, "docs/lib/xray/install.ts")
	if !strings.Contains(installTS, "ghcr.io/${productRepo.user}/${productRepo.repo}") {
		t.Fatal("install.ts docker image is not built from productRepo")
	}
	if strings.Contains(installTS, forbidden) || strings.Contains(installTS, "hsanaeii/3x-ui") {
		t.Fatal("install.ts still points at an upstream image")
	}
	if !strings.Contains(readRepoFile(t, root, "docs/lib/shared.ts"), "user: 'yobo2u'") {
		t.Fatal("productRepo is not yobo2u, so the published image name would change")
	}
	for _, name := range []string{
		"README.md",
		"docs/content/docs/en/guide/installation.mdx",
		"docs/content/docs/zh/guide/installation.mdx",
		"docs/content/docs/fa/guide/installation.mdx",
		"docs/content/docs/ru/guide/installation.mdx",
	} {
		text := readRepoFile(t, root, name)
		if strings.Contains(text, forbidden) {
			t.Fatalf("%s still tells users to run %s", name, forbidden)
		}
	}
}

func readRepoFile(t *testing.T, root, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(body)
}

func TestInstallScriptsDoNotDownloadUpstreamReleases(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	for _, name := range []string{"install.sh", "update.sh", "x-ui.sh"} {
		text := readRepoFile(t, root, name)
		if strings.Contains(text, "MHSanaei/3x-ui") || strings.Contains(text, "mhsanaei/3x-ui") {
			t.Fatalf("%s still fetches releases or scripts from MHSanaei/3x-ui", name)
		}
		if !strings.Contains(text, "yobo2u/3x-ui") {
			t.Fatalf("%s no longer names a panel repository", name)
		}
	}
}
