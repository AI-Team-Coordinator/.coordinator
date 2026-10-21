package marketplace

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Controller struct {
	Service *Service
}

func (c *Controller) Marketplace(w http.ResponseWriter, r *http.Request) {
	if c.Service == nil {
		http.Error(w, "not configured", http.StatusServiceUnavailable)
		return
	}
	list, err := c.Service.List(r.Context())
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"connectors": list})
}

func (c *Controller) GetConfig(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	view, err := c.Service.GetConfig(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(view)
}

func (c *Controller) PutConfig(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	var req PutConfigRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	view, err := c.Service.PutConfig(r.Context(), id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not connectable") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(view)
}

type disableBody struct {
	ClearSecrets bool `json:"clear_secrets"`
}

func (c *Controller) Disable(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	var body disableBody
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body)
	view, err := c.Service.Disable(r.Context(), id, body.ClearSecrets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(view)
}
