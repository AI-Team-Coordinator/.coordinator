package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"coordinator/app"
	"coordinator/infra"
)

func main() {
	cfg := app.DefaultConfig()
	flag.StringVar(&cfg.Port, "port", "4321", "HTTP server port")
	flag.StringVar(&cfg.Workspace, "workspace", "", "Path to workspace root")
	flag.StringVar(&cfg.DataDir, "data", "", "Path to data directory (settings, progress, logs)")
	flag.StringVar(&cfg.DocsDir, "docs", "", "Path to docs directory")
	flag.StringVar(&cfg.BusDir, "bus", "", "Path to git data-bus repo (settings sync / pull)")
	flag.StringVar(&cfg.CursorDir, "cursor", "", "Path to .cursor directory")
	flag.StringVar(&cfg.FrontendDir, "frontend", "", "Path to local UI static assets")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		infra.LogError("getwd: %v", err)
		os.Exit(1)
	}
	cfg.ResolvePaths(cwd)

	infra.LogInfo("App root:   %s", cfg.AppRoot)
	infra.LogInfo("Workspace:  %s", cfg.Workspace)
	infra.LogInfo("Data dir:   %s", cfg.DataDir)
	infra.LogInfo("Docs dir:   %s", cfg.DocsDir)
	infra.LogInfo("Bus dir:    %s", cfg.BusDir)
	infra.LogInfo("Cursor dir: %s", cfg.CursorDir)
	infra.LogInfo("UI dir:     %s", cfg.FrontendDir)

	application, err := app.NewApp(cfg)
	if err != nil {
		infra.LogError("create app: %v", err)
		os.Exit(1)
	}

	setupGracefulShutdown(application)

	if err := application.Start("127.0.0.1:" + cfg.Port); err != nil {
		infra.LogError("server failed: %v", err)
		os.Exit(1)
	}
}

func setupGracefulShutdown(application *app.App) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		infra.LogInfo("shutting down")
		_ = application.Shutdown()
	}()
}
