package infra

import "testing"

func TestNormalizePostgresSSLMode(t *testing.T) {
	if normalizePostgresSSLMode("enabled") != "require" {
		t.Fatal("enabled")
	}
	if normalizePostgresSSLMode("disable") != "disable" {
		t.Fatal("disable")
	}
}
