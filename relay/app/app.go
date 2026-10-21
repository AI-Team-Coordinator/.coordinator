package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"relay/domain/activity"
	"relay/domain/config"
	"relay/domain/connector"
	"relay/domain/gateway"
	"relay/domain/marketplace"
	"relay/domain/poll"
	"relay/domain/relay"
	"relay/domain/relay/bus"
	"relay/infra"
	"relay/infra/crypto"
	"relay/infra/db"
	"relay/infra/settings"
	"relay/infra/statebus"
	"relay/infra/store"
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
	config        Config
	server        *http.Server
	pg            interface{ Close() error }
	pollCancel    context.CancelFunc
	gateway       *gateway.Gateway
	runtime       *config.Runtime
	relaySettings *settings.RelayConnectors
	pollFactories []func(*gateway.Gateway, *activity.Recorder) poll.Job
	activityLog   *activity.Recorder
	stateCfg      statebus.Config
	stateSink     *statebus.Sink
	outboxStore   *store.OutboxStore
}

func New(cfg Config, opts ...Option) (*App, error) {
	inst := applyOptions(opts)
	for _, e := range inst.catalog {
		marketplace.RegisterCatalog(e)
	}

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

	pgDB, err := db.Open("infra/migrator/migrations")
	if err != nil {
		return nil, err
	}
	key, err := crypto.LoadConfigKey(cfg.DataDir)
	if err != nil {
		_ = pgDB.Close()
		return nil, err
	}
	secretStore, err := store.NewSecretStore(pgDB, key)
	if err != nil {
		_ = pgDB.Close()
		return nil, err
	}
	eventStore := store.NewEventStore(pgDB)
	activityStore := store.NewActivityStore(pgDB)
	activityRecorder := &activity.Recorder{Store: activityStore}
	outboxStore := store.NewOutboxStore(pgDB)
	relaySettings := settings.NewRelayConnectors(cfg.DataDir)
	runtime := &config.Runtime{Settings: relaySettings, Secrets: secretStore}

	var stock []connector.Connector
	for _, init := range inst.connectorInits {
		if init != nil {
			stock = append(stock, init(runtime, cfg))
		}
	}
	stock = append(stock, inst.connectors...)
	reg, err := connector.NewRegistry(stock...)
	if err != nil {
		_ = pgDB.Close()
		return nil, err
	}
	var busSink gateway.Sink
	stateCfg := statebus.LoadFromEnv()
	stateSink := statebus.NewSink(stateCfg, cfg.DataDir)
	if stateSink != nil {
		infra.LogInfo("relay state bus: origin=%s worktree=%s (outbox)", stateCfg.OriginID, stateCfg.Worktree)
	} else if strings.TrimSpace(os.Getenv("RELAY_DEBUG_JSONL")) == "1" {
		busSink = bus.JSONL{Root: cfg.DataDir}
	}
	var outboxPtr *store.OutboxStore
	if stateSink != nil && outboxStore != nil {
		outboxPtr = outboxStore
	}
	gw := &gateway.Gateway{
		Registry: reg,
		Sink: gateway.CompositeSink{
			Bus:    busSink,
			Store:  eventStore,
			Outbox: outboxPtr,
		},
		IsEnabled: func(source string) bool {
			return runtime.Enabled(source)
		},
	}
	mktSvc := &marketplace.Service{
		Runtime:  runtime,
		Settings: relaySettings,
		Secrets:  secretStore,
		Activity: activityStore,
	}
	var pollJobs []poll.Job
	for _, factory := range inst.pollFactories {
		if factory == nil {
			continue
		}
		pollJobs = append(pollJobs, factory(gw, activityRecorder))
	}
	actionReg := connector.NewActionRegistry()
	for _, reg := range inst.actionRegs {
		if reg != nil {
			reg(actionReg, gw, runtime, relaySettings, pollJobs)
		}
	}
	for _, job := range pollJobs {
		if job.Run == nil || job.SourceID == "" {
			continue
		}
		src := job.SourceID
		actionReg.Register(src, connector.Action{
			ID:    connector.ActionPollNow,
			Label: "Run scheduled check now",
			Run: func(ctx context.Context) error {
				ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
				defer cancel()
				j, ok := poll.FindJob(pollJobs, src)
				if !ok {
					return fmt.Errorf("poll job missing")
				}
				return poll.RunNow(ctx, runtime, relaySettings, j)
			},
		})
	}
	mktCtrl := &marketplace.Controller{Service: mktSvc, Actions: actionReg}

	svc := &relay.Service{
		Gateway:  gw,
		Runtime:  runtime,
		Registry: reg,
	}
	ctrl := &relay.Controller{Service: svc}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /connectors/{source}/webhook", ctrl.ConnectorWebhook)
	mux.HandleFunc("POST /github/webhook", ctrl.GitHubWebhook)

	api := http.NewServeMux()
	api.HandleFunc("GET /sources", ctrl.Sources)
	api.HandleFunc("POST /connectors/{source}/events", ctrl.ConnectorEvents)
	api.HandleFunc("GET /connectors/marketplace", mktCtrl.Marketplace)
	api.HandleFunc("GET /connectors/activity", mktCtrl.ListActivity)
	api.HandleFunc("GET /connectors/{id}/config", mktCtrl.GetConfig)
	api.Handle("PUT /connectors/{id}/config", infra.MasterBearerOnly(cfg.BearerToken, http.HandlerFunc(mktCtrl.PutConfig)))
	api.Handle("POST /connectors/{id}/disable", infra.MasterBearerOnly(cfg.BearerToken, http.HandlerFunc(mktCtrl.Disable)))
	api.Handle("POST /connectors/{id}/actions/{action}", infra.MasterBearerOnly(cfg.BearerToken, http.HandlerFunc(mktCtrl.RunAction)))

	protected := infra.BearerAuth(cfg.BearerToken, cfg.DataDir, api)
	mux.HandleFunc("GET /api/v1/health", ctrl.Health)
	mux.HandleFunc("GET /health", ctrl.Health)
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", protected))

	return &App{
		config:        cfg,
		pg:            pgDB,
		gateway:       gw,
		runtime:       runtime,
		relaySettings: relaySettings,
		pollFactories: inst.pollFactories,
		activityLog:   activityRecorder,
		stateCfg:      stateCfg,
		stateSink:     stateSink,
		outboxStore:   outboxPtr,
		server: &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

func (a *App) Start(addr string) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.pollCancel = cancel
	if a.gateway != nil && a.runtime != nil && a.relaySettings != nil {
		var jobs []poll.Job
		for _, factory := range a.pollFactories {
			if factory == nil {
				continue
			}
			jobs = append(jobs, factory(a.gateway, a.activityLog))
		}
		poll.Start(ctx, a.gateway, a.runtime, a.relaySettings, a.activityLog, jobs)
	}
	if a.stateSink != nil && a.outboxStore != nil {
		go statebus.RunWorker(ctx, a.stateCfg, a.outboxStore, a.stateSink)
	}
	a.server.Addr = addr
	infra.LogInfo("starting relay on http://%s", addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *App) Shutdown() error {
	if a.pollCancel != nil {
		a.pollCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := a.server.Shutdown(ctx)
	if a.pg != nil {
		_ = a.pg.Close()
	}
	return err
}
