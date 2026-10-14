package repository

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"coordinator/model"
)

const maxTaskDocBytes = 2 << 20

var taskIDFile = regexp.MustCompile(`(?i)^(?:FIX-)?\d{8}-\d{4}-[A-Z]{2,4}-(.+)$`)

func (r *FileRepository) TaskTitle(_ context.Context, taskID string) string {
	return r.taskTitle(taskID)
}

func (r *FileRepository) taskTitle(taskID string) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return ""
	}
	if path, ok := r.findTaskDocPath(taskID); ok {
		if title := firstHeading(path); title != "" {
			return title
		}
	}
	return humanizeTaskID(taskID)
}

func (r *FileRepository) GetTaskDoc(_ context.Context, taskID string) (*model.TaskDoc, error) {
	taskID = strings.TrimSpace(taskID)
	if !model.ValidTaskID(taskID) {
		return nil, os.ErrInvalid
	}
	path, ok := r.findTaskDocPath(taskID)
	if !ok {
		return nil, os.ErrNotExist
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	if info.IsDir() || info.Size() > maxTaskDocBytes {
		return nil, os.ErrInvalid
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(r.docRoot(), path)
	if err != nil {
		rel = filepath.Base(path)
	}
	title := firstHeadingFrom(string(raw))
	if title == "" {
		title = humanizeTaskID(taskID)
	}
	return &model.TaskDoc{
		TaskID:   taskID,
		Title:    title,
		RelPath:  filepath.ToSlash(rel),
		Markdown: string(raw),
	}, nil
}

func (r *FileRepository) findTaskDocPath(taskID string) (string, bool) {
	if !model.ValidTaskID(taskID) {
		return "", false
	}
	docs := r.docsDir()
	docsAbs, err := filepath.Abs(docs)
	if err != nil {
		return "", false
	}
	name := taskID + ".md"
	candidates := []string{filepath.Join(docs, name)}
	if matches, globErr := filepath.Glob(filepath.Join(docs, "*", name)); globErr == nil {
		candidates = append(candidates, matches...)
	}
	for _, candidate := range candidates {
		abs, absErr := filepath.Abs(candidate)
		if absErr != nil {
			continue
		}
		rel, relErr := filepath.Rel(docsAbs, abs)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			continue
		}
		if filepath.Base(abs) != name {
			continue
		}
		info, statErr := os.Stat(abs)
		if statErr != nil || info.IsDir() {
			continue
		}
		return abs, true
	}
	return "", false
}

func (r *FileRepository) docsDir() string {
	if r.docsPath != "" {
		return r.docsPath
	}
	return filepath.Join(r.busPath, "docs")
}

func (r *FileRepository) docRoot() string {
	docs := r.docsDir()
	if filepath.Base(docs) == "docs" {
		return filepath.Dir(docs)
	}
	return docs
}

func firstHeading(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if title := headingLine(scanner.Text()); title != "" {
			return title
		}
	}
	return ""
}

func firstHeadingFrom(markdown string) string {
	for _, line := range strings.Split(markdown, "\n") {
		if title := headingLine(line); title != "" {
			return title
		}
	}
	return ""
}

func headingLine(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "# ") {
		return strings.TrimSpace(strings.TrimPrefix(line, "# "))
	}
	return ""
}

func humanizeTaskID(taskID string) string {
	slug := taskID
	if m := taskIDFile.FindStringSubmatch(taskID); len(m) == 2 {
		slug = m[1]
	}
	slug = strings.ReplaceAll(slug, "_", " ")
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return taskID
	}
	runes := []rune(slug)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
