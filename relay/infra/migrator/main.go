package main

import (
	"relay/infra"
	migrator "relay/infra/migrator/pkg"
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
)

// migrationsDir can be overridden by environment variable
var migrationsDir = "infra/migrator/migrations"

func main() {
	_ = infra.LoadEnv()
	infra.InitMigratorLog()

	// Check if we're running in Docker container
	if os.Getenv("MIGRATIONS_PATH") != "" {
		migrationsDir = os.Getenv("MIGRATIONS_PATH")
	}

	// Get migrations path
	var migrationsPath string
	if filepath.IsAbs(migrationsDir) {
		// If absolute path, use as is
		migrationsPath = migrationsDir
	} else {
		// If relative path, resolve from current working directory
		projectRoot, err := os.Getwd()
		if err != nil {
			infra.Log.Error("failed to get current working directory", zap.Error(err))
			os.Exit(1)
		}
		migrationsPath = filepath.Join(projectRoot, migrationsDir)
	}

	// Check command line arguments
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "status":
			runStatus(migrationsPath)
		case "up":
			runNewMigrator(migrationsPath)
		case "fix-checksum":
			if len(os.Args) < 3 {
				infra.Log.Error("missing migration version for fix-checksum command")
				os.Exit(1)
			}
			versionStr := os.Args[2]
			runFixChecksum(migrationsPath, versionStr)
		default:
			infra.Log.Error("unknown command", zap.String("command", command))
			os.Exit(1)
		}
	} else {
		// Default behavior: run migrations
		runNewMigrator(migrationsPath)
	}
}

// runNewMigrator runs the new migration system
func runNewMigrator(migrationsPath string) {
	infra.Log.Info("Using new migration system", zap.String("migrations_path", migrationsPath))

	// Create new migrator
	migrator, err := migrator.NewMigratorV2(migrationsPath)
	if err != nil {
		infra.Log.Error("failed to create new migrator", zap.Error(err))
		os.Exit(1)
	}

	// Validate migrations
	if err := migrator.ValidateMigrations(); err != nil {
		infra.Log.Error("migration validation failed", zap.Error(err))
		os.Exit(1)
	}

	// Check for dirty migrations
	dirtyMigrations := migrator.GetDirtyMigrations()
	if len(dirtyMigrations) > 0 {
		infra.Log.Error("database is in dirty state",
			zap.Int("dirty_count", len(dirtyMigrations)),
			zap.Any("dirty_migrations", dirtyMigrations))
		os.Exit(1)
	}

	// Get pending migrations
	pendingMigrations := migrator.GetPendingMigrations()
	if len(pendingMigrations) == 0 {
		status := migrator.GetStatus()
		infra.Log.Info("database is up to date, no migrations needed", zap.Any("status", status))
		return
	}

	// Apply pending migrations
	infra.Log.Info("starting database migrations...",
		zap.Int("pending_count", len(pendingMigrations)),
		zap.String("migrations_path", migrationsPath))

	for _, migration := range pendingMigrations {
		infra.Log.Info("applying migration",
			zap.String("name", migration.Name),
			zap.Uint64("version", migration.Version),
			zap.String("filename", migration.Filename))

		if err := migrator.ApplyMigration(migration); err != nil {
			infra.Log.Error("failed to apply migration",
				zap.String("name", migration.Name),
				zap.Uint64("version", migration.Version),
				zap.Error(err))
			os.Exit(1)
		}
	}

	status := migrator.GetStatus()
	infra.Log.Info("migrations completed successfully", zap.Any("status", status))
}

// runFixChecksum updates the checksum for a specific migration
func runFixChecksum(migrationsPath string, versionStr string) {
	infra.Log.Info("Attempting to fix checksum",
		zap.String("migrations_path", migrationsPath),
		zap.String("version", versionStr))

	version, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		infra.Log.Error("invalid migration version format", zap.Error(err))
		os.Exit(1)
	}

	migrator, err := migrator.NewMigratorV2(migrationsPath)
	if err != nil {
		infra.Log.Error("failed to create new migrator", zap.Error(err))
		os.Exit(1)
	}

	if err := migrator.FixChecksum(version); err != nil {
		infra.Log.Error("failed to fix checksum",
			zap.String("version", versionStr),
			zap.Error(err))
		os.Exit(1)
	}

	infra.Log.Info("Checksum fixed successfully", zap.String("version", versionStr))
}

// runStatus shows migration status
func runStatus(migrationsPath string) {
	infra.Log.Info("Checking migration status", zap.String("migrations_path", migrationsPath))

	// Create new migrator
	migrator, err := migrator.NewMigratorV2(migrationsPath)
	if err != nil {
		infra.Log.Error("failed to create new migrator", zap.Error(err))
		os.Exit(1)
	}

	// Validate migrations
	if err := migrator.ValidateMigrations(); err != nil {
		infra.Log.Error("migration validation failed", zap.Error(err))
		os.Exit(1)
	}

	// Check for dirty migrations
	dirtyMigrations := migrator.GetDirtyMigrations()
	if len(dirtyMigrations) > 0 {
		infra.Log.Error("database is in dirty state",
			zap.Int("dirty_count", len(dirtyMigrations)),
			zap.Any("dirty_migrations", dirtyMigrations))
		os.Exit(1)
	}

	// Get status information
	status := migrator.GetStatus()
	applied := migrator.GetAppliedMigrations()
	pending := migrator.GetPendingMigrations()

	// Log status information
	infra.Log.Info("Migration status",
		zap.Int("total_files", status["total_files"].(int)),
		zap.Int("applied", status["applied"].(int)),
		zap.Int("pending", status["pending"].(int)),
		zap.Int("dirty", status["dirty"].(int)),
		zap.Uint64("latest_version", status["latest_version"].(uint64)),
		zap.String("latest_migration", status["latest_migration"].(string)))

	if len(pending) > 0 {
		infra.Log.Info("Pending migrations found", zap.Int("count", len(pending)))
		for _, migration := range pending {
			infra.Log.Info("Pending migration",
				zap.Uint64("version", migration.Version),
				zap.String("name", migration.Name),
				zap.String("filename", migration.Filename))
		}
	} else {
		infra.Log.Info("No pending migrations")
	}

	if len(applied) > 0 {
		infra.Log.Info("Applied migrations", zap.Int("count", len(applied)))
		for _, migration := range applied {
			infra.Log.Info("Applied migration",
				zap.Uint64("version", migration.Version),
				zap.String("name", migration.Name),
				zap.String("filename", migration.Filename),
				zap.Time("applied_at", migration.AppliedAt))
		}
	}
}






