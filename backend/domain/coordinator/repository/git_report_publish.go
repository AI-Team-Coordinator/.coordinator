package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"coordinator/infra"
	"coordinator/model"
)

const gitPublishInterval = 2 * time.Minute

func (r *FileRepository) maybePublishGitReport(author string, members []model.Member) {
	if author == "" {
		return
	}
	var self *model.Member
	for i := range members {
		if members[i].Alias == author {
			self = &members[i]
			break
		}
	}
	if self == nil || self.Status != "in_progress" {
		return
	}
	next := buildGitReportFromRepos(self.Repos, time.Now())
	if gitReportEqual(self.GitReport, next) {
		return
	}
	if !r.gitPublishRun.CompareAndSwap(false, true) {
		return
	}
	hadStored := self.GitReport != nil && len(self.GitReport.Repos) > 0
	go func(alias string, report *model.GitReport, debounce bool) {
		defer r.gitPublishRun.Store(false)
		r.gitPublishMu.Lock()
		defer r.gitPublishMu.Unlock()
		if debounce && !r.gitPublishAt.IsZero() && time.Since(r.gitPublishAt) < gitPublishInterval {
			return
		}
		if err := r.publishGitReport(alias, report); err != nil {
			infra.LogWarn("git_report publish: %v", err)
			return
		}
		r.gitPublishAt = time.Now()
	}(author, next, hadStored)
}

func (r *FileRepository) publishGitReport(alias string, report *model.GitReport) error {
	abs := filepath.Join(r.progressDir(), ".current_task_"+alias)
	if r.Collaboration() == model.CollaborationSolo {
		return writeGitReport(abs, report)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	r.gitMu.Lock()
	defer r.gitMu.Unlock()

	if err := r.ensureGitRepo(ctx); err != nil {
		return err
	}
	branch, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	if strings.TrimSpace(branch) != "main" {
		return fmt.Errorf("git_report skipped: Common is on %s", strings.TrimSpace(branch))
	}
	if err := writeGitReport(abs, report); err != nil {
		return err
	}
	if err := r.execCoordinatorState(ctx, "push", fmt.Sprintf("chore(progress): %s git_report", alias)); err != nil {
		return err
	}
	infra.LogInfo("git_report pushed for %s", alias)
	return nil
}

func writeGitReport(path string, report *model.GitReport) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var snap map[string]any
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	status, _ := snap["status"].(string)
	if status != "in_progress" {
		return fmt.Errorf("git_report skipped: status %s", status)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		return err
	}
	var reportVal any
	if err := json.Unmarshal(encoded, &reportVal); err != nil {
		return err
	}
	snap["git_report"] = reportVal
	out, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(path, out, 0o644)
}

func (r *FileRepository) execCoordinatorState(ctx context.Context, args ...string) error {
	script := r.appScript("coordinator_state.sh")
	if script == "" || !fileExists(script) {
		return fmt.Errorf("coordinator_state.sh not found")
	}
	cmd := exec.CommandContext(ctx, script, args...)
	cmd.Dir = r.appDir()
	cmd.Env = append(os.Environ(), "COORDINATOR_ROOT="+r.appDir(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("coordinator_state %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}
