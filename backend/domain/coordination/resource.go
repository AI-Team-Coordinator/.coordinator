package coordination

import "strings"

// Kind classifies a resource for overlap policy.
type Kind string

const (
	KindProduct   Kind = "product"
	KindBus       Kind = "bus"
	KindWorkspace Kind = "workspace"
)

// Resource is a zone of responsibility, not a file.
// Label "LLM/stt" is service LLM with path stt. Files are out of scope.
type Resource struct {
	Label string
	Key   string
	Path  string
	Kind  Kind
}

// Parse turns a services[] label into a Resource.
func Parse(raw string) Resource {
	label := strings.TrimSpace(raw)
	if label == "" {
		return Resource{}
	}
	trimmed := strings.TrimPrefix(label, ".")
	parts := strings.Split(trimmed, "/")
	key := normalize(parts[0])
	if key == "" {
		return Resource{Label: label}
	}
	pathParts := make([]string, 0, len(parts)-1)
	for _, p := range parts[1:] {
		if n := normalize(p); n != "" {
			pathParts = append(pathParts, n)
		}
	}
	return Resource{
		Label: label,
		Key:   key,
		Path:  strings.Join(pathParts, "/"),
		Kind:  kindOf(key),
	}
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, ".")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func kindOf(key string) Kind {
	switch key {
	case "common":
		return KindBus
	case "cursor", "coordinator":
		return KindWorkspace
	default:
		return KindProduct
	}
}

// Relates is true when two resources share a service and one path covers the other.
// Empty path means the whole service.
func (r Resource) Relates(other Resource) bool {
	if r.Key == "" || other.Key == "" || r.Key != other.Key {
		return false
	}
	if r.Path == other.Path {
		return true
	}
	if r.Path == "" || other.Path == "" {
		return true
	}
	return strings.HasPrefix(r.Path+"/", other.Path+"/") || strings.HasPrefix(other.Path+"/", r.Path+"/")
}

func (r Resource) IsZero() bool {
	return r.Key == ""
}
