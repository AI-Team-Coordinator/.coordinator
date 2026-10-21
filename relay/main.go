package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"relay/app"
	"relay/connectors/github"
	"relay/infra"
)

func main() {
	cfg := app.DefaultConfig()
	flag.StringVar(&cfg.Port, "port", cfg.Port, "HTTP port")
	flag.StringVar(&cfg.DataDir, "data", cfg.DataDir, "Common data dir (progress/external is written here)")
	flag.Parse()
	_ = infra.LoadEnv()
	infra.InitMigratorLog()
	cfg.ApplyEnv()
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		cfg.Port = p
	}

	application, err := app.New(cfg, github.Options()...)
	if err != nil {
		infra.LogError("create app: %v", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		infra.LogInfo("shutting down")
		_ = application.Shutdown()
	}()

	listen := os.Getenv("RELAY_LISTEN")
	if listen == "" {
		listen = "127.0.0.1:" + cfg.Port
	}
	if err := application.Start(listen); err != nil {
		infra.LogError("server failed: %v", err)
		os.Exit(1)
	}
}
