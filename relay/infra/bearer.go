package infra

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const relayBearerFile = ".relay_bearer_token"

// LoadOrCreateBearerToken reads RELAY_BEARER_TOKEN, else data/.relay_bearer_token, else creates one.
func LoadOrCreateBearerToken(dataDir string) (string, error) {
	if v := strings.TrimSpace(os.Getenv("RELAY_BEARER_TOKEN")); v != "" {
		return v, nil
	}
	if dataDir == "" {
		return "", fmt.Errorf("data dir is empty")
	}
	path := filepath.Join(dataDir, relayBearerFile)
	if data, err := os.ReadFile(path); err == nil {
		t := strings.TrimSpace(string(data))
		if t != "" {
			return t, nil
		}
	}
	token, err := newBearerToken()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return "", err
	}
	LogInfo("relay API bearer token written to %s (set RELAY_BEARER_TOKEN to override)", path)
	return token, nil
}

func newBearerToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// RequireBearer protects relay HTTP handlers (constant-time compare).
func RequireBearer(want string, next http.Handler) http.Handler {
	want = strings.TrimSpace(want)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want == "" || !bearerMatches(r.Header.Get("Authorization"), want) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerMatches(header, want string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	got := strings.TrimSpace(header[len(prefix):])
	if got == "" {
		return false
	}
	if len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}
