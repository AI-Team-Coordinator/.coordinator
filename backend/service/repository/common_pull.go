package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const commonPullTimeout = 45 * time.Second

// PullCommonOrigin fast-forwards Common main from origin when the tree is clean.
// Returns true when HEAD moved. Skips (no error) if dirty, not on main, or not a fast-forward.
func (r *FileRepository) PullCommonOrigin(ctx context.Context) (bool, error) {
	r.gitMu.Lock()
	defer r.gitMu.Unlock()

	if err := r.ensureGitRepo(ctx); err != nil {
		return false, err
	}

	branch, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(branch) != "main" {
		return false, nil
	}

	status, err := r.gitOutput(ctx, 10*time.Second, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(status) != "" {
		return false, nil
	}

	before, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "HEAD")
	if err != nil {
		return false, err
	}

	if _, err := r.gitOutput(ctx, commonPullTimeout, "fetch", "origin", "main"); err != nil {
		return false, err
	}
	if _, err := r.gitOutput(ctx, 20*time.Second, "merge", "--ff-only", "FETCH_HEAD"); err != nil {
		if isNonFastForward(err) {
			return false, nil
		}
		return false, err
	}

	after, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "HEAD")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(before) != strings.TrimSpace(after), nil
}

// PullCoordinatorState copies origin/coordinator-state into Common/data/progress
// without switching Common off main.
func (r *FileRepository) PullCoordinatorState(ctx context.Context) error {
	r.gitMu.Lock()
	defer r.gitMu.Unlock()
	return r.execCoordinatorState(ctx, "pull")
}

const docsSyncQuietDefault = 30 * time.Second

// docsSyncQuiet is how long a file under docs/ must sit unchanged before background sync commits it.
// Tests set this to zero.
var docsSyncQuiet = docsSyncQuietDefault

// CommonDocsSync is what one background pass did to Common main.
type CommonDocsSync struct {
	Committed bool
	Pulled    bool
	Pushed    bool
}

// SyncCommonDocs keeps Common docs on origin/main.
// It commits only paths under docs/ that have been quiet, fast-forwards or rebases main when the
// tree allows it, then pushes. Files outside docs/ are left untouched. Not on main, or mid-merge, is a no-op.
func (r *FileRepository) SyncCommonDocs(ctx context.Context) (CommonDocsSync, error) {
	r.gitMu.Lock()
	defer r.gitMu.Unlock()
	return r.syncCommonDocsLocked(ctx)
}

func (r *FileRepository) syncCommonDocsLocked(ctx context.Context) (CommonDocsSync, error) {
	var res CommonDocsSync
	if err := r.ensureGitRepo(ctx); err != nil {
		return res, err
	}
	branch, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return res, err
	}
	if strings.TrimSpace(branch) != "main" {
		return res, nil
	}
	if r.midMergeOrRebase(ctx) {
		return res, nil
	}
	if _, err := r.gitOutput(ctx, commonPullTimeout, "fetch", "origin", "main"); err != nil {
		return res, err
	}

	committed, err := r.commitSettledDocs(ctx)
	if err != nil {
		return res, err
	}
	res.Committed = committed

	behind, err := r.revCount(ctx, "HEAD..origin/main")
	if err != nil {
		return res, err
	}
	ahead, err := r.revCount(ctx, "origin/main..HEAD")
	if err != nil {
		return res, err
	}
	if behind > 0 {
		clean, err := r.worktreeClean(ctx)
		if err != nil {
			return res, err
		}
		if !clean {
			return res, nil
		}
		before, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "HEAD")
		if err != nil {
			return res, err
		}
		if ahead == 0 {
			if _, err := r.gitOutput(ctx, 20*time.Second, "merge", "--ff-only", "origin/main"); err != nil {
				if isNonFastForward(err) {
					return res, nil
				}
				return res, err
			}
		} else if _, err := r.gitOutput(ctx, 60*time.Second, "rebase", "origin/main"); err != nil {
			_, _ = r.gitOutput(ctx, 20*time.Second, "rebase", "--abort")
			return res, nil
		}
		after, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "HEAD")
		if err != nil {
			return res, err
		}
		if strings.TrimSpace(before) != strings.TrimSpace(after) {
			res.Pulled = true
		}
		ahead, err = r.revCount(ctx, "origin/main..HEAD")
		if err != nil {
			return res, err
		}
	}
	if ahead > 0 {
		if _, err := r.gitOutput(ctx, commonPullTimeout, "push", "origin", "HEAD:main"); err != nil {
			return res, err
		}
		res.Pushed = true
	}
	return res, nil
}

func (r *FileRepository) commitSettledDocs(ctx context.Context) (bool, error) {
	paths, err := r.settledDocPaths(ctx)
	if err != nil || len(paths) == 0 {
		return false, err
	}
	index, err := os.CreateTemp("", "docs-sync-index-*")
	if err != nil {
		return false, err
	}
	indexPath := index.Name()
	_ = index.Close()
	defer os.Remove(indexPath)

	env := []string{"GIT_INDEX_FILE=" + indexPath}
	if _, err := r.gitOutputEnv(ctx, 10*time.Second, env, "read-tree", "HEAD"); err != nil {
		return false, err
	}
	addArgs := append([]string{"add", "--"}, paths...)
	if _, err := r.gitOutputEnv(ctx, 20*time.Second, env, addArgs...); err != nil {
		return false, err
	}
	diff, err := r.gitOutputEnv(ctx, 10*time.Second, env, "diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	for _, name := range strings.Split(diff, "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if !docPathOK(name) {
			return false, fmt.Errorf("refusing to sync %q outside docs/", name)
		}
	}
	if strings.TrimSpace(diff) == "" {
		return false, nil
	}
	if _, err := r.gitOutputEnv(ctx, 20*time.Second, env, "commit", "-m", "docs: sync working copy"); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "nothing to commit") || strings.Contains(msg, "no changes added") {
			return false, nil
		}
		return false, err
	}
	// The side index moved HEAD. Point the real index at those paths so they are not left as a staged revert.
	restore := append([]string{"checkout", "HEAD", "--"}, paths...)
	if _, err := r.gitOutput(ctx, 20*time.Second, restore...); err != nil {
		return false, err
	}
	return true, nil
}

func (r *FileRepository) settledDocPaths(ctx context.Context) ([]string, error) {
	out, err := r.gitOutput(ctx, 10*time.Second, "status", "--porcelain", "--untracked-files=all", "--", "docs/")
	if err != nil {
		return nil, err
	}
	var paths []string
	seen := map[string]struct{}{}
	for _, line := range strings.Split(out, "\n") {
		path := porcelainPath(line)
		if !docPathOK(path) {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		full := filepath.Join(r.busPath, filepath.FromSlash(path))
		info, statErr := os.Stat(full)
		if statErr == nil && docsSyncQuiet > 0 && time.Since(info.ModTime()) < docsSyncQuiet {
			continue
		}
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	return paths, nil
}

func porcelainPath(line string) string {
	line = strings.TrimRight(line, "\r")
	if len(line) < 4 {
		return ""
	}
	rest := strings.TrimSpace(line[2:])
	if i := strings.LastIndex(rest, " -> "); i >= 0 {
		rest = strings.TrimSpace(rest[i+4:])
	}
	return strings.Trim(rest, "\"")
}

func docPathOK(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	if path == "" || strings.Contains(path, "..") || !strings.HasPrefix(path, "docs/") {
		return false
	}
	base := filepath.Base(path)
	if base == ".current_author" || strings.Contains(base, ".env") {
		return false
	}
	return true
}

func (r *FileRepository) revCount(ctx context.Context, rangeSpec string) (int, error) {
	out, err := r.gitOutput(ctx, 10*time.Second, "rev-list", "--count", rangeSpec)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

func (r *FileRepository) worktreeClean(ctx context.Context) (bool, error) {
	out, err := r.gitOutput(ctx, 10*time.Second, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

func (r *FileRepository) midMergeOrRebase(ctx context.Context) bool {
	for _, rel := range []string{"MERGE_HEAD", "rebase-merge", "rebase-apply"} {
		out, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "--git-path", rel)
		if err != nil {
			continue
		}
		path := strings.TrimSpace(out)
		if path == "" {
			continue
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(r.busPath, path)
		}
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func isNonFastForward(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not possible to fast-forward") ||
		strings.Contains(msg, "diverging") ||
		strings.Contains(msg, "non-fast-forward")
}
