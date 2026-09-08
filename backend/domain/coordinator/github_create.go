package coordinator

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"coordinator/model"
)

func createOrgRepository(ctx context.Context, org, slug, description string) error {
	gh := lookPathGH()
	if gh == "" {
		return &ValidationError{Msg: "gh CLI is not installed"}
	}

	args := []string{"repo", "create", org + "/" + slug, "--private", "--add-readme"}
	if description != "" {
		args = append(args, "--description", description)
	}

	_, err := ghOutputTimeout(ctx, 45*time.Second, gh, args...)
	if err != nil {
		msg := err.Error()
		lower := strings.ToLower(msg)
		if strings.Contains(lower, "already exists") || strings.Contains(lower, "name already exists") {
			return &GitHubOpError{Kind: "exists", Msg: "GitHub repository already exists in the organization"}
		}
		return &GitHubOpError{Kind: "failed", Msg: "failed to create GitHub repository: " + msg}
	}
	return nil
}

func cloneOrgRepository(ctx context.Context, binding model.GitHubBinding, slug, dest string) error {
	git := lookPathGit()
	if git == "" {
		return &GitHubOpError{Kind: "failed", Msg: "git is not installed"}
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, git, "clone", "--", sshCloneURL(binding, slug), dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return &GitHubOpError{Kind: "failed", Msg: "clone failed: " + truncate(msg, 400)}
	}
	return nil
}

func lookPathGit() string {
	if p, err := exec.LookPath("git"); err == nil {
		return p
	}
	for _, p := range []string{"/opt/homebrew/bin/git", "/usr/local/bin/git"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func ghOutputTimeout(ctx context.Context, timeout time.Duration, gh string, args ...string) (string, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	return ghOutput(ctx, gh, args...)
}

func sshCloneURL(binding model.GitHubBinding, slug string) string {
	host := strings.TrimSpace(binding.SSHHost)
	if host == "" {
		host = "github.com"
	}
	return "git@" + host + ":" + strings.TrimSpace(binding.Org) + "/" + slug + ".git"
}
