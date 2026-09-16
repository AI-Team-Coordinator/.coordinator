// Package identity is who participates in coordination.
//
// Person is accountable. Agent acts (a Cursor chat today, later a subagent).
// Session is the Composer tab bound to a task or research.
package identity

// Person is a human teammate from team.json.
type Person struct {
	Alias string
	Name  string
}

// Agent is who is acting. Not the Pulse subject: the team map still shows Person.
type Agent struct {
	ID     string // session_id today
	Person string // Person.Alias
	Kind   string // cursor_chat
	Title  string // Composer tab title, optional
}

const KindCursorChat = "cursor_chat"

// Session is a Composer tab.
type Session struct {
	ID    string
	Title string
}
