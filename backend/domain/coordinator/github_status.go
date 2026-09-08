package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"coordinator/domain/coordinator/dto"
)

func probeGitHubCLI(ctx context.Context, org string) *dto.GitHubLiveResponse {
	org = strings.TrimSpace(org)
	if org == "" {
		return nil
	}

	gh := lookPathGH()
	if gh == "" {
		return &dto.GitHubLiveResponse{Auth: "missing", Detail: "gh CLI is not installed"}
	}

	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 4*time.Second {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
	}

	login, err := ghOutput(ctx, gh, "api", "user", "-q", ".login")
	if err != nil {
		return &dto.GitHubLiveResponse{Auth: "error", Detail: shortGHError(err)}
	}

	role := ""
	roleJSON, err := ghOutput(ctx, gh, "api", "user/memberships/orgs/"+org)
	if err == nil {
		var body struct {
			Role string `json:"role"`
		}
		if json.Unmarshal([]byte(roleJSON), &body) == nil {
			role = body.Role
		}
	}

	return &dto.GitHubLiveResponse{
		Auth:  "ok",
		Login: login,
		Role:  role,
	}
}

func lookPathGH() string {
	if p, err := exec.LookPath("gh"); err == nil {
		return p
	}
	for _, p := range []string{"/opt/homebrew/bin/gh", "/usr/local/bin/gh"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func ghOutput(ctx context.Context, gh string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, gh, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return "", err
		}
		if line := strings.Split(msg, "\n")[0]; line != "" {
			return "", fmt.Errorf("%s", truncate(msg, 400))
		}
		return "", err
	}
	return out, nil
}

func shortGHError(err error) string {
	msg := err.Error()
	if idx := strings.LastIndex(msg, ": "); idx >= 0 && idx+2 < len(msg) {
		msg = msg[idx+2:]
	}
	if len(msg) > 180 {
		return msg[:180]
	}
	return msg
}
