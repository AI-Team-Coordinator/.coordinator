package coordinator

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeRelayURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"http://127.0.0.1:4322", "http://127.0.0.1:4322", true},
		{"127.0.0.1:4322", "http://127.0.0.1:4322", true},
		{"localhost:4322", "http://localhost:4322", true},
		{"https://relay.example.com", "https://relay.example.com", true},
		{"http://evil.com", "", false},
	}
	for _, tc := range cases {
		got, err := normalizeRelayURL(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
		if tc.ok && got != tc.want {
			t.Fatalf("%q => %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestPingRelayHealthBearer(t *testing.T) {
	const want = "relay-secret-token-value!!"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok","service":"relay"}`))
	}))
	defer srv.Close()

	if err := pingRelayHealth(srv.URL, "wrong"); err == nil {
		t.Fatal("expected error for wrong token")
	}
	if err := pingRelayHealth(srv.URL, want); err != nil {
		t.Fatalf("valid token: %v", err)
	}
}
