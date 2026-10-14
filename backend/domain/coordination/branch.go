package coordination

import (
	"strings"
)

// CanonicalBranch strips a leading assignee prefix (ek/feat/x → feat/x). Trunk names stay as-is.
func CanonicalBranch(ownerAlias, branch string) string {
	b := strings.TrimSpace(branch)
	if b == "" {
		return ""
	}
	low := strings.ToLower(b)
	if low == "main" || low == "master" {
		return b
	}
	owner := strings.ToLower(strings.TrimSpace(ownerAlias))
	if owner == "" {
		return b
	}
	prefix := owner + "/"
	if strings.HasPrefix(strings.ToLower(b), prefix) {
		return b[len(prefix):]
	}
	return b
}

// BranchCandidates lists git branch names to try for a slot (exact, then alias/canonical).
func BranchCandidates(ownerAlias, branch string) []string {
	b := strings.TrimSpace(branch)
	if b == "" {
		return nil
	}
	out := []string{b}
	canonical := CanonicalBranch(ownerAlias, b)
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

// BranchMatches reports whether gitBranch satisfies the slot branch for ownerAlias.
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
