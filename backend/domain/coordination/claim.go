package coordination

import (
	"strings"

	"coordinator/domain/identity"
)

// Claim is an agent's declared zone for a task: "I intend to work here."
// Person owns it. Agents are who is executing it (Composer tabs today).
// It is not a git lock. Exclusive vs advisory is overlap policy, not a second record.
type Claim struct {
	Person   string
	Agents   []identity.Agent
	TaskID   string
	Title    string
	Branch   string
	Resource Resource
}

// Dependency is a different edge: "I need what is happening here," without owning it.
type Dependency struct {
	Person   string
	TaskID   string
	Resource Resource
	Reason   string
}

// ParseClaims projects today's services[] into claims. Nested labels (Core/billing)
// stay one claim on Core with a path; the snapshot format does not change.
func ParseClaims(person, taskID, title, branch string, services []string, agents []identity.Agent) []Claim {
	copied := append([]identity.Agent(nil), agents...)
	out := make([]Claim, 0, len(services))
	seen := make(map[string]struct{})
	for _, raw := range services {
		res := Parse(raw)
		if res.IsZero() {
			continue
		}
		fp := res.Key + "\x00" + res.Path
		if _, dup := seen[fp]; dup {
			continue
		}
		seen[fp] = struct{}{}
		out = append(out, Claim{
			Person:   strings.TrimSpace(person),
			Agents:   copied,
			TaskID:   strings.TrimSpace(taskID),
			Title:    strings.TrimSpace(title),
			Branch:   strings.TrimSpace(branch),
			Resource: res,
		})
	}
	return out
}
