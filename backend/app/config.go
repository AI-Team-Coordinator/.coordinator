package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port        string
	AppRoot     string
	Workspace   string
	DataDir     string
	DocsDir     string
	BusDir      string
	CursorDir   string
	FrontendDir string
	// UIMode is "vite" (Alina Assist HMR) or "static" (built dist on PORT).
	UIMode string
}

func DefaultConfig() *Config {
	return &Config{Port: "4321"}
}

func (c *Config) ResolvePaths(cwd string) {
	c.AppRoot = detectAppRoot(cwd)
	loadDotEnv(c.AppRoot)

	if c.Port == "" || c.Port == "4321" {
		if port := os.Getenv("PORT"); port != "" {
			c.Port = port
		}
	}

	c.Workspace = firstNonEmpty(c.Workspace, os.Getenv("WORKSPACE_ROOT"))
	c.DataDir = firstNonEmpty(c.DataDir, os.Getenv("DATA_DIR"))
	c.DocsDir = firstNonEmpty(c.DocsDir, os.Getenv("DOCS_DIR"))
	c.BusDir = firstNonEmpty(c.BusDir, os.Getenv("BUS_DIR"), os.Getenv("COMMON_DIR"))
	c.CursorDir = firstNonEmpty(c.CursorDir, os.Getenv("CURSOR_DIR"))
	c.FrontendDir = firstNonEmpty(c.FrontendDir, os.Getenv("FRONTEND_DIR"))

	c.Workspace = resolvePath(c.AppRoot, c.Workspace, "..")
	c.BusDir = resolvePath(c.AppRoot, c.BusDir, filepath.Join("..", "Common"))
	c.DataDir = resolvePath(c.AppRoot, c.DataDir, filepath.Join(c.BusDir, "data"))
	c.DocsDir = resolvePath(c.AppRoot, c.DocsDir, filepath.Join(c.BusDir, "docs"))
	c.CursorDir = resolvePath(c.AppRoot, c.CursorDir, filepath.Join("..", ".cursor"))

	c.UIMode = detectUIMode(c.DataDir)
	if c.UIMode == "vite" {
		// API only — Vite on UI_PORT is the board. Do not serve a stale dist.
		c.FrontendDir = ""
		return
	}

	if c.FrontendDir != "" {
		c.FrontendDir = resolvePath(c.AppRoot, c.FrontendDir, "")
		if _, err := os.Stat(filepath.Join(c.FrontendDir, "dist", "index.html")); err == nil {
			c.FrontendDir = filepath.Join(c.FrontendDir, "dist")
		}
		return
	}
	c.FrontendDir = detectDir("index.html", []string{
		filepath.Join(c.AppRoot, "frontend", "dist"),
		filepath.Join(c.AppRoot, "frontend"),
	})
}

func detectUIMode(dataDir string) string {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("UI_MODE")))
	if raw == "vite" || raw == "static" {
		return raw
	}
	if isAlinaAssistProfile(dataDir) {
		return "vite"
	}
	return "static"
}

func isAlinaAssistProfile(dataDir string) bool {
	path := filepath.Join(dataDir, "settings", "project_profile.json")
	body, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var file struct {
		Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
	}
	if json.Unmarshal(body, &file) != nil {
		return false
	}
	id := strings.ToLower(strings.TrimSpace(file.Project.ID))
	name := strings.ToLower(strings.TrimSpace(file.Project.Name))
	return id == "alina-assist" || name == "alina assist"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func detectDir(marker string, candidates []string) string {
	for _, cand := range candidates {
		if _, err := os.Stat(filepath.Join(cand, marker)); err == nil {
			return cand
		}
	}
	return ""
}
