package service

import (
	"fmt"
	"strings"
)

type GitSyncError struct {
	Msg    string
	Output string
	Kind   string
}

func (e *GitSyncError) Error() string {
	if e.Output == "" {
		return e.Msg
	}
	return fmt.Sprintf("%s: %s", e.Msg, truncate(e.Output, 800))
}

type GitHubOpError struct {
	Kind string
	Msg  string
}

func (e *GitHubOpError) Error() string {
	return e.Msg
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
