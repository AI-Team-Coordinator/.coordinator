package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"coordinator/domain/validate"
	"coordinator/model"
	"coordinator/service/repository"
)

func (s *Service) CurrentAuthorAlias(ctx context.Context) (string, error) {
	alias, err := s.repo.CurrentAuthor(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(alias) == "" {
		return "", &validate.Error{Msg: "set Common/data/.current_author before triage actions"}
	}
	return strings.TrimSpace(alias), nil
}

type PatchExternalEventTriageRequest struct {
	Source       string `json:"source"`
	ExternalID   string `json:"external_id"`
	Status       string `json:"status"`
	LinkedTaskID string `json:"linked_task_id,omitempty"`
	DismissNote  string `json:"dismiss_note,omitempty"`
}

func (s *Service) PatchExternalEventTriage(ctx context.Context, req PatchExternalEventTriageRequest, actorAlias string) (*model.ExternalEventTriage, error) {
	source := strings.TrimSpace(req.Source)
	externalID := strings.TrimSpace(req.ExternalID)
	if source == "" || externalID == "" {
		return nil, &validate.Error{Msg: "source and external_id are required"}
	}
	status := model.ExternalEventTriageStatus(strings.TrimSpace(req.Status))
	if !model.ValidExternalTriageStatus(string(status)) {
		return nil, &validate.Error{Msg: "invalid triage status"}
	}
	actorAlias = strings.TrimSpace(actorAlias)
	if actorAlias == "" {
		return nil, &validate.Error{Msg: "set Common/data/.current_author before triage actions"}
	}

	store, ok := s.repo.(*repository.FileRepository)
	if !ok {
		return nil, fmt.Errorf("triage store unavailable")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	row := model.ExternalEventTriage{
		Source:     source,
		ExternalID: externalID,
		Status:     status,
	}
	if prev, ok := store.GetExternalEventTriage(source, externalID); ok && prev != nil {
		row = *prev
		row.Status = status
	}

	switch status {
	case model.ExternalTriageOpen:
		row.AssignedAlias = ""
		row.AssignedAt = ""
	case model.ExternalTriageInProgress:
		if row.AssignedAlias == "" {
			row.AssignedAlias = actorAlias
			row.AssignedAt = now
		}
	case model.ExternalTriageResolved:
		row.ResolvedBy = actorAlias
		row.ResolvedAt = now
		if tid := strings.TrimSpace(req.LinkedTaskID); tid != "" {
			row.LinkedTaskID = tid
		}
	case model.ExternalTriageDismissed:
		row.ResolvedBy = actorAlias
		row.ResolvedAt = now
		row.DismissNote = strings.TrimSpace(req.DismissNote)
	}

	if err := store.UpsertExternalEventTriage(row); err != nil {
		return nil, err
	}
	out, _ := store.GetExternalEventTriage(source, externalID)
	return out, nil
}
