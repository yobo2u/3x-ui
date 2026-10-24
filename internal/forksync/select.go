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

// SelectUnsynced returns the latest stable release that still needs a sync PR.
// releases is newest-first; a current ancestor or existing sync branch means none.
func SelectUnsynced(releases []Release, ancestor func(commit string) bool, branchExists func(branch string) bool) (string, bool) {
	for _, rel := range releases {
		if rel.Draft || rel.Prerelease || !stableTag.MatchString(rel.Tag) {
			continue
		}
		if rel.Commit == "" {
			return "", false
		}
		if ancestor(rel.Commit) || branchExists(SyncBranch(rel.Tag)) {
			return "", false
		}
		return rel.Tag, true
	}
	return "", false
}
