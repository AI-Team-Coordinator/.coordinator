package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"relay/app"
	"relay/infra"
)

func main() {
	cfg := app.DefaultConfig()
	flag.StringVar(&cfg.Port, "port", cfg.Port, "HTTP port")
	flag.StringVar(&cfg.DataDir, "data", cfg.DataDir, "Common data dir (progress/external is written here)")
	flag.Parse()
	cfg.ApplyEnv()

	application, err := app.New(cfg)
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

	if err := application.Start("127.0.0.1:" + cfg.Port); err != nil {
		infra.LogError("server failed: %v", err)
		os.Exit(1)
	}
}
