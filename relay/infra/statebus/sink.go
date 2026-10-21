package statebus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"relay/domain/relay/bus"
	"relay/infra"
	"relay/model"
)

// Sink appends ExternalEvent lines under data/progress/external/events/<origin_id>/ and pushes coordinator-state.
type Sink struct {
	cfg Config
	mu  sync.Mutex
}

func NewSink(cfg Config, dataDir string) *Sink {
	if !cfg.Ready() {
		return nil
	}
	extra, err := gitExtraEnv(dataDir)
	if err != nil {
		infra.LogWarn("statebus git ssh: %v (git push disabled until fixed)", err)
	} else {
		cfg.gitEnv = extra
	}
	if cfg.Enabled && !hasGitSSH(cfg.gitEnv) && strings.TrimSpace(os.Getenv("GIT_SSH_COMMAND")) == "" {
		infra.LogWarn("statebus: set RELAY_STATE_GIT_SSH_KEY_B64 (or mount key + RELAY_STATE_GIT_SSH_KEY_PATH); git push may fail")
	}
	return &Sink{cfg: cfg}
}

func hasGitSSH(extra []string) bool {
	for _, e := range extra {
		if strings.HasPrefix(e, "GIT_SSH_COMMAND=") {
			return true
		}
	}
	return false
}

func (s *Sink) Append(ctx context.Context, ev model.ExternalEvent) error {
	if s == nil {
		return fmt.Errorf("state bus not configured")
	}
	line, err := bus.MarshalLine(ev)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureWorktree(ctx); err != nil {
		return err
	}
	year := time.Now().UTC().Format("2006")
	rel := filepath.Join("data", "progress", "external", "events", s.cfg.OriginID, year+".jsonl")
	abs := filepath.Join(s.cfg.Worktree, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(line); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return s.commitAndPush(ctx, rel)
}

func (s *Sink) ensureWorktree(ctx context.Context) error {
	if fileExists(filepath.Join(s.cfg.Worktree, ".git")) {
		return nil
	}
	if err := os.MkdirAll(s.cfg.Worktree, 0o755); err != nil {
		return err
	}
	if err := s.runGit(ctx, s.cfg.Worktree, "clone", "--depth", "1", "--branch", s.cfg.Branch, s.cfg.GitURL, "."); err != nil {
		// branch may not exist yet — clone repo and create orphan branch on first push
		if err2 := s.runGit(ctx, s.cfg.Worktree, "clone", "--depth", "1", s.cfg.GitURL, "."); err2 != nil {
			return fmt.Errorf("clone: %v (branch %v)", err, err2)
		}
		_ = s.runGit(ctx, s.cfg.Worktree, "checkout", "-B", s.cfg.Branch)
	}
	return nil
}

func (s *Sink) commitAndPush(ctx context.Context, relPath string) error {
	wt := s.cfg.Worktree
	_ = s.runGit(ctx, wt, "checkout", s.cfg.Branch)
	if err := s.runGit(ctx, wt, "add", relPath); err != nil {
		return err
	}
	msg := fmt.Sprintf("chore(relay): external event %s", time.Now().UTC().Format(time.RFC3339))
	_ = s.runGit(ctx, wt, "commit", "-m", msg)
	_ = s.runGit(ctx, wt, "pull", "--rebase", "origin", s.cfg.Branch)
	if err := s.runGit(ctx, wt, "push", "-u", "origin", s.cfg.Branch); err != nil {
		infra.LogWarn("statebus push: %v", err)
		return err
	}
	return nil
}

func (s *Sink) runGit(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), s.cfg.gitEnv...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
