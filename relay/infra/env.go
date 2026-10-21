package infra

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}
	for {
		for _, name := range []string{".env", filepath.Join("..", "Common", ".env")} {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return godotenv.Load(path)
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil
		}
		dir = parent
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func GetEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("environment variable %s is not set", key)
	}
	return v, nil
}
