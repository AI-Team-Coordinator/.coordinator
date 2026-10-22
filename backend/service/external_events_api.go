package service

import (
	"context"
	"fmt"
	"strings"

	pulsedto "coordinator/service/pulse/dto"
	"coordinator/model"
	"coordinator/service/repository"
)

func (s *Service) GetExternalEvents(ctx context.Context, source string, limit, offset int) (*pulsedto.ExternalEventsListResponse, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset must be >= 0")
	}
	rows, total, sources, err := s.repo.ListExternalEvents(source, limit, offset)
	if err != nil {
		return nil, err
	}
	triageMap := s.externalEventTriageMap()
	events := make([]pulsedto.ExternalEventResponse, 0, len(rows))
	for _, row := range rows {
		events = append(events, mapExternalEventRow(row, triageMap))
	}
	return &pulsedto.ExternalEventsListResponse{
		Events:  events,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Sources: sources,
	}, nil
}

func (s *Service) externalEventTriageMap() map[string]model.ExternalEventTriage {
	store, ok := s.repo.(*repository.FileRepository)
	if !ok {
		return map[string]model.ExternalEventTriage{}
	}
	return store.ListExternalEventTriage()
}

func mapExternalEventRow(row repository.ExternalEventRow, triage map[string]model.ExternalEventTriage) pulsedto.ExternalEventResponse {
	requires := row.RequiresReactionEffective()
	key := strings.TrimSpace(row.Source) + "\x00" + strings.TrimSpace(row.ExternalID)
	var tri *pulsedto.ExternalEventTriageResponse
	if t, ok := triage[key]; ok {
		tri = mapTriageDTO(t)
	} else if requires {
		tri = &pulsedto.ExternalEventTriageResponse{
			Status: string(model.ExternalTriageOpen),
		}
	}
	return pulsedto.ExternalEventResponse{
		ID:               row.ID,
		Source:           row.Source,
		ExternalID:       row.ExternalID,
		Class:            row.Class,
		Kind:             row.Kind,
		Title:            row.Title,
		Summary:          row.Summary,
		URL:              row.URL,
		ReceivedAt:       row.ReceivedAt,
		OccurredAt:       row.OccurredAt,
		Tags:             row.Tags,
		Data:             row.Data,
		RequiresReaction: requires,
		Triage:           tri,
	}
}

func mapTriageDTO(t model.ExternalEventTriage) *pulsedto.ExternalEventTriageResponse {
	return &pulsedto.ExternalEventTriageResponse{
		Status:        string(t.Status),
		AssignedAlias: t.AssignedAlias,
		AssignedAt:    t.AssignedAt,
		LinkedTaskID:  t.LinkedTaskID,
		ResolvedBy:    t.ResolvedBy,
		ResolvedAt:    t.ResolvedAt,
		DismissNote:   t.DismissNote,
	}
}
