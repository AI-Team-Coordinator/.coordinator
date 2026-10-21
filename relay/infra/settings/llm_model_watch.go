package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const llmModelWatchFile = "relay_llm_model_watch.json"

// LLMModelWatch lists model ids to monitor for provider end-of-life / expiration.
type LLMModelWatch struct {
	Version int      `json:"version"`
	Models  []string `json:"models"`
}

func (r *RelayConnectors) llmModelWatchPath() string {
	return filepath.Join(r.dataDir, "settings", llmModelWatchFile)
}

func (r *RelayConnectors) LoadLLMModelWatch() (LLMModelWatch, error) {
	path := r.llmModelWatchPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LLMModelWatch{Version: 1, Models: nil}, nil
		}
		return LLMModelWatch{}, err
	}
	var out LLMModelWatch
	if err := json.Unmarshal(data, &out); err != nil {
		return LLMModelWatch{}, err
	}
	if out.Version == 0 {
		out.Version = 1
	}
	return out, nil
}
