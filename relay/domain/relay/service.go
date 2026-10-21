package relay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"relay/domain/config"
	"relay/domain/connector"
	"relay/domain/gateway"
	"relay/domain/marketplace"
	"relay/model"
)

// Service is the Relay HTTP and use-case layer.
type Service struct {
	Gateway  *gateway.Gateway
	Runtime  *config.Runtime
	Registry *connector.Registry
}

func (s *Service) IngestWebhook(ctx context.Context, source string, d connector.Delivery) error {
	if s == nil || s.Gateway == nil {
		return errors.New("relay service not configured")
	}
	err := s.Gateway.Webhook(ctx, source, d)
	if errors.Is(err, gateway.ErrIgnored) {
		return err
	}
	return err
}

func (s *Service) PublishEvents(ctx context.Context, source string, events []model.ExternalEvent) error {
	if s == nil || s.Gateway == nil {
		return errors.New("relay service not configured")
	}
	return s.Gateway.Publish(ctx, source, events)
}

// Controller is the HTTP edge.
type Controller struct {
	Service *Service
}

func (c *Controller) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","service":"relay"}`))
}

func (c *Controller) Sources(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var list []connector.Descriptor
	if c.Service != nil && c.Service.Registry != nil {
		for _, d := range c.Service.Registry.ListDescriptors() {
			if c.Service.Runtime != nil {
				ent, ok := marketplace.EntryByID(d.ID)
				if ok {
					d.Enabled = ent.Connectable && c.Service.Runtime.Connected(r.Context(), d.ID, ent.SecretKeys)
				} else {
					d.Enabled = c.Service.Runtime.Enabled(d.ID)
				}
			}
			list = append(list, d)
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"sources": list})
}

func readWebhookDelivery(r *http.Request) (connector.Delivery, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return connector.Delivery{}, err
	}
	return connector.Delivery{
		Event: r.Header.Get("X-GitHub-Event"),
		Body:  body,
		Sig:   r.Header.Get("X-Hub-Signature-256"),
	}, nil
}

func (c *Controller) ConnectorWebhook(w http.ResponseWriter, r *http.Request) {
	source := strings.TrimSpace(r.PathValue("source"))
	d, err := readWebhookDelivery(r)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	err = c.Service.IngestWebhook(r.Context(), source, d)
	if errors.Is(err, gateway.ErrIgnored) || errors.Is(err, connector.ErrIgnored) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if errors.Is(err, connector.ErrBadSignature) {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, gateway.ErrUnknownConnector) {
		http.Error(w, "unknown connector", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "ingest failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) ConnectorEvents(w http.ResponseWriter, r *http.Request) {
	source := strings.TrimSpace(r.PathValue("source"))
	var events []model.ExternalEvent
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&events); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(events) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := c.Service.PublishEvents(r.Context(), source, events); err != nil {
		if errors.Is(err, gateway.ErrUnknownConnector) {
			http.Error(w, "unknown connector", http.StatusNotFound)
			return
		}
		http.Error(w, "publish failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GitHubWebhook keeps the legacy path for existing GitHub App hook URLs.
func (c *Controller) GitHubWebhook(w http.ResponseWriter, r *http.Request) {
	d, err := readWebhookDelivery(r)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	err = c.Service.IngestWebhook(r.Context(), "github", d)
	if errors.Is(err, gateway.ErrIgnored) || errors.Is(err, connector.ErrIgnored) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if errors.Is(err, connector.ErrBadSignature) {
		http.Error(w, "bad signature", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "ingest failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
