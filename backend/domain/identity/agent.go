package identity

import "strings"

// CursorChat is one Composer tab acting for a person.
func CursorChat(person, sessionID string) Agent {
	id := strings.TrimSpace(sessionID)
	if id == "" {
		return Agent{}
	}
	return Agent{
		ID:     id,
		Person: strings.TrimSpace(person),
		Kind:   KindCursorChat,
	}
}

// CursorChats maps session ids to agents. Empty ids are skipped.
func CursorChats(person string, sessionIDs []string) []Agent {
	out := make([]Agent, 0, len(sessionIDs))
	seen := make(map[string]struct{}, len(sessionIDs))
	for _, id := range sessionIDs {
		a := CursorChat(person, id)
		if a.ID == "" {
			continue
		}
		if _, dup := seen[a.ID]; dup {
			continue
		}
		seen[a.ID] = struct{}{}
		out = append(out, a)
	}
	return out
}

// WithChatTitles copies titles from Composer tabs onto matching agents.
func WithChatTitles(agents []Agent, titles map[string]string) []Agent {
	if len(agents) == 0 || len(titles) == 0 {
		return agents
	}
	out := append([]Agent(nil), agents...)
	for i := range out {
		if t := strings.TrimSpace(titles[out[i].ID]); t != "" {
			out[i].Title = t
		}
	}
	return out
}

func (a Agent) IsZero() bool {
	return strings.TrimSpace(a.ID) == ""
}
