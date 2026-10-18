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

func TestInstallScriptsDoNotDownloadUpstreamReleases(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	for _, name := range []string{"install.sh", "update.sh", "x-ui.sh"} {
		body, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(body)
		if strings.Contains(text, "MHSanaei/3x-ui") || strings.Contains(text, "mhsanaei/3x-ui") {
			t.Fatalf("%s still fetches releases or scripts from MHSanaei/3x-ui", name)
		}
		if !strings.Contains(text, "yobo2u/3x-ui") {
			t.Fatalf("%s no longer names a panel repository", name)
		}
	}
}
