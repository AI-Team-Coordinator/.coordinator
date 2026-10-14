package service

import (
	"context"
	"strings"
	"time"

	"coordinator/domain/task"
	"coordinator/model"
	taskdto "coordinator/service/task/dto"
)

func (s *Service) GetWorkItems(ctx context.Context, query taskdto.WorkItemsQuery) (*taskdto.WorkItemsResponse, error) {
	items, err := s.repo.ListWorkTasks(ctx)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.GetMembersLight(ctx)
	if err != nil {
		return nil, err
	}
	slotsByTask := indexSlotsByTask(members)
	statusFilter := strings.TrimSpace(strings.ToLower(query.Status))
	assigneeFilter := strings.TrimSpace(query.Assignee)

	out := make([]taskdto.WorkItemResponse, 0, len(items))
	for _, wt := range items {
		slots := slotsByTask[wt.TaskID]
		status := task.Status(wt.Status, len(slots))
		if statusFilter != "" && statusFilter != "all" && status != statusFilter {
			continue
		}
		if assigneeFilter != "" && !strings.EqualFold(wt.Assignee, assigneeFilter) {
			continue
		}
		title := s.repo.TaskTitle(ctx, wt.TaskID)
		if title == "" {
			title = wt.Summary
		}
		resp := taskdto.WorkItemResponse{
			TaskID:       wt.TaskID,
			Title:        title,
			Summary:      wt.Summary,
			Doc:          wt.Doc,
			Branch:       wt.Branch,
			Services:     wt.Services,
			CreatedBy:    wt.CreatedBy,
			Assignee:     wt.Assignee,
			Status:       status,
			RelatedTasks: wt.RelatedTasks,
			RelatedDocs:  wt.RelatedDocs,
		}
		if !wt.CreatedAt.IsZero() {
			resp.CreatedAt = wt.CreatedAt.UTC().Format(time.RFC3339)
		}
		if !wt.UpdatedAt.IsZero() {
			resp.UpdatedAt = wt.UpdatedAt.UTC().Format(time.RFC3339)
		}
		if !wt.CompletedAt.IsZero() {
			resp.CompletedAt = wt.CompletedAt.UTC().Format(time.RFC3339)
		}
		for _, sl := range slots {
			slot := taskdto.WorkItemSlotResponse{
				Alias:  sl.Alias,
				Status: sl.Status,
				Branch: sl.Branch,
			}
			if !sl.UpdatedAt.IsZero() {
				slot.UpdatedAt = sl.UpdatedAt.UTC().Format(time.RFC3339)
			}
			resp.Slots = append(resp.Slots, slot)
		}
		out = append(out, resp)
	}
	return &taskdto.WorkItemsResponse{Items: out, Total: len(out)}, nil
}

func indexSlotsByTask(members []model.Member) map[string][]model.WorkTaskSlot {
	out := make(map[string][]model.WorkTaskSlot)
	for _, m := range members {
		for _, slot := range m.Slots() {
			tid := strings.TrimSpace(slot.TaskID)
			if tid == "" {
				continue
			}
			out[tid] = append(out[tid], model.WorkTaskSlot{
				Alias:     m.Alias,
				Status:    slot.Status,
				Branch:    slot.Branch,
				UpdatedAt: slot.UpdatedAt,
			})
		}
	}
	return out
}
