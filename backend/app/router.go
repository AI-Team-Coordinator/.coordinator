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
	mux.HandleFunc("GET "+api+"/team", a.CoordinatorController.GetTeam)
	mux.HandleFunc("PUT "+api+"/team", a.CoordinatorController.PutTeam)
	mux.HandleFunc("GET "+api+"/project", a.CoordinatorController.GetProject)
	mux.HandleFunc("POST "+api+"/project/services", a.CoordinatorController.PostCreateService)
	mux.HandleFunc("GET "+api+"/sync", a.CoordinatorController.GetSyncStatus)
	mux.HandleFunc("POST "+api+"/sync", a.CoordinatorController.PostSync)
	mux.HandleFunc("GET "+api+"/members", a.CoordinatorController.GetMembers)
	mux.HandleFunc("GET "+api+"/conflicts", a.CoordinatorController.GetConflicts)
	mux.HandleFunc("GET "+api+"/stats", a.CoordinatorController.GetStats)
	mux.HandleFunc("GET "+api+"/events", a.CoordinatorController.GetEvents)
	mux.HandleFunc("GET "+api+"/tasks", a.CoordinatorController.GetTasks)
	mux.HandleFunc("GET "+api+"/docs/{taskID}", a.CoordinatorController.GetTaskDoc)
	mux.HandleFunc("GET "+api+"/stream", a.CoordinatorController.Stream)

	a.mountStatic(mux)
	a.handler = CORSMiddleware(LoggingMiddleware(RecoveryMiddleware(mux)))
}

func (a *App) mountStatic(mux *http.ServeMux) {
	if a.config.WebDir == "" {
		return
	}
	indexPath := filepath.Join(a.config.WebDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<h2>Coordinator API</h2><p>See <a href='/api/v1/health'>/api/v1/health</a></p>")
		})
		return
	}

	fileServer := http.FileServer(http.Dir(a.config.WebDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}
		targetPath := filepath.Join(a.config.WebDir, filepath.Clean(r.URL.Path))
		if fi, err := os.Stat(targetPath); err == nil && !fi.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}
