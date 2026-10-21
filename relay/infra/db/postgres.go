package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"

	"relay/infra"
	migrator "relay/infra/migrator/pkg"
)

func resolveMigrationsDir(dir string) string {
	if p := infra.GetEnv("RELAY_MIGRATIONS_PATH"); p != "" {
		return p
	}
	if dir == "" {
		dir = "infra/migrator/migrations"
	}
	if filepath.IsAbs(dir) {
		return dir
	}
	wd, err := os.Getwd()
	if err != nil {
		return dir
	}
	return filepath.Join(wd, dir)
}

// Open connects to Relay Postgres, runs pending migrations, returns a pool.
func Open(migrationsDir string) (*sql.DB, error) {
	dsn, err := infra.GetDsn()
	if err != nil {
		return nil, err
	}
	migrationsDir = resolveMigrationsDir(migrationsDir)
	if err := migrator.RunPendingMigrationsWithDSN(dsn, migrationsDir); err != nil {
		return nil, fmt.Errorf("relay migrations: %w", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
