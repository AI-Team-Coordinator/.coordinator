package repository

import (
	"context"
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

func isNonFastForward(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not possible to fast-forward") ||
		strings.Contains(msg, "diverging") ||
		strings.Contains(msg, "non-fast-forward")
}
