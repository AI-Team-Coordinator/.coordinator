package coordinator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"coordinator/domain/coordinator/dto"
	"coordinator/infra"
)

var processStartedAt = time.Now().UTC().Format(time.RFC3339)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) GetHealth(w http.ResponseWriter, r *http.Request) {
	infra.ReturnJSON(w, http.StatusOK, dto.HealthResponse{
		Status:    "ok",
		Service:   "coordinator",
		StartedAt: processStartedAt,
	})
}

func (c *Controller) GetTeam(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetTeam(r.Context())
	if err != nil {
		infra.LogError("GetTeam failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load team", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PutTeam(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.SaveTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	out, err := c.service.SaveTeam(r.Context(), req)
	if err != nil {
		var vErr *ValidationError
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PutTeam failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to save team", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetSyncStatus(r.Context())
	if err != nil {
		infra.LogError("GetSyncStatus failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load sync status", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PostSync(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.SyncCursor(r.Context())
	if err != nil {
		var vErr *ValidationError
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		var gErr *GitSyncError
		if errors.As(err, &gErr) {
			code := "E500"
			status := http.StatusInternalServerError
			if gErr.Kind == "conflict" {
				code = "E409"
				status = http.StatusConflict
			}
			infra.ReturnError(w, code, gErr.Error(), status)
			return
		}
		infra.LogError("PostSync failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to sync Common settings", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetProject(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetProject(r.Context())
	if err != nil {
		infra.LogError("GetProject failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load project profile", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetSetup(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetSetup(r.Context())
	if err != nil {
		infra.LogError("GetSetup failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load setup", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PostSetup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.CompleteSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	out, err := c.service.CompleteSetup(r.Context(), req)
	if err != nil {
		var vErr *ValidationError
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PostSetup failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to save setup", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PostCreateService(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	out, err := c.service.CreateService(r.Context(), req)
	if err != nil {
		var vErr *ValidationError
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		var gErr *GitHubOpError
		if errors.As(err, &gErr) {
			code := "E502"
			status := http.StatusBadGateway
			if gErr.Kind == "exists" {
				code = "E409"
				status = http.StatusConflict
			}
			infra.ReturnError(w, code, gErr.Msg, status)
			return
		}
		infra.LogError("PostCreateService failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to create project unit", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusCreated, out)
}

func (c *Controller) GetPulse(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetPulse(r.Context())
	if err != nil {
		infra.LogError("GetPulse failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load pulse", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetMembers(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetMembers(r.Context())
	if err != nil {
		infra.LogError("GetMembers failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load members", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetConflicts(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetConflicts(r.Context())
	if err != nil {
		infra.LogError("GetConflicts failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load conflicts", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetStats(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetStats(r.Context())
	if err != nil {
		infra.LogError("GetStats failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load stats", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetTasks(w http.ResponseWriter, r *http.Request) {
	query, err := parseTasksQuery(r)
	if err != nil {
		infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
		return
	}

	out, err := c.service.GetTasks(r.Context(), query)
	if err != nil {
		if errors.Is(err, ErrInvalidLimit) || errors.Is(err, ErrInvalidOffset) || errors.Is(err, ErrInvalidStatus) || errors.Is(err, ErrInvalidKind) {
			infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
			return
		}
		infra.LogError("GetTasks failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load tasks", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetEvents(w http.ResponseWriter, r *http.Request) {
	query, err := parseEventsQuery(r)
	if err != nil {
		infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
		return
	}

	out, err := c.service.GetEvents(r.Context(), query)
	if err != nil {
		if errors.Is(err, ErrInvalidDays) || errors.Is(err, ErrInvalidLimit) || errors.Is(err, ErrInvalidOffset) {
			infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
			return
		}
		infra.LogError("GetEvents failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load events", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetTaskDoc(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetTaskDoc(r.Context(), r.PathValue("taskID"))
	if err != nil {
		if errors.Is(err, ErrInvalidTaskID) {
			infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrDocNotFound) {
			infra.ReturnError(w, "E404", err.Error(), http.StatusNotFound)
			return
		}
		infra.LogError("GetTaskDoc failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load document", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		infra.ReturnError(w, "E500", "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			out, err := c.service.GetPulse(r.Context())
			if err != nil {
				infra.LogWarn("stream pulse: %v", err)
				continue
			}
			data, err := json.Marshal(out)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func parseTasksQuery(r *http.Request) (dto.TasksQuery, error) {
	q := dto.TasksQuery{
		Limit:  50,
		Status: r.URL.Query().Get("status"),
		Alias:  r.URL.Query().Get("alias"),
		Kind:   r.URL.Query().Get("kind"),
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return q, errors.New("limit must be an integer")
		}
		q.Limit = limit
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil {
			return q, errors.New("offset must be an integer")
		}
		q.Offset = offset
	}
	return q, nil
}

func parseEventsQuery(r *http.Request) (dto.EventsQuery, error) {
	q := dto.EventsQuery{
		Limit:   50,
		Alias:   r.URL.Query().Get("alias"),
		Event:   r.URL.Query().Get("event"),
		Service: r.URL.Query().Get("service"),
	}

	if raw := r.URL.Query().Get("days"); raw != "" {
		days, err := strconv.Atoi(raw)
		if err != nil {
			return q, errors.New("days must be an integer")
		}
		q.Days = days
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return q, errors.New("limit must be an integer")
		}
		q.Limit = limit
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil {
			return q, errors.New("offset must be an integer")
		}
		q.Offset = offset
	}
	return q, nil
}
