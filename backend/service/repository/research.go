package repository

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"coordinator/infra"
	"coordinator/model"
)

const researchIdleTimeout = 30 * time.Minute

type composerSessionsCache struct {
	Sessions map[string]composerSessionRec `json:"sessions"`
}

type composerSessionRec struct {
	Alias       string `json:"alias"`
	LastAgentAt int64  `json:"last_agent_at"`
	LastText    string `json:"last_text"`
	Title       string `json:"title"`
}

func (r *FileRepository) maybeCloseStaleResearch(members []model.Member) {
	alias, err := r.CurrentAuthor(context.Background())
	if err != nil || alias == "" {
		return
	}
	cache := r.loadComposerSessions()
	for i := range members {
		if members[i].Alias != alias || members[i].Research == nil {
			continue
		}
		if !researchIsStale(members[i].Research, cache) {
			continue
		}
		findings := "Auto-closed after 30m without a reply in the research tab."
		if rec := cacheSession(cache, members[i].Research.SessionID); rec != nil && rec.LastText != "" {
			findings = clipResearch("Auto-closed idle. "+rec.LastText, 500)
		}
		if err := r.execResearchCompleted(alias, findings); err != nil {
			infra.LogWarn("research auto-close: %v", err)
			continue
		}
		members[i].Research = nil
	}
}

func researchIsStale(rs *model.Research, cache composerSessionsCache) bool {
	if rs == nil {
		return false
	}
	last := rs.StartedAt
	if rec := cacheSession(cache, rs.SessionID); rec != nil && rec.LastAgentAt > 0 {
		last = time.Unix(rec.LastAgentAt, 0)
	}
	if last.IsZero() {
		return false
	}
	return time.Since(last) >= researchIdleTimeout
}

func cacheSession(cache composerSessionsCache, sessionID string) *composerSessionRec {
	if sessionID == "" || cache.Sessions == nil {
		return nil
	}
	rec, ok := cache.Sessions[sessionID]
	if !ok {
		return nil
	}
	return &rec
}

func (r *FileRepository) loadComposerSessions() composerSessionsCache {
	raw, err := os.ReadFile(filepath.Join(r.dataDir(), ".cache", "composer_sessions.json"))
	if err != nil {
		return composerSessionsCache{}
	}
	var cache composerSessionsCache
	if json.Unmarshal(raw, &cache) != nil {
		return composerSessionsCache{}
	}
	return cache
}

func (r *FileRepository) execResearchCompleted(alias, findings string) error {
	script := r.appScript("sync_event.sh")
	if !fileExists(script) {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	args := []string{alias, "research_completed"}
	if findings != "" {
		args = append(args, "findings="+findings)
	}
	cmd := exec.CommandContext(ctx, script, args...)
	cmd.Dir = r.WorkspaceDir()
	return cmd.Run()
}

func clipResearch(text string, n int) string {
	if len(text) <= n {
		return text
	}
	return text[:n-3] + "..."
}
