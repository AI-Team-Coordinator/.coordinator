package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"coordinator/model"
)

type SetupLayout struct {
	Kind    string
	DocsRel string
	DataRel string
}

func (r *FileRepository) WriteCurrentAuthor(_ context.Context, alias string) error {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil
	}
	body := "# Local identity for this machine. Not committed.\n# First non-empty line that is not a comment is the alias from settings/team.json.\n\n" + alias + "\n"
	return writeFileAtomic(filepath.Join(r.dataDir(), ".current_author"), []byte(body))
}

func (r *FileRepository) GetLocale(_ context.Context) (*model.LocaleFile, error) {
	path := filepath.Join(r.settingsDir(), "locale.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.LocaleFile{Version: 1, ChatLanguage: "en", DocsLanguage: "en"}, nil
		}
		return nil, err
	}
	var loc model.LocaleFile
	if err := json.Unmarshal(data, &loc); err != nil {
		return nil, err
	}
	if loc.ChatLanguage == "" {
		loc.ChatLanguage = "en"
	}
	if loc.DocsLanguage == "" {
		loc.DocsLanguage = loc.ChatLanguage
	}
	if loc.Version == 0 {
		loc.Version = 1
	}
	return &loc, nil
}

// CoordinatorName is the Coordinator's own name: settings/coordinator.json.
// Missing file or empty name falls back to the default for the chat language.
func (r *FileRepository) CoordinatorName(ctx context.Context) string {
	lang := ""
	if loc, err := r.GetLocale(ctx); err == nil && loc != nil {
		lang = loc.ChatLanguage
	}
	name := strings.TrimSpace(r.coordinatorFileFromDisk().Name)
	if name != "" {
		return name
	}
	return DefaultCoordinatorName(lang)
}

// DefaultCoordinatorName is Otto — it rhymes with octopus.
func DefaultCoordinatorName(lang string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(lang)), "ru") {
		return "Отто"
	}
	return "Otto"
}

func (r *FileRepository) SaveLocale(_ context.Context, loc *model.LocaleFile) error {
	if loc == nil {
		return nil
	}
	if loc.Version == 0 {
		loc.Version = 1
	}
	data, err := json.MarshalIndent(loc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(filepath.Join(r.settingsDir(), "locale.json"), data)
}

func (r *FileRepository) coordinatorFileFromDisk() model.CoordinatorFile {
	data, err := os.ReadFile(filepath.Join(r.settingsDir(), "coordinator.json"))
	if err != nil {
		return model.CoordinatorFile{Version: 1}
	}
	var cfg model.CoordinatorFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return model.CoordinatorFile{Version: 1}
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	return cfg
}

func (r *FileRepository) CoordinatorFile(_ context.Context) model.CoordinatorFile {
	return r.coordinatorFileFromDisk()
}

// Collaboration is team unless the file explicitly says solo.
// Missing field keeps existing installs on the git bus.
func (r *FileRepository) Collaboration() string {
	return model.NormalizeCollaboration(r.coordinatorFileFromDisk().Collaboration, model.CollaborationTeam)
}

func (r *FileRepository) SaveCoordinatorFile(_ context.Context, cfg model.CoordinatorFile) error {
	current := r.coordinatorFileFromDisk()
	if strings.TrimSpace(cfg.Name) == "" {
		cfg.Name = current.Name
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	cfg.Collaboration = model.NormalizeCollaboration(cfg.Collaboration, model.CollaborationTeam)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(filepath.Join(r.settingsDir(), "coordinator.json"), data)
}

func (r *FileRepository) SetupLayout() SetupLayout {
	ws := r.WorkspaceDir()
	data := r.dataDir()
	docs := r.docsPath
	kind := "existing"
	if filepath.Base(data) == "coordinator-data" {
		if filepath.Clean(filepath.Dir(data)) == filepath.Clean(ws) {
			kind = "in-repo"
		} else {
			kind = "workspace-parent"
		}
	}
	return SetupLayout{
		Kind:    kind,
		DocsRel: relToWorkspace(ws, docs),
		DataRel: relToWorkspace(ws, data),
	}
}

func (r *FileRepository) ApplySetupFolders(docsRel, dataRel string) error {
	ws := r.WorkspaceDir()
	docsAbs, err := pathInsideWorkspace(ws, docsRel)
	if err != nil {
		return fmt.Errorf("docs folder: %w", err)
	}
	dataAbs, err := pathInsideWorkspace(ws, dataRel)
	if err != nil {
		return fmt.Errorf("data folder: %w", err)
	}
	if err := os.MkdirAll(docsAbs, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dataAbs, "settings"), 0o755); err != nil {
		return err
	}
	r.docsPath = docsAbs
	r.dataPath = dataAbs
	if r.appPath == "" {
		return nil
	}
	return upsertEnvKeys(filepath.Join(r.appPath, ".env"), map[string]string{
		"DOCS_DIR": relToWorkspace(r.appPath, docsAbs),
		"DATA_DIR": relToWorkspace(r.appPath, dataAbs),
	})
}

func pathInsideWorkspace(workspace, raw string) (string, error) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if raw == "" {
		return "", fmt.Errorf("path is required")
	}
	var abs string
	if filepath.IsAbs(raw) {
		abs = filepath.Clean(raw)
	} else {
		abs = filepath.Clean(filepath.Join(workspace, filepath.FromSlash(raw)))
	}
	rel, err := filepath.Rel(workspace, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("must stay inside the project folder")
	}
	return abs, nil
}

func upsertEnvKeys(path string, updates map[string]string) error {
	current, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	seen := make(map[string]bool, len(updates))
	out := make([]string, 0)
	if len(current) > 0 {
		for _, line := range strings.Split(strings.TrimSuffix(string(current), "\n"), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				key, _, ok := strings.Cut(trimmed, "=")
				key = strings.TrimSpace(key)
				if ok {
					if val, hit := updates[key]; hit {
						out = append(out, key+"="+val)
						seen[key] = true
						continue
					}
				}
			}
			out = append(out, line)
		}
	}
	for key, val := range updates {
		if seen[key] {
			continue
		}
		out = append(out, key+"="+val)
	}
	body := strings.Join(out, "\n")
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return writeFileAtomic(path, []byte(body))
}

func (r *FileRepository) GuessProjectName(ctx context.Context) string {
	if name := gitOriginRepoName(ctx, r.WorkspaceDir()); name != "" {
		return name
	}
	if name := packageJSONName(r.WorkspaceDir()); name != "" {
		return name
	}
	base := filepath.Base(r.WorkspaceDir())
	if base == "" || base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

func relToWorkspace(workspace, path string) string {
	if workspace == "" || path == "" {
		return path
	}
	rel, err := filepath.Rel(workspace, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

func gitOriginRepoName(ctx context.Context, dir string) string {
	if dir == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "remote", "get-url", "origin")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return ""
	}
	url := strings.TrimSpace(stdout.String())
	url = strings.TrimSuffix(url, ".git")
	url = strings.ReplaceAll(url, "\\", "/")
	if i := strings.LastIndex(url, "/"); i >= 0 && i+1 < len(url) {
		return url[i+1:]
	}
	if i := strings.LastIndex(url, ":"); i >= 0 && i+1 < len(url) {
		return url[i+1:]
	}
	return ""
}

func packageJSONName(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var file struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return ""
	}
	name := strings.TrimSpace(file.Name)
	name = strings.TrimPrefix(name, "@")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return strings.TrimSpace(name)
}
