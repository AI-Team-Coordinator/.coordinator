package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func (a *App) SetupRoutes() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", a.CoordinatorController.GetHealth)

	api := "/api/v1"
	mux.HandleFunc("GET "+api+"/health", a.CoordinatorController.GetHealth)
	mux.HandleFunc("GET "+api+"/pulse", a.CoordinatorController.GetPulse)
	mux.HandleFunc("GET "+api+"/external/events", a.CoordinatorController.GetExternalEvents)
	mux.HandleFunc("GET "+api+"/team", a.CoordinatorController.GetTeam)
	mux.HandleFunc("PUT "+api+"/team", a.CoordinatorController.PutTeam)
	mux.HandleFunc("GET "+api+"/org-topology", a.CoordinatorController.GetOrgTopology)
	mux.HandleFunc("PUT "+api+"/org-topology", a.CoordinatorController.PutOrgTopology)
	mux.HandleFunc("GET "+api+"/project", a.CoordinatorController.GetProject)
	mux.HandleFunc("GET "+api+"/setup", a.CoordinatorController.GetSetup)
	mux.HandleFunc("POST "+api+"/setup", a.CoordinatorController.PostSetup)
	mux.HandleFunc("GET "+api+"/me/settings", a.CoordinatorController.GetUserSettings)
	mux.HandleFunc("PUT "+api+"/me/settings", a.CoordinatorController.PutUserSettings)
	mux.HandleFunc("GET "+api+"/me/avatar", a.CoordinatorController.GetUserAvatar)
	mux.HandleFunc("POST "+api+"/me/avatar", a.CoordinatorController.PostUserAvatar)
	mux.HandleFunc("DELETE "+api+"/me/avatar", a.CoordinatorController.DeleteUserAvatar)
	mux.HandleFunc("GET "+api+"/relay/access", a.CoordinatorController.GetRelayAccess)
	mux.HandleFunc("POST "+api+"/relay/connect", a.CoordinatorController.PostRelayConnect)
	mux.HandleFunc("POST "+api+"/relay/disconnect", a.CoordinatorController.PostRelayDisconnect)
	mux.HandleFunc("GET "+api+"/relay/sources", a.CoordinatorController.GetRelaySources)
	mux.HandleFunc("GET "+api+"/relay/connectors/marketplace", a.CoordinatorController.GetRelayMarketplace)
	mux.HandleFunc("GET "+api+"/relay/connectors/activity", a.CoordinatorController.GetRelayConnectorActivity)
	mux.HandleFunc("GET "+api+"/relay/connectors/{id}/config", a.CoordinatorController.GetRelayConnectorConfig)
	mux.HandleFunc("PUT "+api+"/relay/connectors/{id}/config", a.CoordinatorController.PutRelayConnectorConfig)
	mux.HandleFunc("POST "+api+"/relay/connectors/{id}/disable", a.CoordinatorController.PostRelayConnectorDisable)
	mux.HandleFunc("POST "+api+"/relay/connectors/{id}/actions/{action}", a.CoordinatorController.PostRelayConnectorAction)
	mux.HandleFunc("GET "+api+"/relay/tokens", a.CoordinatorController.GetRelayIssuedTokens)
	mux.HandleFunc("POST "+api+"/relay/tokens", a.CoordinatorController.PostRelayIssuedToken)
	mux.HandleFunc("DELETE "+api+"/relay/tokens/{id}", a.CoordinatorController.DeleteRelayIssuedToken)
	mux.HandleFunc("POST "+api+"/project/services", a.CoordinatorController.PostCreateService)
	mux.HandleFunc("GET "+api+"/sync", a.CoordinatorController.GetSyncStatus)
	mux.HandleFunc("POST "+api+"/sync", a.CoordinatorController.PostSync)
	mux.HandleFunc("GET "+api+"/members", a.CoordinatorController.GetMembers)
	mux.HandleFunc("GET "+api+"/conflicts", a.CoordinatorController.GetConflicts)
	mux.HandleFunc("GET "+api+"/stats", a.CoordinatorController.GetStats)
	mux.HandleFunc("GET "+api+"/events", a.CoordinatorController.GetEvents)
	mux.HandleFunc("GET "+api+"/tasks", a.CoordinatorController.GetTasks)
	mux.HandleFunc("GET "+api+"/work-items", a.CoordinatorController.GetWorkItems)
	mux.HandleFunc("GET "+api+"/docs/{taskID}", a.CoordinatorController.GetTaskDoc)
	mux.HandleFunc("GET "+api+"/stream", a.CoordinatorController.Stream)

	a.mountStatic(mux)
	a.handler = CORSMiddleware(LoggingMiddleware(RecoveryMiddleware(mux)))
}

func (a *App) mountStatic(mux *http.ServeMux) {
	apiOnlyStub := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h2>Coordinator API</h2><p>See <a href='/api/v1/health'>/api/v1/health</a>. The board is Vite on UI_PORT (Alina Assist).</p>")
	}
	if a.config.FrontendDir == "" {
		mux.HandleFunc("GET /{$}", apiOnlyStub)
		return
	}
	indexPath := filepath.Join(a.config.FrontendDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<h2>Coordinator API</h2><p>See <a href='/api/v1/health'>/api/v1/health</a></p>")
		})
		return
	}

	fileServer := http.FileServer(http.Dir(a.config.FrontendDir))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, indexPath)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}
		targetPath := filepath.Join(a.config.FrontendDir, filepath.Clean(r.URL.Path))
		if fi, err := os.Stat(targetPath); err == nil && !fi.IsDir() {
			if filepath.Base(targetPath) == "index.html" {
				w.Header().Set("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, indexPath)
	})
}
