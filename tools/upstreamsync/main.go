package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mhsanaei/3x-ui/v3/internal/forksync"
)

// Reads TSV rows: tag, commit, draft, prerelease, ancestor, branchExists.
// Prints the selected tag, or nothing when no sync PR should be opened.
func main() {
	releases, ancestor, branchExists, err := readRows(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	tag, ok := forksync.SelectUnsynced(releases, func(commit string) bool {
		return ancestor[commit]
	}, func(branch string) bool {
		return branchExists[branch]
	})
	if ok {
		fmt.Println(tag)
	}
}

func readRows(r *os.File) ([]forksync.Release, map[string]bool, map[string]bool, error) {
	var releases []forksync.Release
	ancestor := map[string]bool{}
	branchExists := map[string]bool{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 6 {
			return nil, nil, nil, fmt.Errorf("upstreamsync: want 6 tab fields, got %q", line)
		}
		draft, err1 := strconv.ParseBool(fields[2])
		pre, err2 := strconv.ParseBool(fields[3])
		anc, err3 := strconv.ParseBool(fields[4])
		exists, err4 := strconv.ParseBool(fields[5])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			return nil, nil, nil, fmt.Errorf("upstreamsync: bad bools in %q", line)
		}
		releases = append(releases, forksync.Release{
			Tag: fields[0], Commit: fields[1], Draft: draft, Prerelease: pre,
		})
		ancestor[fields[1]] = anc
		branchExists[forksync.SyncBranch(fields[0])] = exists
	}
	if err := sc.Err(); err != nil {
		return nil, nil, nil, err
	}
	return releases, ancestor, branchExists, nil
}
