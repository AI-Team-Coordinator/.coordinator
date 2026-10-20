package task

import "strings"

// CanonicalBranch is the name stored on the task, without an assignee prefix.
// Trunk names stay as written. ek/feat/voice for owner EK becomes feat/voice.
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
	if strings.HasPrefix(low, prefix) {
		return b[len(prefix):]
	}
	return b
}
