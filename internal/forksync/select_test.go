package forksync

import "testing"

func TestSelectUnsyncedSkipsWhenLatestIsAlreadyInHistory(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: "new"},
		{Tag: "v3.10.0", Commit: "old"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return true }, func(string) bool { return false })
	if ok {
		t.Fatalf("selected %s; an ancestor release must not open a sync PR", tag)
	}
}

func TestSelectUnsyncedPicksLatestReleaseMissingFromHistory(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: "new"},
		{Tag: "v3.10.0", Commit: "mid"},
		{Tag: "v3.9.0", Commit: "old"},
	}
	inHistory := map[string]bool{"old": true}
	tag, ok := SelectUnsynced(releases, func(commit string) bool { return inHistory[commit] }, func(string) bool { return false })
	if !ok || tag != "v3.10.2" {
		t.Fatalf("selected %q ok=%v, want latest missing tag v3.10.2", tag, ok)
	}
}

func TestSelectUnsyncedSkipsOlderRewrittenTagsWhenLatestIsInHistory(t *testing.T) {
	// Production failure: rewritten pre-2.5 tags are not ancestors of HEAD, but
	// v3.10.2 is. Walking oldest-first selected v1.7.6 and tried to merge it.
	releases := []Release{
		{Tag: "dev-latest", Commit: "rolling", Prerelease: true},
		{Tag: "v3.10.2", Commit: "current"},
		{Tag: "v3.10.0", Commit: "recent"},
		{Tag: "v2.4.4", Commit: "rewritten-new"},
		{Tag: "v1.7.6", Commit: "rewritten-old"},
	}
	inHistory := map[string]bool{"current": true, "recent": true}
	tag, ok := SelectUnsynced(releases, func(commit string) bool { return inHistory[commit] }, func(string) bool { return false })
	if ok {
		t.Fatalf("selected %s; latest stable is already in history so no PR", tag)
	}
}

func TestSelectUnsyncedWaitsWhileLatestSyncBranchExists(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: "new"},
		{Tag: "v3.10.0", Commit: "mid"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return false }, func(branch string) bool {
		return branch == "sync/upstream-v3.10.2"
	})
	if ok {
		t.Fatalf("selected %s while sync/upstream-v3.10.2 is still open", tag)
	}
}

func TestSelectUnsyncedDoesNotWaitOnOlderSyncBranch(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: "new"},
		{Tag: "v1.7.6", Commit: "old"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return false }, func(branch string) bool {
		return branch == "sync/upstream-v1.7.6"
	})
	if !ok || tag != "v3.10.2" {
		t.Fatalf("selected %q ok=%v, want v3.10.2 even if an ancient sync branch exists", tag, ok)
	}
}

func TestSelectUnsyncedPicksLatestStableBehindDraftNotRewrittenTag(t *testing.T) {
	releases := []Release{
		{Tag: "v3.12.0", Commit: "draft", Draft: true},
		{Tag: "v3.12.0-rc.1", Commit: "rc", Prerelease: true},
		{Tag: "v3.11.0", Commit: "stable"},
		{Tag: "v1.7.6", Commit: "rewritten"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return false }, func(string) bool { return false })
	if !ok || tag != "v3.11.0" {
		t.Fatalf("selected %q ok=%v, want v3.11.0; a newer draft must not hide it and v1.7.6 must not be chosen", tag, ok)
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

func TestSelectUnsyncedDoesNotFallThroughWhenLatestTagCommitIsUnknown(t *testing.T) {
	releases := []Release{
		{Tag: "v3.10.2", Commit: ""},
		{Tag: "v1.7.6", Commit: "old"},
	}
	tag, ok := SelectUnsynced(releases, func(string) bool { return false }, func(string) bool { return false })
	if ok {
		t.Fatalf("selected %s after the latest stable tag failed to resolve", tag)
	}
}
