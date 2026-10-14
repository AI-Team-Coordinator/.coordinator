package relay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"relay/domain/adapter"
	"relay/domain/relay/github"
	"relay/model"
)

// Appender writes one normalized event to the bus.
type Appender interface {
	Append(ctx context.Context, ev model.ExternalEvent) error
}

// Service is the Relay use-case layer.
type Service struct {
	GitHub adapter.Adapter
	Bus    Appender
}

func (s *Service) IngestGitHub(ctx context.Context, eventName string, body []byte, signature string) error {
	events, err := s.GitHub.Normalize(ctx, adapter.Delivery{Event: eventName, Body: body, Sig: signature})
	if err != nil {
		return err
	}
	for _, ev := range events {
		if err := s.Bus.Append(ctx, ev); err != nil {
			return err
		}
	}
	return nil
}

// Controller is the HTTP edge. Same shape as the Coordinator controller: decode, call service, status.
type Controller struct {
	Service *Service
}

func (c *Controller) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","service":"relay"}`))
}

func (c *Controller) Sources(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enabled := false
	if c.Service != nil {
		if gh, ok := c.Service.GitHub.(github.Adapter); ok {
			enabled = strings.TrimSpace(gh.WebhookSecret) != ""
		}
	}
	out := map[string]any{
		"sources": []map[string]any{
			{
				"id":      "github",
				"name":    "GitHub",
				"detail":  "Pull requests, reviews, and Actions (read-only).",
				"enabled": enabled,
			},
		},
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (c *Controller) GitHubWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	err = c.Service.IngestGitHub(r.Context(), r.Header.Get("X-GitHub-Event"), body, r.Header.Get("X-Hub-Signature-256"))
	if errors.Is(err, github.ErrIgnored) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if errors.Is(err, github.ErrBadSignature) {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "ingest failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
