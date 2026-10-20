package slot

import (
	"strings"

	"coordinator/domain/task"
)

// BranchCandidates lists git names for this slot: as stored, the task name, and alias/task name.
func BranchCandidates(ownerAlias, branch string) []string {
	b := strings.TrimSpace(branch)
	if b == "" {
		return nil
	}
	out := []string{b}
	canonical := task.CanonicalBranch(ownerAlias, b)
	if canonical != b && canonical != "" {
		out = append(out, canonical)
	}
	owner := strings.ToLower(strings.TrimSpace(ownerAlias))
	if owner != "" && canonical != "" {
		prefixed := owner + "/" + canonical
		if prefixed != b && prefixed != canonical {
			out = append(out, prefixed)
		}
	}
	return out
}

// BranchMatches reports whether the checkout satisfies the slot branch for this owner.
func BranchMatches(ownerAlias, slotBranch, gitBranch string) bool {
	gitBranch = strings.TrimSpace(gitBranch)
	if gitBranch == "" {
		return false
	}
	for _, want := range BranchCandidates(ownerAlias, slotBranch) {
		if gitBranch == want {
			return true
		}
	}
	return false
}
