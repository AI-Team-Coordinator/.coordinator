package main

import (
	"relay/infra"
	migrator "relay/infra/migrator/pkg"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func main() {
	_ = infra.LoadEnv()
	infra.InitMigratorLog()
	infra.LoadEnv()

	// Get migrations path
	migrationsDir := "infra/migrator/migrations"
	if os.Getenv("MIGRATIONS_PATH") != "" {
		migrationsDir = os.Getenv("MIGRATIONS_PATH")
	}

	var migrationsPath string
	if filepath.IsAbs(migrationsDir) {
		migrationsPath = migrationsDir
	} else {
		projectRoot, err := os.Getwd()
		if err != nil {
			infra.Log.Error("failed to get current working directory", zap.Error(err))
			os.Exit(1)
		}
		migrationsPath = filepath.Join(projectRoot, migrationsDir)
	}

	// Create new migrator
	migrator, err := migrator.NewMigratorV2(migrationsPath)
	if err != nil {
		infra.Log.Error("failed to create migrator", zap.Error(err))
		os.Exit(1)
	}

	// Get status
	status := migrator.GetStatus()
	applied := migrator.GetAppliedMigrations()
	pending := migrator.GetPendingMigrations()
	dirty := migrator.GetDirtyMigrations()

	// Print status
	fmt.Println("=== Migration Status ===")
	fmt.Printf("Total migration files: %d\n", status["total_files"])
	fmt.Printf("Applied migrations:   %d\n", status["applied"])
	fmt.Printf("Pending migrations:   %d\n", status["pending"])
	fmt.Printf("Dirty migrations:     %d\n", status["dirty"])
	fmt.Printf("Latest version:       %d\n", status["latest_version"])
	fmt.Printf("Latest migration:     %s\n", status["latest_migration"])

	if len(dirty) > 0 {
		fmt.Println("\n=== Dirty Migrations ===")
		for _, migration := range dirty {
			fmt.Printf("  Version %d: %s (%s)\n", migration.Version, migration.Name, migration.Filename)
		}
	}

	if len(pending) > 0 {
		fmt.Println("\n=== Pending Migrations ===")
		for _, migration := range pending {
			fmt.Printf("  Version %d: %s (%s)\n", migration.Version, migration.Name, migration.Filename)
		}
	}

	if len(applied) > 0 {
		fmt.Println("\n=== Applied Migrations ===")
		for _, migration := range applied {
			fmt.Printf("  Version %d: %s (%s) - %s\n", 
				migration.Version, 
				migration.Name, 
				migration.Filename,
				migration.AppliedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Validate migrations
	if err := migrator.ValidateMigrations(); err != nil {
		fmt.Println("\n=== Validation Errors ===")
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("\n✅ All migrations are valid")
	}
}






