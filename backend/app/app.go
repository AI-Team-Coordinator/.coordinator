package app

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"coordinator/domain/coordinator"
	"coordinator/domain/coordinator/repository"
	"coordinator/infra"
	"coordinator/infra/db"
)

type App struct {
	config                *Config
	httpServer            *http.Server
	handler               http.Handler
	eventStore            repository.EventStore
	repo                  *repository.FileRepository
	CoordinatorService    *coordinator.Service
	CoordinatorController *coordinator.Controller
}

func NewApp(config *Config) (*App, error) {
	if config == nil {
		config = DefaultConfig()
	}

	repo := repository.NewFileRepositoryWithPaths(repository.Paths{
		Workspace: config.Workspace,
		Data:      config.DataDir,
		Docs:      config.DocsDir,
		Bus:       config.BusDir,
		Cursor:    config.CursorDir,
		App:       config.AppRoot,
	})
	store, err := db.Open(
		filepath.Join(config.DataDir, ".cache"),
		filepath.Join(config.DataDir, "progress", "events"),
		filepath.Join(config.BusDir, "progress", "events"),
	)
	if err != nil {
		return nil, fmt.Errorf("open event store: %w", err)
	}
	if err := store.Sync(context.Background()); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("index jsonl: %w", err)
	}
	infra.LogInfo("event store ready (sqlite cache under data/.cache)")

	svc := coordinator.NewService(repo, store)
	ctrl := coordinator.NewController(svc)

	application := &App{
		config:                config,
		eventStore:            store,
		repo:                  repo,
		CoordinatorService:    svc,
		CoordinatorController: ctrl,
	}
	application.SetupRoutes()
	return application, nil
}

func (a *App) Start(addr string) error {
	go a.pullCommonLoop()
	a.httpServer = &http.Server{
		Addr:              addr,
		Handler:           a.handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	infra.LogInfo("starting API server on http://%s", addr)
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

const commonPullInterval = 15 * time.Second

func (a *App) pullCommonLoop() {
	if a.repo == nil {
		return
	}
	infra.LogInfo("common origin pull every %s", commonPullInterval)
	a.pullCommonOnce()
	ticker := time.NewTicker(commonPullInterval)
	defer ticker.Stop()
	for range ticker.C {
		a.pullCommonOnce()
	}
}

func (a *App) pullCommonOnce() {
	updated, err := a.repo.PullCommonOrigin(context.Background())
	if err != nil {
		infra.LogWarn("common pull origin/main: %v", err)
	} else if updated {
		infra.LogInfo("common pulled origin/main")
	}
	if err := a.repo.PullCoordinatorState(context.Background()); err != nil {
		infra.LogWarn("coordinator-state pull: %v", err)
	}
}

func (a *App) Shutdown() error {
	if a.eventStore != nil {
		_ = a.eventStore.Close()
	}
	if a.httpServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.httpServer.Shutdown(ctx)
}
