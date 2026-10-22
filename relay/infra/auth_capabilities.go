package infra

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

// AuthCapabilities reports whether the bearer token may change connector config (master only).
func AuthCapabilities(masterToken string) http.HandlerFunc {
	masterToken = strings.TrimSpace(masterToken)
	return func(w http.ResponseWriter, r *http.Request) {
		got := bearerValue(r.Header.Get("Authorization"))
		configure := masterToken != "" && got != "" && len(got) == len(masterToken) &&
			subtle.ConstantTimeCompare([]byte(got), []byte(masterToken)) == 1
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{
			"configure": configure,
		})
	}
}
