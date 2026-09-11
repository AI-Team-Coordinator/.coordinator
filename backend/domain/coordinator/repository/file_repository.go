package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"coordinator/model"
)

// FileRepository reads the data bus (settings, progress snapshots) with a one-release fallback to legacy paths.
type FileRepository struct {
	busPath       string
	dataPath      string
	docsPath      string
	workspace     string
	cursorPath    string
	appPath       string
	gitMu         sync.Mutex
	usageMu       sync.Mutex
	usageSnap     *model.CursorUsage
	usageAt       time.Time
	gitPublishMu  sync.Mutex
	gitPublishAt  time.Time
	gitPublishRun atomic.Bool
}

// Paths locates workspace folders. Names are not assumed — callers pass them from env.
type Paths struct {
	Workspace string
	Data      string
	Docs      string
	Bus       string
	Cursor    string
	App       string
}

func NewFileRepository(busPath, cursorPath string) *FileRepository {
	return NewFileRepositoryWithPaths(Paths{
		Bus:       busPath,
		Cursor:    cursorPath,
		Data:      filepath.Join(busPath, "data"),
		Docs:      filepath.Join(busPath, "docs"),
		Workspace: filepath.Dir(busPath),
		App:       filepath.Join(busPath, "coordinator"),
	})
}

func NewFileRepositoryWithPaths(p Paths) *FileRepository {
	return &FileRepository{
		busPath:    p.Bus,
		dataPath:   p.Data,
		docsPath:   p.Docs,
		workspace:  p.Workspace,
		cursorPath: p.Cursor,
		appPath:    p.App,
	}
}

func (r *FileRepository) dataDir() string {
	if r.dataPath != "" {
		return r.dataPath
	}
	return filepath.Join(r.busPath, "data")
}

func (r *FileRepository) settingsDir() string {
	preferred := filepath.Join(r.dataDir(), "settings")
	if fileExists(filepath.Join(preferred, "team.json")) {
		return preferred
	}
	if r.cursorPath != "" && fileExists(filepath.Join(r.cursorPath, "team.json")) {
		return r.cursorPath
	}
	return preferred
}

func (r *FileRepository) progressDir() string {
	preferred := filepath.Join(r.dataDir(), "progress")
	if dirExists(preferred) {
		return preferred
	}
	legacy := filepath.Join(r.busPath, "progress")
	if dirExists(legacy) {
		return legacy
	}
	return preferred
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (r *FileRepository) GetTeam(_ context.Context) ([]model.TeamPerson, error) {
	path := filepath.Join(r.settingsDir(), "team.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.TeamPerson{}, nil
		}
		return nil, err
	}

	var file model.TeamFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if file.Members == nil {
		return []model.TeamPerson{}, nil
	}
	return file.Members, nil
}

func (r *FileRepository) SaveTeam(_ context.Context, members []model.TeamPerson) error {
	if members == nil {
		members = []model.TeamPerson{}
	}

	file := model.TeamFile{
		Version: 1,
		Members: members,
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := writeFileAtomic(filepath.Join(r.settingsDir(), "team.json"), data); err != nil {
		return err
	}

	authorPath := filepath.Join(r.settingsDir(), "author.md")
	current, err := os.ReadFile(authorPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	updated := rewriteAuthorMarkdown(string(current), members)
	return writeFileAtomic(authorPath, []byte(updated))
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func (r *FileRepository) WorkspaceDir() string {
	if r.workspace != "" {
		return r.workspace
	}
	if r.busPath == "" {
		return ""
	}
	return filepath.Dir(r.busPath)
}

func (r *FileRepository) appDir() string {
	if r.appPath != "" {
		return r.appPath
	}
	if r.busPath != "" {
		return filepath.Join(r.busPath, "coordinator")
	}
	return ""
}

func (r *FileRepository) appScript(name string) string {
	dir := r.appDir()
	nested := filepath.Join(dir, "utils", name)
	if fileExists(nested) {
		return nested
	}
	return filepath.Join(dir, name)
}

func (r *FileRepository) SaveProject(_ context.Context, profile *model.ProjectProfile) error {
	if profile == nil {
		return fmt.Errorf("project profile is nil")
	}
	if profile.Groups == nil {
		profile.Groups = []model.ServiceGroup{}
	}
	if profile.Services == nil {
		profile.Services = []model.ServiceNode{}
	}
	if profile.Edges == nil {
		profile.Edges = []model.ServiceEdge{}
	}
	if profile.Version == 0 {
		profile.Version = 2
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(filepath.Join(r.settingsDir(), "project_profile.json"), data)
}

func (r *FileRepository) GetProject(_ context.Context) (*model.ProjectProfile, error) {
	path := filepath.Join(r.settingsDir(), "project_profile.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.ProjectProfile{
				Groups:   []model.ServiceGroup{},
				Services: []model.ServiceNode{},
				Edges:    []model.ServiceEdge{},
			}, nil
		}
		return nil, err
	}

	var profile model.ProjectProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	if profile.Groups == nil {
		profile.Groups = []model.ServiceGroup{}
	}
	if profile.Services == nil {
		profile.Services = []model.ServiceNode{}
	}
	if profile.Edges == nil {
		profile.Edges = []model.ServiceEdge{}
	}
	return &profile, nil
}

func (r *FileRepository) GetAuthors(ctx context.Context) (map[string]string, error) {
	authors := make(map[string]string)

	team, err := r.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	for _, person := range team {
		if person.Alias == "" {
			continue
		}
		name := person.Name
		if name == "" {
			name = person.Alias
		}
		authors[person.Alias] = name
	}

	fromMarkdown, err := r.authorsFromMarkdown()
	if err != nil {
		return nil, err
	}
	for alias, name := range fromMarkdown {
		if _, exists := authors[alias]; exists {
			continue
		}
		authors[alias] = name
	}
	return authors, nil
}

func (r *FileRepository) authorsFromMarkdown() (map[string]string, error) {
	authors := make(map[string]string)
	authorFile := filepath.Join(r.settingsDir(), "author.md")

	file, err := os.Open(authorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return authors, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inTable := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "| alias |") {
			inTable = true
			continue
		}
		if inTable && strings.HasPrefix(line, "|-") {
			continue
		}
		if inTable && strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) < 3 {
				continue
			}
			alias := strings.TrimSpace(parts[1])
			name := strings.TrimSpace(parts[2])
			if alias == "" || name == "" {
				continue
			}
			authors[alias] = name
		}
	}
	return authors, scanner.Err()
}

func (r *FileRepository) GetMembers(ctx context.Context) ([]model.Member, error) {
	authors, err := r.GetAuthors(ctx)
	if err != nil {
		return nil, err
	}
	roster, err := r.teamByAlias(ctx)
	if err != nil {
		return nil, err
	}

	progressDir := r.progressDir()
	entries, err := os.ReadDir(progressDir)
	if err != nil {
		if os.IsNotExist(err) {
			return r.withRepoWork(r.idleAuthors(authors, roster, time.Now())), nil
		}
		return nil, err
	}

	now := time.Now()
	members := make([]model.Member, 0)
	seen := make(map[string]struct{})

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, ".current_task_") {
			continue
		}
		alias := strings.TrimPrefix(name, ".current_task_")
		data, err := os.ReadFile(filepath.Join(progressDir, name))
		if err != nil {
			continue
		}

		var raw snapshotFile
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		if raw.Alias == "" {
			raw.Alias = alias
		}
		if raw.Status == "" {
			raw.Status = "idle"
		}

		memberName := authors[raw.Alias]
		if memberName == "" {
			memberName = raw.Alias
		}

		updatedAt, parseErr := time.Parse(time.RFC3339, raw.UpdatedAt)
		if parseErr != nil {
			updatedAt = now
		}

		seen[raw.Alias] = struct{}{}
		person := roster[raw.Alias]
		member := model.Member{
			Alias:       raw.Alias,
			Name:        memberName,
			Role:        person.Role,
			Access:      person.Access,
			Focus:       person.Focus,
			Services:    raw.Services,
			Status:      raw.Status,
			TaskID:      raw.TaskID,
			TaskTitle:   r.taskTitle(raw.TaskID),
			TaskDoc:     raw.Doc,
			TaskSummary: raw.Summary,
			Branch:      raw.Branch,
			UpdatedAt:   updatedAt,
			CursorUsage: raw.CursorUsage,
			GitReport:   raw.GitReport,
			Tasks:       r.parseSnapshotTasks(raw, now),
		}
		if len(member.Tasks) > 0 {
			member.Status = "in_progress"
			newest := member.Tasks[len(member.Tasks)-1]
			if member.TaskID == "" {
				member.TaskID = newest.TaskID
				member.TaskTitle = newest.Title
				member.TaskDoc = newest.Doc
				member.TaskSummary = newest.Summary
				member.Branch = newest.Branch
				member.Services = newest.Services
				member.CursorUsage = newest.CursorUsage
			}
		}
		if raw.Research != nil && raw.Research.Status == "active" {
			started := updatedAt
			if raw.Research.StartedAt != "" {
				if t, err := time.Parse(time.RFC3339, raw.Research.StartedAt); err == nil {
					started = t
				}
			}
			d := int64(now.Sub(started).Seconds())
			if d < 0 {
				d = 0
			}
			member.Research = &model.Research{
				Status:          "active",
				Summary:         raw.Research.Summary,
				StartedAt:       started,
				SessionID:       raw.Research.SessionID,
				CursorUsage:     raw.Research.CursorUsage,
				DurationSeconds: d,
			}
		}
		members = append(members, member)
	}

	if len(members) == 0 {
		return r.withRepoWork(r.idleAuthors(authors, roster, now)), nil
	}

	for alias, name := range authors {
		if _, ok := seen[alias]; ok {
			continue
		}
		person := roster[alias]
		members = append(members, model.Member{
			Alias:     alias,
			Name:      name,
			Role:      person.Role,
			Access:    person.Access,
			Focus:     person.Focus,
			Status:    "idle",
			UpdatedAt: now,
		})
	}

	return r.withRepoWork(members), nil
}

func (r *FileRepository) withRepoWork(members []model.Member) []model.Member {
	r.maybeCloseStaleResearch(members)
	r.attachChatTabs(members)
	r.attachRepoWork(members)
	author, _ := r.CurrentAuthor(context.Background())
	for i := range members {
		if author != "" && members[i].Alias == author {
			continue
		}
		members[i].Repos = overlayReportedRepos(members[i].Repos, members[i].GitReport)
	}
	r.maybePublishGitReport(author, members)
	r.attachLiveUsage(members)
	return members
}

func (r *FileRepository) teamByAlias(ctx context.Context) (map[string]model.TeamPerson, error) {
	team, err := r.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]model.TeamPerson, len(team))
	for _, person := range team {
		if person.Alias == "" {
			continue
		}
		if person.Focus == nil {
			person.Focus = []string{}
		}
		out[person.Alias] = person
	}
	return out, nil
}

func (r *FileRepository) idleAuthors(authors map[string]string, roster map[string]model.TeamPerson, now time.Time) []model.Member {
	members := make([]model.Member, 0, len(authors))
	for alias, name := range authors {
		person := roster[alias]
		members = append(members, model.Member{
			Alias:     alias,
			Name:      name,
			Role:      person.Role,
			Access:    person.Access,
			Focus:     person.Focus,
			Status:    "idle",
			UpdatedAt: now,
		})
	}
	return members
}

type snapshotTask struct {
	TaskID      string             `json:"task_id"`
	Branch      string             `json:"branch"`
	Services    []string           `json:"services"`
	Doc         string             `json:"doc"`
	Summary     string             `json:"summary"`
	StartedAt   string             `json:"started_at"`
	UpdatedAt   string             `json:"updated_at"`
	CursorUsage *model.CursorUsage `json:"cursor_usage"`
	SessionID   string             `json:"session_id"`
	SessionIDs  []string           `json:"session_ids"`
}

type snapshotFile struct {
	Alias       string             `json:"alias"`
	TaskID      string             `json:"task_id"`
	Branch      string             `json:"branch"`
	Status      string             `json:"status"`
	UpdatedAt   string             `json:"updated_at"`
	Services    []string           `json:"services"`
	Doc         string             `json:"doc"`
	Summary     string             `json:"summary"`
	CursorUsage *model.CursorUsage `json:"cursor_usage"`
	GitReport   *model.GitReport   `json:"git_report"`
	Tasks       []snapshotTask     `json:"tasks"`
	Research    *struct {
		Status      string             `json:"status"`
		Summary     string             `json:"summary"`
		StartedAt   string             `json:"started_at"`
		SessionID   string             `json:"session_id"`
		CursorUsage *model.CursorUsage `json:"cursor_usage"`
	} `json:"research"`
}

func (r *FileRepository) parseSnapshotTasks(raw snapshotFile, now time.Time) []model.MemberTask {
	rows := raw.Tasks
	if len(rows) == 0 && raw.Status == "in_progress" && raw.TaskID != "" {
		rows = []snapshotTask{{
			TaskID:      raw.TaskID,
			Branch:      raw.Branch,
			Services:    raw.Services,
			Doc:         raw.Doc,
			Summary:     raw.Summary,
			StartedAt:   raw.UpdatedAt,
			UpdatedAt:   raw.UpdatedAt,
			CursorUsage: raw.CursorUsage,
		}}
	}
	out := make([]model.MemberTask, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.TaskID) == "" {
			continue
		}
		started := parseSnapshotTime(row.StartedAt, now)
		updated := parseSnapshotTime(row.UpdatedAt, started)
		d := int64(now.Sub(started).Seconds())
		if d < 0 {
			d = 0
		}
		out = append(out, model.MemberTask{
			TaskID:          row.TaskID,
			Title:           r.taskTitle(row.TaskID),
			Doc:             row.Doc,
			Summary:         row.Summary,
			Branch:          row.Branch,
			Services:        row.Services,
			StartedAt:       started,
			UpdatedAt:       updated,
			DurationSeconds: d,
			CursorUsage:     row.CursorUsage,
			SpendKind:       model.SpendKind(row.Services),
			SessionIDs:      collectSessionIDs(row.SessionID, row.SessionIDs),
		})
	}
	return out
}

func parseSnapshotTime(raw string, fallback time.Time) time.Time {
	if raw == "" {
		return fallback
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return fallback
	}
	return t
}
