package forksync

import "regexp"

// stableTag is a published x.y.z release. Rolling tags such as dev-latest are
// not merged automatically; they move on every upstream commit.
var stableTag = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

// Release is one upstream GitHub release reduced to the fields the selector needs.
type Release struct {
	Tag        string
	Commit     string
	Draft      bool
	Prerelease bool
}

// SyncBranch is the fork branch opened for one upstream release tag.
func SyncBranch(tag string) string {
	return "sync/upstream-" + tag
}

// SelectUnsynced returns the oldest stable release missing from fork history.
// releases is newest-first. An older open sync branch blocks a newer PR.
func SelectUnsynced(releases []Release, ancestor func(commit string) bool, branchExists func(branch string) bool) (string, bool) {
	for i := len(releases) - 1; i >= 0; i-- {
		rel := releases[i]
		if rel.Draft || rel.Prerelease || rel.Commit == "" || !stableTag.MatchString(rel.Tag) {
			continue
		}
		if ancestor(rel.Commit) {
			continue
		}
		if branchExists(SyncBranch(rel.Tag)) {
			return "", false
		}
		return rel.Tag, true
	}
	return "", false
}
