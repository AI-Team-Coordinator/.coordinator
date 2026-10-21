package infra

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// MasterBearerOnly rejects coordinator-issued client tokens (config writes require the relay master token).
func MasterBearerOnly(masterToken string, next http.Handler) http.Handler {
	masterToken = strings.TrimSpace(masterToken)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := bearerValue(r.Header.Get("Authorization"))
		if masterToken == "" || got == "" || len(got) != len(masterToken) || subtle.ConstantTimeCompare([]byte(got), []byte(masterToken)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "admin token required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
