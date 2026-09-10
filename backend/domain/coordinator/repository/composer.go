package repository

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"coordinator/model"

	_ "modernc.org/sqlite"
)

func collectSessionIDs(sessionID string, sessionIDs []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(sessionIDs)+1)
	add := func(raw string) {
		id := strings.TrimSpace(raw)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range sessionIDs {
		add(id)
	}
	add(sessionID)
	return out
}

func (r *FileRepository) attachChatTabs(members []model.Member) {
	cache := r.loadComposerSessions()
	var db *sql.DB
	defer func() {
		if db != nil {
			_ = db.Close()
		}
	}()
	titleOf := func(id string) string {
		if rec := cacheSession(cache, id); rec != nil {
			if name := strings.TrimSpace(rec.Title); name != "" {
				return name
			}
		}
		if db == nil {
			db = openComposerStateDB()
		}
		return composerTitleFromDB(db, id)
	}
	for i := range members {
		for j := range members[i].Tasks {
			members[i].Tasks[j].Chats = chatTabs(members[i].Tasks[j].SessionIDs, titleOf)
		}
		if members[i].Research == nil {
			continue
		}
		id := strings.TrimSpace(members[i].Research.SessionID)
		if id == "" {
			continue
		}
		tabs := chatTabs([]string{id}, titleOf)
		if len(tabs) == 0 {
			continue
		}
		tab := tabs[0]
		members[i].Research.Chat = &tab
	}
}

func chatTabs(ids []string, titleOf func(string) string) []model.ChatTab {
	if len(ids) == 0 {
		return nil
	}
	out := make([]model.ChatTab, 0, len(ids))
	for _, id := range ids {
		out = append(out, model.ChatTab{
			SessionID: id,
			Title:     strings.TrimSpace(titleOf(id)),
		})
	}
	return out
}

func openComposerStateDB() *sql.DB {
	path := composerStateDBPath()
	if path == "" {
		return nil
	}
	dsn := "file:" + filepath.ToSlash(path) + "?mode=ro"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil
	}
	return db
}

func composerStateDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "Cursor", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "AppData", "Roaming", "Cursor", "User", "globalStorage", "state.vscdb"),
	}
	for _, path := range candidates {
		if fileExists(path) {
			return path
		}
	}
	return ""
}

func composerTitleFromDB(db *sql.DB, sessionID string) string {
	if db == nil || sessionID == "" {
		return ""
	}
	var raw string
	err := db.QueryRow("SELECT value FROM composerHeaders WHERE composerId=?", sessionID).Scan(&raw)
	if err == nil {
		if name := jsonName(raw); name != "" {
			return name
		}
	}
	err = db.QueryRow("SELECT value FROM cursorDiskKV WHERE key=?", "composerData:"+sessionID).Scan(&raw)
	if err != nil {
		return ""
	}
	return jsonName(raw)
}

func jsonName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var payload struct {
		Name string `json:"name"`
	}
	if json.Unmarshal([]byte(raw), &payload) != nil {
		return ""
	}
	return strings.TrimSpace(payload.Name)
}
