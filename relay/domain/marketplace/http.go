package marketplace

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"relay/domain/connector"
)

type Controller struct {
	Service *Service
	Actions *connector.ActionRegistry
}

func (c *Controller) ListActivity(w http.ResponseWriter, r *http.Request) {
	if c.Service == nil {
		http.Error(w, "not configured", http.StatusServiceUnavailable)
		return
	}
	connectorID := strings.TrimSpace(r.URL.Query().Get("connector_id"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	out, err := c.Service.ListActivity(r.Context(), connectorID, limit, offset)
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
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

func (c *Controller) RunAction(w http.ResponseWriter, r *http.Request) {
	if c.Actions == nil {
		http.Error(w, "actions not configured", http.StatusNotImplemented)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	action := strings.TrimSpace(r.PathValue("action"))
	if id == "" || action == "" {
		http.Error(w, "connector id and action are required", http.StatusBadRequest)
		return
	}
	if err := c.Actions.Run(r.Context(), id, action); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
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
