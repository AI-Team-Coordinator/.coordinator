package statebus

import (
	"testing"
)

func TestStateGitURLFromEnvPrefersStateGitURL(t *testing.T) {
	t.Setenv("RELAY_STATE_GIT_URL", "git@example.com:team/bus.git")
	t.Setenv("RELAY_COMMON_GIT_URL", "git@example.com:legacy/common.git")
	if got := stateGitURLFromEnv(); got != "git@example.com:team/bus.git" {
		t.Fatalf("got %q", got)
	}
}

func TestStateGitURLFromEnvLegacyCommon(t *testing.T) {
	t.Setenv("RELAY_STATE_GIT_URL", "")
	t.Setenv("RELAY_COMMON_GIT_URL", "git@example.com:legacy/common.git")
	if got := stateGitURLFromEnv(); got != "git@example.com:legacy/common.git" {
		t.Fatalf("got %q", got)
	}
}
