package statebus

import (
	"os"
	"strings"
)

// Config drives git coordinator-state writes from relay.
type Config struct {
	Enabled   bool
	OriginID  string
	Branch    string
	Worktree  string
	GitURL    string
	gitEnv    []string
}

func LoadFromEnv() Config {
	enabled := strings.TrimSpace(os.Getenv("RELAY_STATE_ENABLED"))
	on := enabled == "1" || strings.EqualFold(enabled, "true")
	origin := strings.TrimSpace(os.Getenv("RELAY_ORIGIN_ID"))
	if origin == "" {
		origin = strings.TrimSpace(os.Getenv("RELAY_INSTANCE"))
	}
	if origin == "" {
		origin = "default"
	}
	branch := strings.TrimSpace(os.Getenv("RELAY_STATE_BRANCH"))
	if branch == "" {
		branch = "coordinator-state"
	}
	worktree := strings.TrimSpace(os.Getenv("RELAY_STATE_WORKTREE"))
	if worktree == "" {
		worktree = strings.TrimSpace(os.Getenv("RELAY_STATE_WORKTREE_DIR"))
	}
	return Config{
		Enabled:  on,
		OriginID: origin,
		Branch:   branch,
		Worktree: worktree,
		GitURL:   stateGitURLFromEnv(),
	}
}

// stateGitURLFromEnv is the coordinator-state bus repository (not the local "Common" folder name).
// Prefer RELAY_STATE_GIT_URL; RELAY_COMMON_GIT_URL is deprecated alias (Alina stores state in Common git).
func stateGitURLFromEnv() string {
	if u := strings.TrimSpace(os.Getenv("RELAY_STATE_GIT_URL")); u != "" {
		return u
	}
	return strings.TrimSpace(os.Getenv("RELAY_COMMON_GIT_URL"))
}

func (c Config) Ready() bool {
	return c.Enabled && c.Worktree != "" && c.GitURL != ""
}
