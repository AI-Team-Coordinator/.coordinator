package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// InstanceFiles holds optional operator-edited JSON on RELAY_DATA (not connector marketplace meta).
type InstanceFiles struct {
	dataDir string
}

func NewInstanceFiles(dataDir string) *InstanceFiles {
	return &InstanceFiles{dataDir: dataDir}
}

const llmModelWatchFile = "relay_llm_model_watch.json"
const openRouterWatchFile = "relay_openrouter_watch.json"

// LLMModelWatch lists model ids to monitor for provider end-of-life / expiration.
type LLMModelWatch struct {
	Version int      `json:"version"`
	Models  []string `json:"models"`
}

// OpenRouterWatch lists OpenRouter model ids to monitor for expiration_date.
type OpenRouterWatch struct {
	Version int      `json:"version"`
	Models  []string `json:"models"`
}

func (f *InstanceFiles) settingsDir() string {
	return filepath.Join(f.dataDir, "settings")
}

func (f *InstanceFiles) LoadLLMModelWatch() (LLMModelWatch, error) {
	return loadWatchFile[LLMModelWatch](filepath.Join(f.settingsDir(), llmModelWatchFile))
}

func (f *InstanceFiles) LoadOpenRouterWatch() (OpenRouterWatch, error) {
	return loadWatchFile[OpenRouterWatch](filepath.Join(f.settingsDir(), openRouterWatchFile))
}

func loadWatchFile[T any](path string) (T, error) {
	var zero T
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, nil
		}
		return zero, err
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return zero, err
	}
	return out, nil
}

// NormalizeModelIDs deduplicates and trims model slug lists.
func NormalizeModelIDs(ids []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
