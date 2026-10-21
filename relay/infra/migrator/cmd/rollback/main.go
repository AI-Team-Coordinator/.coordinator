package main

import (
	"relay/infra"
	migrator "relay/infra/migrator/pkg"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
)

func main() {
	_ = infra.LoadEnv()
	infra.InitMigratorLog()
	infra.LoadEnv()

	var (
		toVersionStr string
	)
	flag.StringVar(&toVersionStr, "to", "", "Rollback to version (timestamp, inclusive). If not specified — rollback all migrations.")
	flag.Parse()

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

	migrator, err := migrator.NewMigratorV2(migrationsPath)
	if err != nil {
		infra.Log.Error("failed to create migrator", zap.Error(err))
		os.Exit(1)
	}

	applied := migrator.GetAppliedMigrations()
	if len(applied) == 0 {
		fmt.Println("No applied migrations — nothing to rollback.")
		return
	}

	var toVersion uint64
	if toVersionStr != "" {
		parsed, err := strconv.ParseUint(toVersionStr, 10, 64)
		if err != nil {
			fmt.Printf("Error: version must be a number (timestamp): %v\n", err)
			os.Exit(1)
		}
		toVersion = parsed
	}

	// Rollback in reverse order
	for i := len(applied) - 1; i >= 0; i-- {
		record := applied[i]
		if toVersion != 0 && record.Version >= toVersion {
			fmt.Printf("Rolling back migration: %d %s (%s)\n", record.Version, record.Name, record.Filename)
			err := migrator.RollbackMigration(record)
			if err != nil {
				fmt.Printf("Error rolling back migration %d: %v\n", record.Version, err)
				os.Exit(1)
			}
		} else if toVersion != 0 && record.Version < toVersion {
			fmt.Printf("Reached target version %d, rollback completed.\n", toVersion)
			break
		} else if toVersion == 0 {
			// If target version is not specified, rollback all
			fmt.Printf("Rolling back migration: %d %s (%s)\n", record.Version, record.Name, record.Filename)
			err := migrator.RollbackMigration(record)
			if err != nil {
				fmt.Printf("Error rolling back migration %d: %v\n", record.Version, err)
				os.Exit(1)
			}
		}
	}

	fmt.Println("Migration rollback completed.")
}






