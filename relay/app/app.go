package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"relay/domain/relay"
	"relay/domain/relay/bus"
	"relay/domain/relay/github"
	"relay/infra"
)

type Config struct {
	Port                string
	DataDir             string
	GitHubWebhookSecret string
	BearerToken         string
}

func DefaultConfig() Config {
	return Config{Port: "4322", DataDir: "data"}
}

func (c *Config) ApplyEnv() {
	if v := os.Getenv("RELAY_GITHUB_WEBHOOK_SECRET"); v != "" {
		c.GitHubWebhookSecret = v
	}
	if v := os.Getenv("RELAY_DATA"); v != "" {
		c.DataDir = v
	}
}

type App struct {
	config Config
	server *http.Server
}

func New(cfg Config) (*App, error) {
	if cfg.DataDir == "" {
		return nil, fmt.Errorf("data dir is empty")
	}
	if strings.TrimSpace(cfg.BearerToken) == "" {
		token, err := infra.LoadOrCreateBearerToken(cfg.DataDir)
		if err != nil {
			return nil, err
		}
		cfg.BearerToken = token
	}
	svc := &relay.Service{
		GitHub: github.Adapter{WebhookSecret: cfg.GitHubWebhookSecret},
		Bus:    bus.JSONL{Root: cfg.DataDir},
	}
	ctrl := &relay.Controller{Service: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /github/webhook", ctrl.GitHubWebhook)

	api := http.NewServeMux()
	api.HandleFunc("GET /health", ctrl.Health)
	api.HandleFunc("GET /sources", ctrl.Sources)
	protected := infra.BearerAuth(cfg.BearerToken, cfg.DataDir, api)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", protected))

	return &App{
		config: cfg,
		server: &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func (a *App) Start(addr string) error {
	a.server.Addr = addr
	infra.LogInfo("starting relay on http://%s", addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.server.Shutdown(ctx)
}
