package app

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func loadDotEnv(appRoot string) {
	path := filepath.Join(appRoot, ".env")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

func resolvePath(appRoot, raw, fallback string) string {
	if strings.TrimSpace(raw) == "" {
		raw = fallback
	}
	if raw == "" {
		return ""
	}
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw)
	}
	abs, err := filepath.Abs(filepath.Join(appRoot, raw))
	if err != nil {
		return filepath.Join(appRoot, raw)
	}
	return abs
}

func detectAppRoot(cwd string) string {
	if root := strings.TrimSpace(os.Getenv("COORDINATOR_ROOT")); root != "" {
		if abs, err := filepath.Abs(root); err == nil {
			return abs
		}
		return root
	}
	candidates := []string{cwd, filepath.Join(cwd, ".."), filepath.Join(cwd, "../..")}
	for _, cand := range candidates {
		abs, err := filepath.Abs(cand)
		if err != nil {
			continue
		}
		if fileExists(filepath.Join(abs, ".env")) || fileExists(filepath.Join(abs, ".env.example")) || fileExists(filepath.Join(abs, "run.sh")) {
			return abs
		}
	}
	abs, _ := filepath.Abs(cwd)
	return abs
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
