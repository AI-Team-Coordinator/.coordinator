package infra

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const relayIssuedTokensFile = ".relay_issued_tokens.json"

type issuedTokenFile struct {
	Version int `json:"version"`
	Tokens  []struct {
		TokenSHA256 string `json:"token_sha256"`
	} `json:"tokens"`
}

// BearerAuth accepts the relay master token or any coordinator-issued client token.
func BearerAuth(masterToken, dataDir string, next http.Handler) http.Handler {
	v := &issuedTokenCache{dataDir: dataDir}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !v.valid(r.Header.Get("Authorization"), masterToken) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type issuedTokenCache struct {
	dataDir string
	mu      sync.Mutex
	hashes  []string
	modTime time.Time
}

func (c *issuedTokenCache) valid(header, master string) bool {
	got := bearerValue(header)
	if got == "" {
		return false
	}
	master = strings.TrimSpace(master)
	if master != "" && len(got) == len(master) && subtle.ConstantTimeCompare([]byte(got), []byte(master)) == 1 {
		return true
	}
	sum := sha256.Sum256([]byte(got))
	hexDigest := hex.EncodeToString(sum[:])
	for _, h := range c.loadHashes() {
		if len(h) == len(hexDigest) && subtle.ConstantTimeCompare([]byte(hexDigest), []byte(h)) == 1 {
			return true
		}
	}
	return false
}

func bearerValue(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}

func (c *issuedTokenCache) loadHashes() []string {
	path := filepath.Join(c.dataDir, relayIssuedTokensFile)
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !info.ModTime().After(c.modTime) && c.hashes != nil {
		return c.hashes
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return c.hashes
	}
	var file issuedTokenFile
	if err := json.Unmarshal(data, &file); err != nil {
		return c.hashes
	}
	out := make([]string, 0, len(file.Tokens))
	for _, t := range file.Tokens {
		h := strings.TrimSpace(t.TokenSHA256)
		if h != "" {
			out = append(out, h)
		}
	}
	c.hashes = out
	c.modTime = info.ModTime()
	return c.hashes
}
