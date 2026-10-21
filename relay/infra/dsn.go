package infra

import (
	"fmt"
	"strings"
)

// GetDsn returns PostgreSQL DSN for Relay's own database.
// 1) RELAY_DATABASE_URL (postgres URL or libpq key=value string)
// 2) RELAY_POSTGRES_HOST, RELAY_POSTGRES_USER, RELAY_POSTGRES_PASSWORD, RELAY_POSTGRES_DB, RELAY_POSTGRES_PORT
func GetDsn() (string, error) {
	if u := GetEnv("RELAY_DATABASE_URL"); u != "" {
		return u, nil
	}
	host, err := requireEnv("RELAY_POSTGRES_HOST")
	if err != nil {
		return "", fmt.Errorf("relay database: %w (or set RELAY_DATABASE_URL)", err)
	}
	user, err := requireEnv("RELAY_POSTGRES_USER")
	if err != nil {
		return "", err
	}
	password, err := requireEnv("RELAY_POSTGRES_PASSWORD")
	if err != nil {
		return "", err
	}
	dbname, err := requireEnv("RELAY_POSTGRES_DB")
	if err != nil {
		return "", err
	}
	port, err := requireEnv("RELAY_POSTGRES_PORT")
	if err != nil {
		return "", err
	}
	sslmode := normalizePostgresSSLMode(GetEnvOrDefault("RELAY_POSTGRES_SSLMODE", "disable"))
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode), nil
}

// normalizePostgresSSLMode maps common Coolify/UI values to libpq sslmode.
func normalizePostgresSSLMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "enabled", "enable", "on", "true", "1":
		return "require"
	case "disabled", "disable", "off", "false", "0":
		return "disable"
	default:
		return mode
	}
}
