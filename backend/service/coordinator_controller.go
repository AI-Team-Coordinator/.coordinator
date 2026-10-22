package service

import (
	"coordinator/domain/validate"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	eventdto "coordinator/domain/timeline/dto"
	"coordinator/infra"
	healthdto "coordinator/service/health/dto"
	orgdto "coordinator/service/org/dto"
	projectdto "coordinator/service/project/dto"
	relaydto "coordinator/service/relay/dto"
	settingsdto "coordinator/service/settings/dto"
	setupdto "coordinator/service/setup/dto"
	statsdto "coordinator/service/stats/dto"
	taskdto "coordinator/service/task/dto"
	teamdto "coordinator/service/team/dto"
)

var processStartedAt = time.Now().UTC().Format(time.RFC3339)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) GetHealth(w http.ResponseWriter, r *http.Request) {
	infra.ReturnJSON(w, http.StatusOK, healthdto.HealthResponse{
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
	var req teamdto.SaveTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	out, err := c.service.SaveTeam(r.Context(), req)
	if err != nil {
		var vErr *validate.Error
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

func (c *Controller) GetOrgTopology(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetOrgTopology(r.Context())
	if err != nil {
		infra.LogError("GetOrgTopology failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load org topology", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PutOrgTopology(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req orgdto.SaveOrgTopologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	out, err := c.service.SaveOrgTopology(r.Context(), req)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PutOrgTopology failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to save org topology", http.StatusInternalServerError)
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
		var vErr *validate.Error
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
	var req setupdto.CompleteSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	out, err := c.service.CompleteSetup(r.Context(), req)
	if err != nil {
		var vErr *validate.Error
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
	var req projectdto.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}

	out, err := c.service.CreateService(r.Context(), req)
	if err != nil {
		var vErr *validate.Error
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
	query, err := parseStatsQuery(r)
	if err != nil {
		infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
		return
	}
	out, err := c.service.GetStats(r.Context(), query)
	if err != nil {
		infra.LogError("GetStats failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load stats", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func parseStatsQuery(r *http.Request) (statsdto.StatsQuery, error) {
	q := statsdto.StatsQuery{}
	if raw := r.URL.Query().Get("from"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return q, errors.New("from must be a unix timestamp")
		}
		q.From = v
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return q, errors.New("to must be a unix timestamp")
		}
		q.To = v
	}
	return q, nil
}

func (c *Controller) GetWorkItems(w http.ResponseWriter, r *http.Request) {
	query := taskdto.WorkItemsQuery{
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
		Assignee: strings.TrimSpace(r.URL.Query().Get("assignee")),
	}
	out, err := c.service.GetWorkItems(r.Context(), query)
	if err != nil {
		infra.LogError("GetWorkItems failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load work items", http.StatusInternalServerError)
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

func (c *Controller) GetExternalEvents(w http.ResponseWriter, r *http.Request) {
	source, limit, offset, err := parseExternalEventsQuery(r)
	if err != nil {
		infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
		return
	}
	out, err := c.service.GetExternalEvents(r.Context(), source, limit, offset)
	if err != nil {
		if strings.Contains(err.Error(), "limit") || strings.Contains(err.Error(), "offset") {
			infra.ReturnError(w, "E400", err.Error(), http.StatusBadRequest)
			return
		}
		infra.LogError("GetExternalEvents failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load external events", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PatchExternalEventTriage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req PatchExternalEventTriageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	alias, err := c.service.CurrentAuthorAlias(r.Context())
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.ReturnError(w, "E500", "Failed to resolve author", http.StatusInternalServerError)
		return
	}
	out, err := c.service.PatchExternalEventTriage(r.Context(), req, alias)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PatchExternalEventTriage failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to update triage", http.StatusInternalServerError)
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

func parseTasksQuery(r *http.Request) (statsdto.TasksQuery, error) {
	q := statsdto.TasksQuery{
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

func (c *Controller) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetUserSettings(r.Context())
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("GetUserSettings failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load user settings", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PutUserSettings(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req settingsdto.SaveUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	out, err := c.service.SaveUserSettings(r.Context(), req)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PutUserSettings failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to save user settings", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetTeamAvatar(w http.ResponseWriter, r *http.Request) {
	alias := strings.TrimSpace(r.PathValue("alias"))
	if alias == "" {
		http.NotFound(w, r)
		return
	}
	data, err := c.service.ReadTeamAvatar(r.Context(), alias)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = w.Write(data)
}

func (c *Controller) GetUserAvatar(w http.ResponseWriter, r *http.Request) {
	settings, err := c.service.GetUserSettings(r.Context())
	if err != nil || settings.AvatarURL == "" {
		http.NotFound(w, r)
		return
	}
	data, err := c.service.ReadUserAvatar(r.Context())
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=60")
	_, _ = w.Write(data)
}

func (c *Controller) PostUserAvatar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 512<<10+1<<20)
	if err := r.ParseMultipartForm(512 << 10); err != nil {
		infra.ReturnError(w, "E400", "Invalid multipart form", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		infra.ReturnError(w, "E400", "Missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 512<<10))
	if err != nil {
		infra.ReturnError(w, "E400", "Failed to read upload", http.StatusBadRequest)
		return
	}
	out, err := c.service.SaveUserAvatar(r.Context(), data)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PostUserAvatar failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to save avatar", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetRelayAccess(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetRelayAccess(r.Context())
	if err != nil {
		infra.LogError("GetRelayAccess failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load relay access", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PostRelayConnect(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req relaydto.RelayConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	out, err := c.service.ConnectRelay(r.Context(), req.URL, req.Token)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PostRelayConnect failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to connect relay", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PostRelayDisconnect(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.DisconnectRelay(r.Context())
	if err != nil {
		infra.LogError("PostRelayDisconnect failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to disconnect relay", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetRelaySources(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetRelaySources(r.Context())
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("GetRelaySources failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load relay sources", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetRelayIssuedTokens(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.ListRelayIssuedTokens(r.Context())
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("GetRelayIssuedTokens failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to list relay tokens", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PostRelayIssuedToken(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req relaydto.RelayIssueTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	out, err := c.service.IssueRelayToken(r.Context(), req)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PostRelayIssuedToken failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to issue relay token", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetRelayMarketplace(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.GetRelayMarketplace(r.Context())
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("GetRelayMarketplace failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load relay marketplace", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetRelayConnectorActivity(w http.ResponseWriter, r *http.Request) {
	connectorID := strings.TrimSpace(r.URL.Query().Get("connector_id"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	out, err := c.service.GetRelayConnectorActivity(r.Context(), connectorID, limit, offset)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("GetRelayConnectorActivity failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load connector activity", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) GetRelayConnectorConfig(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	out, err := c.service.GetRelayConnectorConfig(r.Context(), id)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("GetRelayConnectorConfig failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to load connector config", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PutRelayConnectorConfig(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req relaydto.RelayPutConnectorConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		infra.ReturnError(w, "E400", "Invalid JSON body", http.StatusBadRequest)
		return
	}
	out, err := c.service.PutRelayConnectorConfig(r.Context(), id, req)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PutRelayConnectorConfig failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to save connector config", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) PostRelayConnectorAction(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	action := strings.TrimSpace(r.PathValue("action"))
	err := c.service.TriggerRelayConnectorAction(r.Context(), id, action)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PostRelayConnectorAction failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to run connector action", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (c *Controller) PostRelayConnectorDisable(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var body struct {
		ClearSecrets bool `json:"clear_secrets"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	out, err := c.service.DisableRelayConnector(r.Context(), id, body.ClearSecrets)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("PostRelayConnectorDisable failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to disable connector", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) DeleteRelayIssuedToken(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	out, err := c.service.RevokeRelayIssuedToken(r.Context(), id)
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("DeleteRelayIssuedToken failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to revoke relay token", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func (c *Controller) DeleteUserAvatar(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.DeleteUserAvatar(r.Context())
	if err != nil {
		var vErr *validate.Error
		if errors.As(err, &vErr) {
			infra.ReturnError(w, "E400", vErr.Msg, http.StatusBadRequest)
			return
		}
		infra.LogError("DeleteUserAvatar failed: %v", err)
		infra.ReturnError(w, "E500", "Failed to remove avatar", http.StatusInternalServerError)
		return
	}
	infra.ReturnJSON(w, http.StatusOK, out)
}

func parseExternalEventsQuery(r *http.Request) (source string, limit int, offset int, err error) {
	source = strings.TrimSpace(r.URL.Query().Get("source"))
	limit = 20
	offset = 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			return "", 0, 0, errors.New("limit must be an integer")
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil {
			return "", 0, 0, errors.New("offset must be an integer")
		}
	}
	return source, limit, offset, nil
}

func parseEventsQuery(r *http.Request) (eventdto.EventsQuery, error) {
	q := eventdto.EventsQuery{
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
