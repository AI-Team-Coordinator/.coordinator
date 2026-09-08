package coordinator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"coordinator/domain/coordinator/dto"
	"coordinator/domain/coordinator/repository"
	"coordinator/model"
)

var (
	ErrInvalidDays   = errors.New("days must be >= 0")
	ErrInvalidLimit  = errors.New("limit must be >= 0")
	ErrInvalidOffset = errors.New("offset must be >= 0")
	ErrInvalidStatus = errors.New("status must be all, in_progress, or completed")
	ErrInvalidKind   = errors.New("kind must be feature or fix")
	ErrInvalidTaskID = errors.New("invalid task id")
	ErrDocNotFound   = errors.New("document not found")
)

type Service struct {
	repo       repository.CoordinatorRepository
	events     repository.EventStore
	completeMu sync.Mutex
	createMu   sync.Mutex
}

func NewService(repo repository.CoordinatorRepository, events repository.EventStore) *Service {
	return &Service{repo: repo, events: events}
}

func (s *Service) GetTaskDoc(ctx context.Context, taskID string) (*dto.TaskDocResponse, error) {
	doc, err := s.repo.GetTaskDoc(ctx, taskID)
	if err != nil {
		if errors.Is(err, os.ErrInvalid) {
			return nil, ErrInvalidTaskID
		}
		if os.IsNotExist(err) {
			return nil, ErrDocNotFound
		}
		return nil, err
	}
	return &dto.TaskDocResponse{
		TaskID:   doc.TaskID,
		Title:    doc.Title,
		RelPath:  doc.RelPath,
		Markdown: doc.Markdown,
	}, nil
}

func (s *Service) SaveTeam(ctx context.Context, req dto.SaveTeamRequest) (*dto.TeamResponse, error) {
	profile, err := s.repo.GetProject(ctx)
	if err != nil {
		return nil, err
	}
	serviceIDs := make(map[string]struct{}, len(profile.Services))
	for _, svc := range profile.Services {
		if svc.ID != "" {
			serviceIDs[svc.ID] = struct{}{}
		}
	}

	raw := make([]model.TeamPerson, 0, len(req.Members))
	for _, person := range req.Members {
		raw = append(raw, model.TeamPerson{
			Alias: person.Alias,
			Name:  person.Name,
			Role:  person.Role,
			Focus: person.Focus,
		})
	}

	members, err := normalizeAndValidateTeam(raw, serviceIDs)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveTeam(ctx, members); err != nil {
		return nil, err
	}
	out, err := s.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	if warn := s.settingsPushWarning(ctx); warn != "" {
		out.Warning = warn
	} else {
		out.Pushed = true
	}
	return out, nil
}

func (s *Service) CreateService(ctx context.Context, req dto.CreateServiceRequest) (*dto.CreateServiceResponse, error) {
	s.createMu.Lock()
	defer s.createMu.Unlock()

	profile, err := s.repo.GetProject(ctx)
	if err != nil {
		return nil, err
	}

	node, err := normalizeCreateService(profile, req)
	if err != nil {
		return nil, err
	}

	workspace := s.repo.WorkspaceDir()
	if workspace == "" {
		return nil, &ValidationError{Msg: "workspace root is unknown"}
	}
	exists, err := cloneDestinationExists(workspace, node.Repo)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &ValidationError{Msg: "local folder already exists"}
	}

	if err := createOrgRepository(ctx, profile.GitHub.Org, node.GitHubRepo, serviceDescription(node)); err != nil {
		return nil, err
	}

	dest := filepath.Join(workspace, node.Repo)
	cloned := true
	warning := ""
	if err := cloneOrgRepository(ctx, profile.GitHub, node.GitHubRepo, dest); err != nil {
		cloned = false
		warning = err.Error()
	}

	profile.Services = append(profile.Services, node)
	if err := s.repo.SaveProject(ctx, profile); err != nil {
		return nil, &GitHubOpError{Kind: "failed", Msg: "GitHub repository created, but saving project_profile.json failed: " + err.Error()}
	}

	project, err := s.GetProject(ctx)
	if err != nil {
		return nil, err
	}
	pushed := true
	if syncWarn := s.settingsPushWarning(ctx); syncWarn != "" {
		pushed = false
		if warning != "" {
			warning = warning + "; " + syncWarn
		} else {
			warning = syncWarn
		}
	}
	return &dto.CreateServiceResponse{
		Project: project,
		Cloned:  cloned,
		HTMLURL: githubRepoHTMLURL(profile.GitHub, node.GitHubRepo),
		Pushed:  pushed,
		Warning: warning,
	}, nil
}

func (s *Service) settingsPushWarning(ctx context.Context) string {
	if err := s.pushSettings(ctx); err != nil {
		return err.Error()
	}
	return ""
}

func (s *Service) pushSettings(ctx context.Context) error {
	alias, err := s.repo.CurrentAuthor(ctx)
	if err != nil {
		return err
	}
	if alias == "" {
		return &ValidationError{Msg: "set Common/data/.current_author before sync"}
	}
	_, err = s.repo.SyncSettings(ctx, alias)
	return mapGitError(err)
}

func (s *Service) GetSyncStatus(ctx context.Context) (*dto.SyncStatusResponse, error) {
	status, err := s.repo.SettingsSyncStatus(ctx)
	if err != nil {
		return nil, err
	}
	files := status.DirtyFiles
	if files == nil {
		files = []string{}
	}
	return &dto.SyncStatusResponse{
		Alias:      status.Alias,
		Branch:     status.Branch,
		DirtyFiles: files,
	}, nil
}

func (s *Service) SyncCursor(ctx context.Context) (*dto.SyncResponse, error) {
	alias, err := s.repo.CurrentAuthor(ctx)
	if err != nil {
		return nil, err
	}
	if alias == "" {
		return nil, &ValidationError{Msg: "set Common/data/.current_author before sync"}
	}

	result, err := s.repo.SyncSettings(ctx, alias)
	if err != nil {
		return nil, mapGitError(err)
	}
	files := result.Files
	if files == nil {
		files = []string{}
	}
	return &dto.SyncResponse{
		Alias:     result.Alias,
		Committed: result.Committed,
		Pushed:    result.Pushed,
		Files:     files,
		Message:   result.Message,
	}, nil
}

func mapGitError(err error) error {
	var fail *repository.GitFailure
	if errors.As(err, &fail) {
		return &GitSyncError{Msg: fail.Msg, Output: fail.Output, Kind: fail.Kind}
	}
	return err
}

func (s *Service) GetTeam(ctx context.Context) (*dto.TeamResponse, error) {
	team, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	members := make([]dto.TeamPersonResponse, 0, len(team))
	for _, person := range team {
		focus := person.Focus
		if focus == nil {
			focus = []string{}
		}
		members = append(members, dto.TeamPersonResponse{
			Alias: person.Alias,
			Name:  person.Name,
			Role:  person.Role,
			Focus: focus,
		})
	}
	return &dto.TeamResponse{Members: members}, nil
}

func (s *Service) GetProject(ctx context.Context) (*dto.ProjectResponse, error) {
	profile, err := s.repo.GetProject(ctx)
	if err != nil {
		return nil, err
	}

	groups := make([]dto.ServiceGroupResponse, 0, len(profile.Groups))
	for _, g := range profile.Groups {
		groups = append(groups, dto.ServiceGroupResponse{
			ID:    g.ID,
			Label: dto.LocalizedText{EN: g.Label.EN, RU: g.Label.RU},
		})
	}

	services := make([]dto.ServiceNodeResponse, 0, len(profile.Services))
	for _, svc := range profile.Services {
		services = append(services, dto.ServiceNodeResponse{
			ID:         svc.ID,
			Name:       svc.Name,
			Group:      svc.Group,
			Kind:       svc.Kind,
			Repo:       svc.Repo,
			GitHubRepo: svc.GitHubRepo,
			GitHubOrg:  svc.GitHubOrg,
			HTMLURL:    serviceHTMLURL(profile.GitHub, svc),
			Purpose:    dto.LocalizedText{EN: svc.Purpose.EN, RU: svc.Purpose.RU},
		})
	}

	edges := make([]dto.ServiceEdgeResponse, 0, len(profile.Edges))
	for _, e := range profile.Edges {
		edges = append(edges, dto.ServiceEdgeResponse{
			From: e.From,
			To:   e.To,
			Via:  e.Via,
			Kind: e.Kind,
		})
	}

	out := &dto.ProjectResponse{
		Version: profile.Version,
		Project: dto.ProjectMetaResponse{
			ID:      profile.Project.ID,
			Name:    profile.Project.Name,
			Tagline: dto.LocalizedText{EN: profile.Project.Tagline.EN, RU: profile.Project.Tagline.RU},
		},
		Groups:   groups,
		Services: services,
		Edges:    edges,
	}
	if org := strings.TrimSpace(profile.GitHub.Org); org != "" {
		out.GitHub = &dto.GitHubBindingResponse{
			Host:    githubHTMLHost(profile.GitHub.Host),
			Org:     org,
			SSHHost: profile.GitHub.SSHHost,
			HTMLURL: githubOrgHTMLURL(profile.GitHub),
		}
		out.GitHubLive = probeGitHubCLI(ctx, org)
	}
	return out, nil
}

func (s *Service) GetPulse(ctx context.Context) (*dto.PulseResponse, error) {
	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	s.overlayRepoFacts(ctx, members)
	if s.maybeCompleteDeployed(ctx) {
		members, err = s.repo.GetMembers(ctx)
		if err != nil {
			return nil, err
		}
		s.overlayRepoFacts(ctx, members)
	}
	stray, err := s.repo.GetStrayWork(ctx, members)
	if err != nil {
		return nil, err
	}
	return &dto.PulseResponse{
		UpdatedAt: time.Now(),
		Members:   mapMembers(members),
		Conflicts: mapConflicts(detectConflicts(members)),
		Stray:     mapStray(stray),
	}, nil
}

func (s *Service) maybeCompleteDeployed(ctx context.Context) bool {
	alias, err := s.repo.CurrentAuthor(ctx)
	if err != nil || alias == "" {
		return false
	}
	s.completeMu.Lock()
	defer s.completeMu.Unlock()

	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return false
	}
	s.overlayRepoFacts(ctx, members)
	for _, member := range members {
		if member.Alias != alias {
			continue
		}
		if !allReposDeployed(member) {
			return false
		}
		if err := s.repo.CompleteTask(ctx, member.Alias, member.TaskID); err != nil {
			return false
		}
		return true
	}
	return false
}

func (s *Service) GetMembers(ctx context.Context) (*dto.MembersResponse, error) {
	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.MembersResponse{Members: mapMembers(members)}, nil
}

func (s *Service) GetConflicts(ctx context.Context) (*dto.ConflictsResponse, error) {
	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.ConflictsResponse{Conflicts: mapConflicts(detectConflicts(members))}, nil
}

func (s *Service) GetStats(ctx context.Context) (*dto.StatsEnvelope, error) {
	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	events, _, err := s.events.List(ctx, model.EventQuery{})
	if err != nil {
		return nil, err
	}
	stats := computeStats(events, members)
	return &dto.StatsEnvelope{
		Stats: dto.StatsResponse{
			TotalCompleted:        stats.TotalCompleted,
			ActiveNow:             stats.ActiveNow,
			AvgCycleTimeMinutes:   stats.AvgCycleTimeMinutes,
			FeaturesCompleted:     stats.FeaturesCompleted,
			FixesCompleted:        stats.FixesCompleted,
			CompletedToday:        stats.CompletedToday,
			CompletedThisWeek:     stats.CompletedThisWeek,
			CostUSDToday:          stats.CostUSDToday,
			CostUSDWeek:           stats.CostUSDWeek,
			CostUSDTotal:          stats.CostUSDTotal,
			CostUSDAvg:            stats.CostUSDAvg,
			CostTasks:             stats.CostTasks,
			BudgetUSDToday:        stats.BudgetUSDToday,
			BudgetUSDWeek:         stats.BudgetUSDWeek,
			BudgetUSDTotal:        stats.BudgetUSDTotal,
			BudgetUSDAvg:          stats.BudgetUSDAvg,
			BudgetTasks:           stats.BudgetTasks,
			BudgetUSDProductToday: stats.BudgetUSDProductToday,
			BudgetUSDProductWeek:  stats.BudgetUSDProductWeek,
			BudgetUSDInfraToday:   stats.BudgetUSDInfraToday,
			BudgetUSDInfraWeek:    stats.BudgetUSDInfraWeek,
			OnDemandUSDToday:      stats.OnDemandUSDToday,
			OnDemandUSDWeek:       stats.OnDemandUSDWeek,
			BudgetUSDOpen:         stats.BudgetUSDOpen,
			CostUSDOpen:           stats.CostUSDOpen,
			OnDemandUSDOpen:       stats.OnDemandUSDOpen,
		},
	}, nil
}

func (s *Service) GetTasks(ctx context.Context, query dto.TasksQuery) (*dto.TasksResponse, error) {
	if query.Limit < 0 {
		return nil, ErrInvalidLimit
	}
	if query.Offset < 0 {
		return nil, ErrInvalidOffset
	}
	status := strings.TrimSpace(query.Status)
	if status == "" {
		status = "all"
	}
	if status != "all" && status != "in_progress" && status != "completed" {
		return nil, ErrInvalidStatus
	}
	kind := strings.TrimSpace(query.Kind)
	if kind != "" && kind != "feature" && kind != "fix" {
		return nil, ErrInvalidKind
	}

	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	events, _, err := s.events.List(ctx, model.EventQuery{})
	if err != nil {
		return nil, err
	}

	filtered := filterTasks(computeTasks(events, members, time.Now()), model.TaskQuery{
		Status: status,
		Alias:  strings.TrimSpace(query.Alias),
		Kind:   kind,
	})
	page := pageTasks(filtered, query.Limit, query.Offset)
	out := make([]dto.TaskResponse, 0, len(page))
	for _, t := range page {
		title := t.Title
		if title == "" {
			title = s.repo.TaskTitle(ctx, t.TaskID)
		}
		out = append(out, dto.TaskResponse{
			TaskID:          t.TaskID,
			Title:           title,
			Status:          t.Status,
			Kind:            t.Kind,
			Alias:           t.Alias,
			Branch:          t.Branch,
			Services:        t.Services,
			StartedAt:       t.StartedAt,
			CompletedAt:     t.CompletedAt,
			DurationSeconds: t.DurationSeconds,
			CostUSD:         t.CostUSD,
			BudgetUSD:       t.BudgetUSD,
			OnDemandUSD:     t.OnDemandUSD,
			SpendKind:       t.SpendKind,
		})
	}
	return &dto.TasksResponse{Tasks: out, Total: len(filtered), Offset: query.Offset, Limit: query.Limit}, nil
}

func (s *Service) GetEvents(ctx context.Context, query dto.EventsQuery) (*dto.EventsResponse, error) {
	if query.Days < 0 {
		return nil, ErrInvalidDays
	}
	if query.Limit < 0 {
		return nil, ErrInvalidLimit
	}
	if query.Offset < 0 {
		return nil, ErrInvalidOffset
	}

	q := model.EventQuery{
		Alias:   query.Alias,
		Event:   query.Event,
		Service: query.Service,
		Limit:   query.Limit,
		Offset:  query.Offset,
	}
	if query.Days > 0 {
		q.Since = time.Now().AddDate(0, 0, -query.Days).Unix()
	}

	events, total, err := s.events.List(ctx, q)
	if err != nil {
		return nil, err
	}

	out := make([]dto.EventResponse, 0, len(events))
	for _, ev := range events {
		out = append(out, dto.EventResponse{
			Timestamp:       ev.Timestamp,
			Event:           ev.Event,
			TaskID:          ev.TaskID,
			Branch:          ev.Branch,
			Alias:           ev.Alias,
			Repo:            ev.Repo,
			Service:         ev.Service,
			Status:          ev.Status,
			CostUSD:         ev.CostUSD,
			BudgetUSD:       ev.BudgetUSD,
			OnDemandUSD:     ev.OnDemandUSD,
			CursorModelsPct: ev.CursorModelsPct,
			OtherModelsPct:  ev.OtherModelsPct,
			UsagePlan:       ev.UsagePlan,
			PlanPriceUSD:    ev.PlanPriceUSD,
			SpendKind:       ev.SpendKind,
		})
	}

	return &dto.EventsResponse{Events: out, Total: total, Offset: query.Offset, Limit: query.Limit}, nil
}

func mapMembers(members []model.Member) []dto.MemberResponse {
	now := time.Now()
	out := make([]dto.MemberResponse, 0, len(members))
	for _, m := range members {
		focus := m.Focus
		if focus == nil {
			focus = []string{}
		}
		services := m.Services
		if services == nil {
			services = []string{}
		}
		item := dto.MemberResponse{
			Alias:           m.Alias,
			Name:            m.Name,
			Role:            m.Role,
			Focus:           focus,
			Services:        services,
			Status:          m.Status,
			TaskID:          m.TaskID,
			TaskTitle:       m.TaskTitle,
			Branch:          m.Branch,
			UpdatedAt:       m.UpdatedAt,
			Repos:           mapRepos(m.Repos),
			CostUSD:         m.CostUSD,
			BudgetUSD:       m.BudgetUSD,
			OnDemandUSD:     m.OnDemandUSD,
			CursorModelsPct: m.CursorModelsPct,
			OtherModelsPct:  m.OtherModelsPct,
			SpendKind:       m.SpendKind,
		}
		if m.Status == "in_progress" {
			d := int64(now.Sub(m.UpdatedAt).Seconds())
			if d < 0 {
				d = 0
			}
			item.DurationSeconds = d
		}
		out = append(out, item)
	}
	return out
}

func (s *Service) overlayRepoFacts(ctx context.Context, members []model.Member) {
	merges, _, err := s.events.List(ctx, model.EventQuery{Event: "repo_merged", Limit: 500})
	if err == nil {
		applyMergeEvents(members, merges)
	}
	deploys, _, err := s.events.List(ctx, model.EventQuery{Event: "deploy_finished", Limit: 500})
	if err == nil {
		applyDeployEvents(members, deploys)
	}
}

func mapRepos(repos []model.RepoWork) []dto.RepoWorkResponse {
	if len(repos) == 0 {
		return nil
	}
	out := make([]dto.RepoWorkResponse, 0, len(repos))
	for _, repo := range repos {
		out = append(out, dto.RepoWorkResponse{
			ID:       repo.ID,
			Name:     repo.Name,
			Repo:     repo.Repo,
			Kind:     repo.Kind,
			State:    repo.State,
			Ahead:    repo.Ahead,
			Dirty:    repo.Dirty,
			Deployed: repo.Deployed,
		})
	}
	return out
}

func mapConflicts(conflicts []model.Conflict) []dto.ConflictResponse {
	out := make([]dto.ConflictResponse, 0, len(conflicts))
	for _, c := range conflicts {
		out = append(out, dto.ConflictResponse{
			Severity:        c.Severity,
			Title:           c.Title,
			Description:     c.Description,
			AffectedAliases: c.AffectedAliases,
		})
	}
	return out
}

func mapStray(items []model.StrayRepo) []dto.StrayRepoResponse {
	out := make([]dto.StrayRepoResponse, 0, len(items))
	for _, item := range items {
		commits := make([]dto.StrayCommitResponse, 0, len(item.Commits))
		for _, c := range item.Commits {
			commits = append(commits, dto.StrayCommitResponse{Hash: c.Hash, Subject: c.Subject})
		}
		out = append(out, dto.StrayRepoResponse{
			ID:      item.ID,
			Name:    item.Name,
			Repo:    item.Repo,
			Branch:  item.Branch,
			Commits: commits,
		})
	}
	return out
}
