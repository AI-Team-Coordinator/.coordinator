package migrator

import (
	"relay/infra"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	ID              uint   `gorm:"primaryKey"`
	Version         uint64 `gorm:"uniqueIndex;not null"`
	Name            string `gorm:"index"`
	Filename        string `gorm:"index"`
	AppliedAt       time.Time
	Dirty           bool
	Checksum        string
	ExecutionTimeMs int `gorm:"column:execution_time_ms"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// MigrationFile represents a migration file on disk
type MigrationFile struct {
	Version   uint64
	Name      string
	Filename  string
	FilePath  string
	Direction string // "up" or "down"
	Checksum  string
}

// MigratorV2 represents the new migration system
type MigratorV2 struct {
	db            *gorm.DB
	migrationsDir string
	migrations    []MigrationFile
	records       []MigrationRecord
}

// NewMigratorV2 creates a new migrator instance (uses infra.GetDsn)
func NewMigratorV2(migrationsDir string) (*MigratorV2, error) {
	dsn, err := infra.GetDsn()
	if err != nil {
		return nil, fmt.Errorf("failed to get DSN: %w", err)
	}
	return NewMigratorV2WithDSN(dsn, migrationsDir)
}

// NewMigratorV2WithDSN creates a migrator with custom DSN (for tests)
func NewMigratorV2WithDSN(dsn, migrationsDir string) (*MigratorV2, error) {
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := ensureMigrationsTable(db); err != nil {
		return nil, fmt.Errorf("failed to ensure extended schema: %w", err)
	}

	m := &MigratorV2{
		db:            db,
		migrationsDir: migrationsDir,
	}

	if err := m.loadMigrations(); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	if err := m.loadRecords(); err != nil {
		return nil, fmt.Errorf("failed to load records: %w", err)
	}

	return m, nil
}

// ensureMigrationsTable ensures the _migrations table exists
func ensureMigrationsTable(db *gorm.DB) error {
	// Check if table exists
	var tableExists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = '_migrations' AND table_schema = 'public'
		)
	`).Scan(&tableExists).Error
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	if tableExists {
		infra.Log.Info("_migrations table already exists")
		return nil
	}

	infra.Log.Info("Creating _migrations table...")

	// Create table from scratch
	createTableQuery := `
		CREATE TABLE _migrations (
			id BIGSERIAL PRIMARY KEY,
			version BIGINT NOT NULL UNIQUE,
			name VARCHAR(255),
			filename VARCHAR(255),
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			dirty BOOLEAN DEFAULT FALSE,
			checksum VARCHAR(64),
			execution_time_ms INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	if err := db.Exec(createTableQuery).Error; err != nil {
		return fmt.Errorf("failed to create _migrations table: %w", err)
	}

	// Create indexes
	indexQueries := []string{
		"CREATE INDEX idx__migrations_version ON _migrations(version)",
		"CREATE INDEX idx__migrations_name ON _migrations(name)",
		"CREATE INDEX idx__migrations_filename ON _migrations(filename)",
	}

	for _, query := range indexQueries {
		if err := db.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to create index '%s': %w", query, err)
		}
	}

	infra.Log.Info("_migrations table created successfully")
	return nil
}

// loadMigrations loads migration files from disk
func (m *MigratorV2) loadMigrations() error {
	files, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Regex to match migration files: YYYYMMDDHHMMSS_name.up.sql or YYYYMMDDHHMMSS_name.down.sql
	re := regexp.MustCompile(`^(\d{14})_(.+)\.(up|down)\.sql$`)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matches := re.FindStringSubmatch(file.Name())
		if len(matches) != 4 {
			infra.Log.Debug("Skipping non-migration file", zap.String("filename", file.Name()))
			continue
		}

		timestamp := matches[1]
		name := matches[2]
		direction := matches[3]

		version, err := strconv.ParseUint(timestamp, 10, 64)
		if err != nil {
			infra.Log.Warn("Failed to parse migration timestamp",
				zap.String("filename", file.Name()),
				zap.String("timestamp", timestamp),
				zap.Error(err))
			continue
		}

		filePath := filepath.Join(m.migrationsDir, file.Name())
		checksum, err := calculateFileChecksum(filePath)
		if err != nil {
			infra.Log.Warn("Failed to calculate file checksum",
				zap.String("filename", file.Name()),
				zap.Error(err))
			continue
		}

		migration := MigrationFile{
			Version:   version,
			Name:      name,
			Filename:  file.Name(),
			FilePath:  filePath,
			Direction: direction,
			Checksum:  checksum,
		}

		m.migrations = append(m.migrations, migration)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	infra.Log.Info("Loaded migration files", zap.Int("count", len(m.migrations)))
	return nil
}

// loadRecords loads migration records from database
func (m *MigratorV2) loadRecords() error {
	var records []MigrationRecord
	if err := m.db.Find(&records).Error; err != nil {
		return fmt.Errorf("failed to load migration records: %w", err)
	}

	m.records = records
	infra.Log.Info("Loaded migration records", zap.Int("count", len(records)))
	return nil
}

// calculateFileChecksum calculates SHA256 checksum of a file
func calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetPendingMigrations returns migrations that need to be applied
func (m *MigratorV2) GetPendingMigrations() []MigrationFile {
	var pending []MigrationFile

	// Create a map of applied versions
	appliedVersions := make(map[uint64]bool)
	for _, record := range m.records {
		if !record.Dirty {
			appliedVersions[record.Version] = true
		}
	}

	// Find up migrations that haven't been applied
	for _, migration := range m.migrations {
		if migration.Direction == "up" && !appliedVersions[migration.Version] {
			pending = append(pending, migration)
		}
	}

	// Sort by version
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})

	return pending
}

// GetAppliedMigrations returns migrations that have been applied
func (m *MigratorV2) GetAppliedMigrations() []MigrationRecord {
	var applied []MigrationRecord
	for _, record := range m.records {
		if !record.Dirty {
			applied = append(applied, record)
		}
	}

	// Sort by version
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].Version < applied[j].Version
	})

	return applied
}

// GetDirtyMigrations returns migrations that failed to apply
func (m *MigratorV2) GetDirtyMigrations() []MigrationRecord {
	var dirty []MigrationRecord
	for _, record := range m.records {
		if record.Dirty {
			dirty = append(dirty, record)
		}
	}
	return dirty
}

// ApplyMigration applies a single migration
func (m *MigratorV2) ApplyMigration(migration MigrationFile) error {
	startTime := time.Now()

	// Read migration file
	content, err := os.ReadFile(migration.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Execute migration SQL within transaction
	if err := tx.Exec(string(content)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record successful migration within transaction
	executionTime := int(time.Since(startTime).Milliseconds())
	if err := m.recordMigrationInTx(tx, migration, executionTime, false); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	infra.Log.Info("Applied migration successfully",
		zap.String("name", migration.Name),
		zap.Uint64("version", migration.Version),
		zap.String("filename", migration.Filename),
		zap.Int("executionTimeMs", executionTime))

	return nil
}

// recordMigrationInTx records a migration within a transaction
func (m *MigratorV2) recordMigrationInTx(tx *gorm.DB, migration MigrationFile, executionTime int, dirty bool) error {
	record := MigrationRecord{
		Version:         migration.Version,
		Name:            migration.Name,
		Filename:        migration.Filename,
		AppliedAt:       time.Now(),
		Dirty:           dirty,
		Checksum:        migration.Checksum,
		ExecutionTimeMs: executionTime,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Upsert record within transaction
	return tx.Where("version = ?", migration.Version).Assign(record).FirstOrCreate(&record).Error
}

// RollbackMigration rolls back a single migration
func (m *MigratorV2) RollbackMigration(record MigrationRecord) error {
	// Find corresponding down migration
	var downMigration *MigrationFile
	for _, migration := range m.migrations {
		if migration.Version == record.Version && migration.Direction == "down" {
			downMigration = &migration
			break
		}
	}

	if downMigration == nil {
		return fmt.Errorf("down migration not found for version %d", record.Version)
	}

	startTime := time.Now()

	// Read migration file
	content, err := os.ReadFile(downMigration.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read down migration file: %w", err)
	}

	// Start transaction
	tx := m.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Execute rollback SQL within transaction
	if err := tx.Exec(string(content)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	// Remove record within transaction
	if err := tx.Delete(&record).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	infra.Log.Info("Rolled back migration successfully",
		zap.String("name", record.Name),
		zap.Uint64("version", record.Version),
		zap.Int("executionTimeMs", executionTime))

	return nil
}

// ValidateMigrations validates that all applied migrations match their files
func (m *MigratorV2) ValidateMigrations() error {
	var errors []string

	for _, record := range m.records {
		// Find corresponding file
		var foundFile *MigrationFile
		for _, migration := range m.migrations {
			if migration.Version == record.Version && migration.Direction == "up" {
				foundFile = &migration
				break
			}
		}

		if foundFile == nil {
			errors = append(errors, fmt.Sprintf("Migration file not found for version %d (%s)", record.Version, record.Name))
			continue
		}

		// Check checksum
		if record.Checksum != "" && foundFile.Checksum != record.Checksum {
			errors = append(errors, fmt.Sprintf("Checksum mismatch for version %d (%s)", record.Version, record.Name))
		}

		// Check filename
		if record.Filename != foundFile.Filename {
			errors = append(errors, fmt.Sprintf("Filename mismatch for version %d: DB=%s, File=%s", record.Version, record.Filename, foundFile.Filename))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("migration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// FixChecksum updates the checksum for a given migration version in the database
func (m *MigratorV2) FixChecksum(version uint64) error {
	var record MigrationRecord
	// Find the record in the database
	if err := m.db.Where("version = ?", version).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("migration record not found for version %d", version)
		}
		return fmt.Errorf("failed to find migration record: %w", err)
	}

	// Find corresponding file
	var foundFile *MigrationFile
	for i := range m.migrations {
		if m.migrations[i].Version == version && m.migrations[i].Direction == "up" {
			foundFile = &m.migrations[i]
			break
		}
	}

	if foundFile == nil {
		return fmt.Errorf("migration file not found for version %d", version)
	}

	if record.Checksum == foundFile.Checksum {
		infra.Log.Info("Checksum already matches, no update needed", zap.Uint64("version", version))
		return nil
	}

	infra.Log.Info("Updating checksum",
		zap.Uint64("version", version),
		zap.String("old_checksum", record.Checksum),
		zap.String("new_checksum", foundFile.Checksum))

	// Update the checksum in the database
	result := m.db.Model(&MigrationRecord{}).Where("version = ?", version).Update("checksum", foundFile.Checksum)
	if result.Error != nil {
		return fmt.Errorf("failed to update checksum in database: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows were updated for version %d", version)
	}

	return nil
}

// RunPendingMigrations runs all pending migrations (DSN from infra.GetDsn).
// This function is designed to be called during application startup.
func RunPendingMigrations(migrationsDir string) error {
	dsn, err := infra.GetDsn()
	if err != nil {
		return fmt.Errorf("failed to get database DSN: %w", err)
	}
	return RunPendingMigrationsWithDSN(dsn, migrationsDir)
}

// RunPendingMigrationsWithDSN runs pending migrations against the given DSN.
func RunPendingMigrationsWithDSN(dsn, migrationsDir string) error {
	if infra.Log != nil {
		infra.Log.Info("Checking for pending migrations", zap.String("migrations_dir", migrationsDir))
	}

	migrator, err := NewMigratorV2WithDSN(dsn, migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Validate migrations
	if err := migrator.ValidateMigrations(); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// Check for dirty migrations
	dirtyMigrations := migrator.GetDirtyMigrations()
	if len(dirtyMigrations) > 0 {
		return fmt.Errorf("database is in dirty state: %d dirty migrations found", len(dirtyMigrations))
	}

	// Get pending migrations
	pendingMigrations := migrator.GetPendingMigrations()
	if len(pendingMigrations) == 0 {
		status := migrator.GetStatus()
		infra.Log.Info("Database is up to date, no migrations needed", zap.Any("status", status))
		return nil
	}

	// Apply pending migrations
	infra.Log.Info("Starting database migrations...",
		zap.Int("pending_count", len(pendingMigrations)),
		zap.String("migrations_path", migrationsDir))

	for _, migration := range pendingMigrations {
		infra.Log.Info("Applying migration",
			zap.String("name", migration.Name),
			zap.Uint64("version", migration.Version),
			zap.String("filename", migration.Filename))

		if err := migrator.ApplyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s (version %d): %w", migration.Name, migration.Version, err)
		}
	}

	status := migrator.GetStatus()
	infra.Log.Info("Migrations completed successfully", zap.Any("status", status))
	return nil
}

// GetStatus returns migration status information
func (m *MigratorV2) GetStatus() map[string]interface{} {
	pending := m.GetPendingMigrations()
	applied := m.GetAppliedMigrations()
	dirty := m.GetDirtyMigrations()

	var latestVersion uint64
	if len(applied) > 0 {
		latestVersion = applied[len(applied)-1].Version
	}

	return map[string]interface{}{
		"total_files":    len(m.migrations) / 2, // Divide by 2 because each migration has up/down files
		"applied":        len(applied),
		"pending":        len(pending),
		"dirty":          len(dirty),
		"latest_version": latestVersion,
		"latest_migration": func() string {
			if len(applied) > 0 {
				return applied[len(applied)-1].Name
			}
			return "none"
		}(),
	}
}

// TableName tells GORM to use the _migrations table
func (MigrationRecord) TableName() string {
	return "_migrations"
}




