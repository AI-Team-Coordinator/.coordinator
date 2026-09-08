package app

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port      string
	AppRoot   string
	Workspace string
	DataDir   string
	DocsDir   string
	BusDir    string
	CursorDir string
	WebDir    string
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
	c.WebDir = firstNonEmpty(c.WebDir, os.Getenv("WEB_DIR"))

	c.Workspace = resolvePath(c.AppRoot, c.Workspace, "..")
	c.BusDir = resolvePath(c.AppRoot, c.BusDir, filepath.Join("..", "Common"))
	c.DataDir = resolvePath(c.AppRoot, c.DataDir, filepath.Join(c.BusDir, "data"))
	c.DocsDir = resolvePath(c.AppRoot, c.DocsDir, filepath.Join(c.BusDir, "docs"))
	c.CursorDir = resolvePath(c.AppRoot, c.CursorDir, filepath.Join("..", ".cursor"))

	if c.WebDir != "" {
		c.WebDir = resolvePath(c.AppRoot, c.WebDir, "")
		if _, err := os.Stat(filepath.Join(c.WebDir, "dist", "index.html")); err == nil {
			c.WebDir = filepath.Join(c.WebDir, "dist")
		}
		return
	}
	c.WebDir = detectDir("index.html", []string{
		filepath.Join(c.AppRoot, "web", "dist"),
		filepath.Join(c.AppRoot, "web"),
	})
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
