package forksync

import "testing"

func TestSelectUnsyncedSkipsWhenEveryReleaseIsAlreadyInHistory(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: "new"},
		{Tag: "v3.10.0", Commit: "old"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return true }, func(string) bool { return false })
	if ok {
		t.Fatalf("selected %s; an ancestor release must not open a sync PR", tag)
	}
}

func TestSelectUnsyncedPicksOldestReleaseMissingFromHistory(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: "new"},
		{Tag: "v3.10.0", Commit: "mid"},
		{Tag: "v3.9.0", Commit: "old"},
	}
	inHistory := map[string]bool{"old": true}
	tag, ok := SelectUnsynced(releases, func(commit string) bool { return inHistory[commit] }, func(string) bool { return false })
	if !ok || tag != "v3.10.0" {
		t.Fatalf("selected %q ok=%v, want oldest missing tag v3.10.0", tag, ok)
	}
}

func TestSelectUnsyncedWaitsWhileOlderSyncBranchExists(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: "new"},
		{Tag: "v3.10.0", Commit: "mid"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return false }, func(branch string) bool {
		return branch == "sync/upstream-v3.10.0"
	})
	if ok {
		t.Fatalf("selected %s while sync/upstream-v3.10.0 is still open", tag)
	}
}

func TestSelectUnsyncedIgnoresDraftsPrereleasesAndNonStableTags(t *testing.T) {
	releases := []Release{
		{Tag: "dev-latest", Commit: "rolling", Prerelease: true},
		{Tag: "v9.0.0", Commit: "draft", Draft: true},
		{Tag: "nightly", Commit: "not-a-release"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return false }, func(string) bool { return false })
	if ok {
		t.Fatalf("selected %s from a list with no stable release", tag)
	}
}
