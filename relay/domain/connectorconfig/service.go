package connectorconfig

import (
	"context"
	"fmt"
	"strings"
)

// Service is the domain API for connector enablement and non-secret configuration.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) IsEnabled(id string) bool {
	if s == nil || s.repo == nil {
		return false
	}
	id = strings.TrimSpace(id)
	row, ok, err := s.repo.Get(context.Background(), id)
	if err != nil || !ok {
		return false
	}
	return row.Enabled
}

func (s *Service) NonSecret(id string) map[string]any {
	if s == nil || s.repo == nil {
		return map[string]any{}
	}
	id = strings.TrimSpace(id)
	row, ok, err := s.repo.Get(context.Background(), id)
	if err != nil || !ok || row.NonSecret == nil {
		return map[string]any{}
	}
	return row.NonSecret
}

func (s *Service) SetEnabled(id string, on bool) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("connector config repository is not configured")
	}
	return s.repo.SetEnabled(context.Background(), strings.TrimSpace(id), on)
}

func (s *Service) MergeNonSecret(id string, patch map[string]any) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("connector config repository is not configured")
	}
	if patch == nil {
		return nil
	}
	return s.repo.MergeNonSecret(context.Background(), strings.TrimSpace(id), patch)
}

// ReconcileEnabledFromSecrets enables a connector when secrets exist but no meta row (legacy file loss).
func (s *Service) ReconcileEnabledFromSecrets(ctx context.Context, hasSecrets bool, connectorID string) error {
	if s == nil || s.repo == nil || !hasSecrets {
		return nil
	}
	connectorID = strings.TrimSpace(connectorID)
	if connectorID == "" || s.IsEnabled(connectorID) {
		return nil
	}
	_, exists, err := s.repo.Get(ctx, connectorID)
	if err != nil || exists {
		return err
	}
	return s.repo.SetEnabled(ctx, connectorID, true)
}
