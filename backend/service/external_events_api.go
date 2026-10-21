package service

import (
	"context"
	"fmt"

	pulsedto "coordinator/service/pulse/dto"
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
	events := make([]pulsedto.ExternalEventResponse, 0, len(rows))
	for _, row := range rows {
		events = append(events, mapExternalEventRow(row))
	}
	return &pulsedto.ExternalEventsListResponse{
		Events:  events,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Sources: sources,
	}, nil
}

func mapExternalEventRow(row repository.ExternalEventRow) pulsedto.ExternalEventResponse {
	return pulsedto.ExternalEventResponse{
		ID:         row.ID,
		Source:     row.Source,
		ExternalID: row.ExternalID,
		Class:      row.Class,
		Kind:       row.Kind,
		Title:      row.Title,
		Summary:    row.Summary,
		URL:        row.URL,
		ReceivedAt: row.ReceivedAt,
		OccurredAt: row.OccurredAt,
		Tags:       row.Tags,
		Data:       row.Data,
	}
}
