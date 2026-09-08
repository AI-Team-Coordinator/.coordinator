package repository

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"coordinator/model"
)

var settingsSyncFiles = []string{
	"data/settings/team.json",
	"data/settings/project_profile.json",
	"data/settings/author.md",
}

func (r *FileRepository) CurrentAuthor(_ context.Context) (string, error) {
	candidates := []string{
		filepath.Join(r.dataDir(), ".current_author"),
		filepath.Join(r.cursorPath, ".current_author"),
	}
	for _, path := range candidates {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if alias := parseCurrentAuthor(data); alias != "" {
			return alias, nil
		}
	}
	return "", nil
}

func (r *FileRepository) SettingsSyncStatus(ctx context.Context) (*model.SyncStatus, error) {
	alias, err := r.CurrentAuthor(ctx)
	if err != nil {
		return nil, err
	}

	branch, _ := r.gitOutput(ctx, 5*time.Second, "rev-parse", "--abbrev-ref", "HEAD")
	dirty, err := r.dirtyWhitelistFiles(ctx)
	if err != nil {
		return nil, err
	}

	return &model.SyncStatus{
		Alias:      alias,
		Branch:     strings.TrimSpace(branch),
		DirtyFiles: dirty,
	}, nil
}

func (r *FileRepository) SyncSettings(ctx context.Context, alias string) (*model.SyncResult, error) {
	r.gitMu.Lock()
	defer r.gitMu.Unlock()
	if err := r.ensureGitRepo(ctx); err != nil {
		return nil, err
	}

	branch, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, gitErr("failed to read current branch", err)
	}
	if strings.TrimSpace(branch) != "main" {
		return nil, gitErr("sync is only allowed on main (current branch: "+strings.TrimSpace(branch)+")", nil)
	}

	existing := make([]string, 0, len(settingsSyncFiles))
	for _, name := range settingsSyncFiles {
		if _, err := os.Stat(filepath.Join(r.busPath, name)); err == nil {
			existing = append(existing, name)
		}
	}
	if len(existing) == 0 {
		return nil, fmt.Errorf("no whitelisted files to sync")
	}

	_, _ = r.gitOutput(ctx, 15*time.Second, "reset", "-q", "HEAD")
	if _, err := r.gitOutput(ctx, 15*time.Second, append([]string{"add", "--"}, existing...)...); err != nil {
		return nil, gitErr("failed to stage files", err)
	}

	staged, err := r.stagedWhitelistFiles(ctx)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("chore(coordinator): %s sync project settings", alias)
	committed := false
	if len(staged) > 0 {
		if _, err := r.gitOutput(ctx, 20*time.Second, "commit", "-m", message); err != nil {
			return nil, gitErr("failed to commit", err)
		}
		committed = true
	}

	if _, err := r.gitOutput(ctx, 60*time.Second, "pull", "--rebase", "origin", "main"); err != nil {
		_, _ = r.gitOutput(context.Background(), 15*time.Second, "rebase", "--abort")
		return nil, gitConflict("rebase onto origin/main failed", err)
	}

	if _, err := r.gitOutput(ctx, 60*time.Second, "push", "origin", "main"); err != nil {
		return nil, gitErr("failed to push origin/main", err)
	}

	files := staged
	if files == nil {
		files = []string{}
	}

	return &model.SyncResult{
		Alias:     alias,
		Committed: committed,
		Pushed:    true,
		Files:     files,
		Message:   message,
	}, nil
}

func (r *FileRepository) ensureGitRepo(ctx context.Context) error {
	out, err := r.gitOutput(ctx, 5*time.Second, "rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(out) != "true" {
		return gitErr("Common is not a git repository", err)
	}
	return nil
}

func (r *FileRepository) dirtyWhitelistFiles(ctx context.Context) ([]string, error) {
	if err := r.ensureGitRepo(ctx); err != nil {
		return nil, err
	}
	out, err := r.gitOutput(ctx, 10*time.Second, append([]string{"status", "--porcelain", "--"}, settingsSyncFiles...)...)
	if err != nil {
		return nil, gitErr("failed to read git status", err)
	}
	files := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
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
		if path != "" {
			files = append(files, filepath.ToSlash(path))
		}
	}
	return files, nil
}

func (r *FileRepository) stagedWhitelistFiles(ctx context.Context) ([]string, error) {
	out, err := r.gitOutput(ctx, 10*time.Second, "diff", "--cached", "--name-only", "--")
	if err != nil {
		return nil, gitErr("failed to list staged files", err)
	}
	allowed := make(map[string]struct{}, len(settingsSyncFiles))
	for _, name := range settingsSyncFiles {
		allowed[name] = struct{}{}
	}
	files := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		name := filepath.ToSlash(strings.TrimSpace(line))
		if name == "" {
			continue
		}
		if _, ok := allowed[name]; ok {
			files = append(files, name)
		}
	}
	return files, nil
}

func (r *FileRepository) gitOutput(ctx context.Context, timeout time.Duration, args ...string) (string, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.busPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	if err != nil {
		if out == "" {
			return "", err
		}
		return out, fmt.Errorf("%w: %s", err, out)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func gitErr(msg string, err error) error {
	out := ""
	if err != nil {
		out = err.Error()
	}
	return &GitFailure{Msg: msg, Output: out, Kind: "error"}
}

func gitConflict(msg string, err error) error {
	out := ""
	if err != nil {
		out = err.Error()
	}
	return &GitFailure{Msg: msg, Output: out, Kind: "conflict"}
}

type GitFailure struct {
	Msg    string
	Output string
	Kind   string
}

func (e *GitFailure) Error() string {
	if e.Output == "" {
		return e.Msg
	}
	return e.Msg + ": " + e.Output
}
