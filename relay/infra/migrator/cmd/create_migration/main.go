package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// createMigration creates a new migration file with the given name
func createMigration(name string) error {
	// Get current timestamp for migration version (YYYYMMDDHHMMSS format)
	timestamp := time.Now().Format("20060102150405")

	// Create migration filename
	upFilename := fmt.Sprintf("%s_%s.up.sql", timestamp, name)
	downFilename := fmt.Sprintf("%s_%s.down.sql", timestamp, name)

	// Get migrations directory
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	migrationsDir := filepath.Join(projectRoot, "infra/migrator/migrations")

	// Ensure migrations directory exists
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create up migration file
	upPath := filepath.Join(migrationsDir, upFilename)
	upContent := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: Add your migration SQL here

-- Example:
-- ALTER TABLE table_name
-- ADD COLUMN column_name TYPE;

`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	downPath := filepath.Join(migrationsDir, downFilename)
	downContent := fmt.Sprintf(`-- Down migration: %s
-- Created: %s
-- Description: Add rollback SQL here

-- Example:
-- ALTER TABLE table_name
-- DROP COLUMN IF EXISTS column_name;

`, name, time.Now().Format("2006-01-02 15:04:05"))

	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files:\n")
	fmt.Printf("  Up:   %s\n", upPath)
	fmt.Printf("  Down: %s\n", downPath)
	fmt.Printf("\nMigration version: %s\n", timestamp)

	return nil
}

// Usage example:
// go run infra/migrator/cmd/create_migration/main.go add_user_status_column
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run infra/migrator/cmd/create_migration/main.go <migration_name>")
		fmt.Println("Example: go run infra/migrator/cmd/create_migration/main.go add_user_status_column")
		os.Exit(1)
	}

	migrationName := os.Args[1]

	if err := createMigration(migrationName); err != nil {
		fmt.Printf("Error creating migration: %v\n", err)
		os.Exit(1)
	}
}






