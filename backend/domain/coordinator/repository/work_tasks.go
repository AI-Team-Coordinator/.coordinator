package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"coordinator/model"
)

type workTaskFile struct {
	TaskID       string   `json:"task_id"`
	Summary      string   `json:"summary"`
	Doc          string   `json:"doc"`
	Branch       string   `json:"branch"`
	Services     []string `json:"services"`
	CreatedBy    string   `json:"created_by"`
	Assignee     string   `json:"assignee"`
	Status       string   `json:"status"`
	RelatedTasks []string `json:"related_tasks"`
	RelatedDocs  []string `json:"related_docs"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	CompletedAt  string   `json:"completed_at"`
}

func (r *FileRepository) workTasksDir() string {
	return filepath.Join(r.dataDir(), "progress", "work_tasks")
}

func (r *FileRepository) ListWorkTasks(ctx context.Context) ([]model.WorkTask, error) {
	dir := r.workTasksDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.WorkTask{}, nil
		}
		return nil, err
	}
	out := make([]model.WorkTask, 0, len(entries))
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, ent.Name())
		row, err := readWorkTaskFile(path)
		if err != nil || row.TaskID == "" {
			continue
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func readWorkTaskFile(path string) (model.WorkTask, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return model.WorkTask{}, err
	}
	var f workTaskFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return model.WorkTask{}, err
	}
	now := time.Now().UTC()
	return model.WorkTask{
		TaskID:       strings.TrimSpace(f.TaskID),
		Summary:      f.Summary,
		Doc:          f.Doc,
		Branch:       f.Branch,
		Services:     f.Services,
		CreatedBy:    f.CreatedBy,
		Assignee:     f.Assignee,
		Status:       normalizeWorkStatus(f.Status),
		RelatedTasks: model.NormalizeRelatedTasks(f.RelatedTasks, f.TaskID),
		RelatedDocs:  model.NormalizeRelatedDocs(f.RelatedDocs),
		CreatedAt:    parseSnapshotTime(f.CreatedAt, now),
		UpdatedAt:    parseSnapshotTime(f.UpdatedAt, now),
		CompletedAt:  parseSnapshotTime(f.CompletedAt, time.Time{}),
	}, nil
}

func normalizeWorkStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "backlog", "open", "done":
		return strings.TrimSpace(strings.ToLower(raw))
	case "active":
		return "open"
	default:
		return "backlog"
	}
}
