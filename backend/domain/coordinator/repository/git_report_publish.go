package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	ctx := context.Background()
	abs := filepath.Join(r.progressDir(), ".current_task_"+alias)
	rel, err := filepath.Rel(r.busPath, abs)
	if err != nil {
		rel = filepath.ToSlash(filepath.Join("data", "progress", ".current_task_"+alias))
	}
	rel = filepath.ToSlash(rel)

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
	if err := r.ensureOnlySnapshotDirty(ctx, rel); err != nil {
		return err
	}
	if err := writeGitReport(abs, report); err != nil {
		return err
	}
	if _, err := r.gitOutput(ctx, 15*time.Second, "add", "--", rel); err != nil {
		return gitErr("git_report add", err)
	}
	if _, err := r.gitOutput(ctx, 20*time.Second, "commit", "-m", fmt.Sprintf("chore(progress): %s git_report", alias), "--", rel); err != nil {
		if !gitNothingToCommit(err) {
			return gitErr("git_report commit", err)
		}
		return nil
	}
	if _, err := r.gitOutput(ctx, 60*time.Second, "pull", "--rebase", "origin", "main"); err != nil {
		_, _ = r.gitOutput(context.Background(), 15*time.Second, "rebase", "--abort")
		return gitErr("git_report rebase", err)
	}
	if _, err := r.gitOutput(ctx, 60*time.Second, "push", "origin", "main"); err != nil {
		return gitErr("git_report push", err)
	}
	infra.LogInfo("git_report pushed for %s", alias)
	return nil
}

func (r *FileRepository) ensureOnlySnapshotDirty(ctx context.Context, rel string) error {
	out, err := r.gitOutput(ctx, 10*time.Second, "status", "--porcelain")
	if err != nil {
		return gitErr("git_report status", err)
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		path := line
		if len(path) >= 3 {
			path = strings.TrimSpace(path[2:])
		}
		path = strings.Trim(path, "\"")
		if i := strings.LastIndex(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		path = filepath.ToSlash(path)
		if path != rel {
			return fmt.Errorf("git_report skipped: Common has other local changes (%s)", path)
		}
	}
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

func gitNothingToCommit(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nothing to commit") || strings.Contains(msg, "no changes added")
}
