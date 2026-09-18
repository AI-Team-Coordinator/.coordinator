package model

import "strings"

// NormalizeRelatedTasks drops blanks, self, and duplicates. Order is kept.
func NormalizeRelatedTasks(ids []string, selfID string) []string {
	selfID = strings.TrimSpace(selfID)
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" || id == selfID {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// NormalizeRelatedDocs keeps Common-relative docs/ paths, unique, in order.
func NormalizeRelatedDocs(paths []string) []string {
	out := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, raw := range paths {
		path := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
		path = strings.TrimPrefix(path, "./")
		path = strings.TrimPrefix(path, "Common/")
		if path == "" || strings.HasPrefix(path, "/") {
			continue
		}
		if !strings.HasPrefix(path, "docs/") {
			path = "docs/" + strings.TrimPrefix(path, "/")
		}
		if _, dup := seen[path]; dup {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	return out
}

// MergeRelated keeps dst unless src is non-empty.
func MergeRelated(dst, src []string) []string {
	if len(src) > 0 {
		return append([]string(nil), src...)
	}
	if len(dst) == 0 {
		return nil
	}
	return append([]string(nil), dst...)
}

// TaskIDFromDocPath returns a task id if the markdown filename is a task doc.
func TaskIDFromDocPath(path string) string {
	base := path
	if i := strings.LastIndex(path, "/"); i >= 0 {
		base = path[i+1:]
	}
	base = strings.TrimSuffix(base, ".md")
	if ValidTaskID(base) {
		return base
	}
	return ""
}
