package infra

import (
	"go.uber.org/zap"
)

// Log is used by the migrator (MigratorV2). Initialized in main before DB migrate.
var Log *zap.Logger

func InitMigratorLog() {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	l, err := cfg.Build()
	if err != nil {
		l = zap.NewNop()
	}
	Log = l
}
