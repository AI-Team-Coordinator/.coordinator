package service

import (
	"context"
	"coordinator/domain/validate"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"coordinator/domain/identity"
	"coordinator/domain/org"
	"coordinator/domain/spend"
	"coordinator/domain/team"
	eventdto "coordinator/domain/timeline/dto"
	"coordinator/model"
	orgdto "coordinator/service/org/dto"
	projectdto "coordinator/service/project/dto"
	pulsedto "coordinator/service/pulse/dto"
	"coordinator/service/repository"
	statsdto "coordinator/service/stats/dto"
	syncdto "coordinator/service/sync/dto"
	taskdto "coordinator/service/task/dto"
	teamdto "coordinator/service/team/dto"
)

var (
	ErrInvalidDays   = errors.New("days must be >= 0")
	ErrInvalidLimit  = errors.New("limit must be >= 0")
	ErrInvalidOffset = errors.New("offset must be >= 0")
	ErrInvalidStatus = errors.New("status must be all, in_progress, parked, or completed")
	ErrInvalidKind   = errors.New("kind must be feature or fix")
	ErrInvalidTaskID = errors.New("invalid task id")
	ErrDocNotFound   = errors.New("document not found")
)

type Service struct {
	repo        repository.CoordinatorRepository
	events      repository.EventStore
	completeMu  sync.Mutex
	createMu    sync.Mutex
	noticeMu    sync.Mutex
	noticeAlias string
	noticeSeen  map[string]struct{}
}

func NewService(repo repository.CoordinatorRepository, events repository.EventStore) *Service {
	return &Service{repo: repo, events: events}
}

func (s *Service) GetTaskDoc(ctx context.Context, taskID string) (*taskdto.TaskDocResponse, error) {
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
	return &taskdto.TaskDocResponse{
		TaskID:   doc.TaskID,
		Title:    doc.Title,
		RelPath:  doc.RelPath,
		Markdown: doc.Markdown,
	}, nil
}

func (s *Service) SaveTeam(ctx context.Context, req teamdto.SaveTeamRequest) (*teamdto.TeamResponse, error) {
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

	previous, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	previousAccess := team.AccessByAlias(previous)

	topology, err := s.repo.GetOrgTopology(ctx)
	if err != nil {
		return nil, err
	}
	crewIDs := org.CrewIDs(topology)

	raw := make([]model.TeamPerson, 0, len(req.Members))
	for _, person := range req.Members {
		raw = append(raw, model.TeamPerson{
			Alias:       person.Alias,
			Name:        person.Name,
			Role:        person.Role,
			Access:      person.Access,
			Focus:       person.Focus,
			Crews:       person.Crews,
			PrimaryCrew: person.PrimaryCrew,
			PulseScope:  person.PulseScope,
		})
	}

	members, err := team.Normalize(raw, serviceIDs, crewIDs)
	if err != nil {
		return nil, err
	}
	members = team.PreserveAccess(members, previousAccess)
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

func (s *Service) CreateService(ctx context.Context, req projectdto.CreateServiceRequest) (*projectdto.CreateServiceResponse, error) {
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
		return nil, &validate.Error{Msg: "workspace root is unknown"}
	}
	exists, err := cloneDestinationExists(workspace, node.Repo)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, &validate.Error{Msg: "local folder already exists"}
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
	return &projectdto.CreateServiceResponse{
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
		return &validate.Error{Msg: "set Common/data/.current_author before sync"}
	}
	_, err = s.repo.SyncSettings(ctx, alias)
	return mapGitError(err)
}

func (s *Service) GetSyncStatus(ctx context.Context) (*syncdto.SyncStatusResponse, error) {
	status, err := s.repo.SettingsSyncStatus(ctx)
	if err != nil {
		return nil, err
	}
	files := status.DirtyFiles
	if files == nil {
		files = []string{}
	}
	return &syncdto.SyncStatusResponse{
		Alias:      status.Alias,
		Branch:     status.Branch,
		DirtyFiles: files,
	}, nil
}

func (s *Service) SyncCursor(ctx context.Context) (*syncdto.SyncResponse, error) {
	alias, err := s.repo.CurrentAuthor(ctx)
	if err != nil {
		return nil, err
	}
	if alias == "" {
		return nil, &validate.Error{Msg: "set Common/data/.current_author before sync"}
	}

	result, err := s.repo.SyncSettings(ctx, alias)
	if err != nil {
		return nil, mapGitError(err)
	}
	files := result.Files
	if files == nil {
		files = []string{}
	}
	return &syncdto.SyncResponse{
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

func (s *Service) GetTeam(ctx context.Context) (*teamdto.TeamResponse, error) {
	team, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	members := make([]teamdto.TeamPersonResponse, 0, len(team))
	for _, person := range team {
		focus := person.Focus
		if focus == nil {
			focus = []string{}
		}
		crews := person.Crews
		if crews == nil {
			crews = []string{}
		}
		members = append(members, teamdto.TeamPersonResponse{
			Alias:       person.Alias,
			Name:        person.Name,
			Role:        person.Role,
			Access:      person.Access,
			Focus:       focus,
			Crews:       crews,
			PrimaryCrew: person.PrimaryCrew,
			PulseScope:  person.PulseScope,
		})
	}
	return &teamdto.TeamResponse{Members: members}, nil
}

func (s *Service) GetOrgTopology(ctx context.Context) (*orgdto.OrgTopologyResponse, error) {
	topology, err := s.repo.GetOrgTopology(ctx)
	if err != nil {
		return nil, err
	}
	return mapOrgTopologyResponse(topology), nil
}

func (s *Service) SaveOrgTopology(ctx context.Context, req orgdto.SaveOrgTopologyRequest) (*orgdto.OrgTopologyResponse, error) {
	raw := model.OrgTopology{
		Version:        req.Version,
		Crews:          make([]model.Crew, 0, len(req.Crews)),
		CrewConnectors: make([]model.CrewConnector, 0, len(req.CrewConnectors)),
	}
	for _, c := range req.Crews {
		raw.Crews = append(raw.Crews, model.Crew{
			ID:            c.ID,
			Label:         model.LocalizedText{EN: c.Label.EN, RU: c.Label.RU},
			Kind:          c.Kind,
			InterestBrief: c.InterestBrief,
			Notes:         c.Notes,
		})
	}
	for _, link := range req.CrewConnectors {
		interests := make([]model.CrewConnectorItem, 0, len(link.Interests))
		for _, item := range link.Interests {
			interests = append(interests, model.CrewConnectorItem{
				Title:  item.Title,
				Detail: item.Detail,
				Tags:   item.Tags,
			})
		}
		raw.CrewConnectors = append(raw.CrewConnectors, model.CrewConnector{
			ID:         link.ID,
			FromCrewID: link.FromCrewID,
			ToCrewID:   link.ToCrewID,
			Summary:    link.Summary,
			Notes:      link.Notes,
			Interests:  interests,
		})
	}

	topology, err := org.Normalize(raw)
	if err != nil {
		return nil, err
	}
	team, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	if err := org.ValidateMembers(topology, team); err != nil {
		return nil, err
	}
	if err := s.repo.SaveOrgTopology(ctx, topology); err != nil {
		return nil, err
	}
	out := mapOrgTopologyResponse(topology)
	if warn := s.settingsPushWarning(ctx); warn != "" {
		out.Warning = warn
	} else {
		out.Pushed = true
	}
	return out, nil
}

func mapOrgTopologyResponse(topology model.OrgTopology) *orgdto.OrgTopologyResponse {
	crews := make([]orgdto.CrewResponse, 0, len(topology.Crews))
	for _, c := range topology.Crews {
		crews = append(crews, orgdto.CrewResponse{
			ID:            c.ID,
			Label:         model.LocalizedText{EN: c.Label.EN, RU: c.Label.RU},
			Kind:          c.Kind,
			InterestBrief: c.InterestBrief,
			Notes:         c.Notes,
		})
	}
	connectors := make([]orgdto.CrewConnectorResponse, 0, len(topology.CrewConnectors))
	for _, link := range topology.CrewConnectors {
		interests := make([]orgdto.CrewConnectorItemResponse, 0, len(link.Interests))
		for _, item := range link.Interests {
			tags := item.Tags
			if tags == nil {
				tags = []string{}
			}
			interests = append(interests, orgdto.CrewConnectorItemResponse{
				Title:  item.Title,
				Detail: item.Detail,
				Tags:   tags,
			})
		}
		connectors = append(connectors, orgdto.CrewConnectorResponse{
			ID:         link.ID,
			FromCrewID: link.FromCrewID,
			ToCrewID:   link.ToCrewID,
			Summary:    link.Summary,
			Notes:      link.Notes,
			Interests:  interests,
		})
	}
	return &orgdto.OrgTopologyResponse{
		Version:        topology.Version,
		Crews:          crews,
		CrewConnectors: connectors,
	}
}

func (s *Service) GetProject(ctx context.Context) (*projectdto.ProjectResponse, error) {
	profile, err := s.repo.GetProject(ctx)
	if err != nil {
		return nil, err
	}

	groups := make([]projectdto.ServiceGroupResponse, 0, len(profile.Groups))
	for _, g := range profile.Groups {
		groups = append(groups, projectdto.ServiceGroupResponse{
			ID:    g.ID,
			Label: model.LocalizedText{EN: g.Label.EN, RU: g.Label.RU},
		})
	}

	services := make([]projectdto.ServiceNodeResponse, 0, len(profile.Services))
	for _, svc := range profile.Services {
		services = append(services, projectdto.ServiceNodeResponse{
			ID:         svc.ID,
			Name:       svc.Name,
			Group:      svc.Group,
			Kind:       svc.Kind,
			Repo:       svc.Repo,
			GitHubRepo: svc.GitHubRepo,
			GitHubOrg:  svc.GitHubOrg,
			HTMLURL:    serviceHTMLURL(profile.GitHub, svc),
			Purpose:    model.LocalizedText{EN: svc.Purpose.EN, RU: svc.Purpose.RU},
		})
	}

	edges := make([]projectdto.ServiceEdgeResponse, 0, len(profile.Edges))
	for _, e := range profile.Edges {
		edges = append(edges, projectdto.ServiceEdgeResponse{
			From: e.From,
			To:   e.To,
			Via:  e.Via,
			Kind: e.Kind,
		})
	}

	out := &projectdto.ProjectResponse{
		Version: profile.Version,
		Project: projectdto.ProjectMetaResponse{
			ID:      profile.Project.ID,
			Name:    profile.Project.Name,
			Tagline: model.LocalizedText{EN: profile.Project.Tagline.EN, RU: profile.Project.Tagline.RU},
		},
		Groups:   groups,
		Services: services,
		Edges:    edges,
	}
	if org := strings.TrimSpace(profile.GitHub.Org); org != "" {
		out.GitHub = &projectdto.GitHubBindingResponse{
			Host:    githubHTMLHost(profile.GitHub.Host),
			Org:     org,
			SSHHost: profile.GitHub.SSHHost,
			HTMLURL: githubOrgHTMLURL(profile.GitHub),
		}
		out.GitHubLive = probeGitHubCLI(ctx, org)
	}
	return out, nil
}

func (s *Service) GetPulse(ctx context.Context) (*pulsedto.PulseResponse, error) {
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
	conflicts := detectConflicts(members)
	s.persistCoordinatorNotices(ctx, members, conflicts)
	return &pulsedto.PulseResponse{
		UpdatedAt: time.Now(),
		Members:   mapMembers(members),
		Conflicts: mapConflicts(conflicts),
		Stray:     mapStray(stray),
	}, nil
}

func (s *Service) persistCoordinatorNotices(ctx context.Context, members []model.Member, conflicts []model.Conflict) {
	if len(conflicts) == 0 {
		return
	}
	alias, err := s.repo.CurrentAuthor(ctx)
	if err != nil || alias == "" {
		return
	}
	s.noticeMu.Lock()
	defer s.noticeMu.Unlock()
	if s.noticeSeen == nil || s.noticeAlias != alias {
		s.noticeSeen = s.loadNoticeFingerprints(ctx, alias)
		s.noticeAlias = alias
	}
	pending := pendingNotices(alias, observerTaskID(members, alias), conflicts, s.noticeSeen)
	if len(pending) == 0 {
		return
	}
	wrote := false
	for _, ev := range pending {
		if err := s.repo.AppendEvent(ctx, ev); err != nil {
			continue
		}
		wrote = true
		if ev.Findings != "" {
			s.noticeSeen[ev.Findings] = struct{}{}
		}
		if ev.Summary != "" {
			s.noticeSeen[ev.Event+"|"+ev.Summary] = struct{}{}
		}
	}
	if wrote {
		_ = s.events.Sync(ctx)
	}
}

func (s *Service) loadNoticeFingerprints(ctx context.Context, alias string) map[string]struct{} {
	known := make(map[string]struct{})
	for _, kind := range []string{eventCoordinatorWarning, eventCoordinatorStop} {
		events, _, err := s.events.List(ctx, model.EventQuery{Alias: alias, Event: kind})
		if err != nil {
			continue
		}
		for k := range noticeFingerprints(events) {
			known[k] = struct{}{}
		}
	}
	return known
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
		for _, slot := range member.ActiveSlots() {
			if !slotReposDeployed(slot) {
				continue
			}
			if err := s.repo.CompleteTask(ctx, member.Alias, slot.TaskID); err != nil {
				return false
			}
			return true
		}
	}
	return false
}

func (s *Service) GetMembers(ctx context.Context) (*pulsedto.MembersResponse, error) {
	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	return &pulsedto.MembersResponse{Members: mapMembers(members)}, nil
}

func (s *Service) GetConflicts(ctx context.Context) (*pulsedto.ConflictsResponse, error) {
	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	return &pulsedto.ConflictsResponse{Conflicts: mapConflicts(detectConflicts(members))}, nil
}

func (s *Service) GetStats(ctx context.Context, query statsdto.StatsQuery) (*statsdto.StatsEnvelope, error) {
	members, err := s.repo.GetMembers(ctx)
	if err != nil {
		return nil, err
	}
	events, _, err := s.events.List(ctx, model.EventQuery{})
	if err != nil {
		return nil, err
	}
	samples := s.repo.LoadUsageSamples()
	snap := s.repo.CursorUsageSnapshot()
	cycleISO, cycleUnix := resolveBillingCycleStart(snap, members, samples)
	periodStart := cycleUnix
	periodEnd := int64(0)
	if query.From > 0 {
		periodStart = query.From
	}
	if query.To > 0 {
		periodEnd = query.To
	}
	stats := computeStatsSince(events, members, periodStart, periodEnd)
	stats.BillingCycleStart = cycleISO
	if snap != nil && snap.PlanPriceUSD != nil && *snap.PlanPriceUSD > 0 {
		p := *snap.PlanPriceUSD
		stats.PlanPriceUSD = &p
	}
	stats.Hours = s.usageHours(members, events)
	stats.Days = spend.FilterDaysInPeriod(spend.BuildDailySpend(stats.Hours, time.Now(), spend.DailySpendDays), periodStart, periodEnd)
	periods := spend.BillingPeriodsFromSamples(samples, time.Now())
	return &statsdto.StatsEnvelope{
		PeriodFrom:     periodStart,
		PeriodTo:       periodEnd,
		BillingPeriods: mapBillingPeriods(periods),
		Stats: statsdto.StatsResponse{
			TotalCompleted:               stats.TotalCompleted,
			ActiveNow:                    stats.ActiveNow,
			AvgCycleTimeMinutes:          stats.AvgCycleTimeMinutes,
			FeaturesCompleted:            stats.FeaturesCompleted,
			FixesCompleted:               stats.FixesCompleted,
			CompletedToday:               stats.CompletedToday,
			CompletedThisWeek:            stats.CompletedThisWeek,
			CostUSDToday:                 stats.CostUSDToday,
			CostUSDWeek:                  stats.CostUSDWeek,
			CostUSDCycle:                 stats.CostUSDCycle,
			CostUSDTotal:                 stats.CostUSDTotal,
			CostUSDAvg:                   stats.CostUSDAvg,
			CostTasks:                    stats.CostTasks,
			BudgetUSDToday:               stats.BudgetUSDToday,
			BudgetUSDWeek:                stats.BudgetUSDWeek,
			BudgetUSDCycle:               stats.BudgetUSDCycle,
			BudgetUSDTotal:               stats.BudgetUSDTotal,
			BudgetUSDAvg:                 stats.BudgetUSDAvg,
			BudgetTasks:                  stats.BudgetTasks,
			BudgetUSDProductToday:        stats.BudgetUSDProductToday,
			BudgetUSDProductWeek:         stats.BudgetUSDProductWeek,
			BudgetUSDProductCycle:        stats.BudgetUSDProductCycle,
			BudgetUSDInfraToday:          stats.BudgetUSDInfraToday,
			BudgetUSDInfraWeek:           stats.BudgetUSDInfraWeek,
			BudgetUSDInfraCycle:          stats.BudgetUSDInfraCycle,
			OnDemandUSDToday:             stats.OnDemandUSDToday,
			OnDemandUSDWeek:              stats.OnDemandUSDWeek,
			OnDemandUSDCycle:             stats.OnDemandUSDCycle,
			BudgetUSDOpen:                stats.BudgetUSDOpen,
			CostUSDOpen:                  stats.CostUSDOpen,
			OnDemandUSDOpen:              stats.OnDemandUSDOpen,
			BudgetUSDResearchToday:       stats.BudgetUSDResearchToday,
			BudgetUSDResearchWeek:        stats.BudgetUSDResearchWeek,
			BudgetUSDResearchCycle:       stats.BudgetUSDResearchCycle,
			BudgetUSDResearchOpen:        stats.BudgetUSDResearchOpen,
			CostUSDResearchToday:         stats.CostUSDResearchToday,
			CostUSDResearchWeek:          stats.CostUSDResearchWeek,
			CostUSDResearchCycle:         stats.CostUSDResearchCycle,
			CostUSDResearchOpen:          stats.CostUSDResearchOpen,
			ResearchCompleted:            stats.ResearchCompleted,
			ResearchCompletedToday:       stats.ResearchCompletedToday,
			ResearchCompletedWeek:        stats.ResearchCompletedWeek,
			BillingCycleStart:            stats.BillingCycleStart,
			PlanPriceUSD:                 stats.PlanPriceUSD,
			CursorModelsPctToday:         stats.CursorModelsPctToday,
			OtherModelsPctToday:          stats.OtherModelsPctToday,
			CursorModelsPctCycle:         stats.CursorModelsPctCycle,
			OtherModelsPctCycle:          stats.OtherModelsPctCycle,
			CursorModelsPctOpen:          stats.CursorModelsPctOpen,
			OtherModelsPctOpen:           stats.OtherModelsPctOpen,
			CursorModelsPctProductCycle:  stats.CursorModelsPctProductCycle,
			OtherModelsPctProductCycle:   stats.OtherModelsPctProductCycle,
			CursorModelsPctInfraCycle:    stats.CursorModelsPctInfraCycle,
			OtherModelsPctInfraCycle:     stats.OtherModelsPctInfraCycle,
			CursorModelsPctResearchToday: stats.CursorModelsPctResearchToday,
			OtherModelsPctResearchToday:  stats.OtherModelsPctResearchToday,
			CursorModelsPctResearchCycle: stats.CursorModelsPctResearchCycle,
			OtherModelsPctResearchCycle:  stats.OtherModelsPctResearchCycle,
			CursorModelsPctResearchOpen:  stats.CursorModelsPctResearchOpen,
			OtherModelsPctResearchOpen:   stats.OtherModelsPctResearchOpen,
			Hours:                        mapHours(stats.Hours),
			Days:                         mapHours(stats.Days),
		},
	}, nil
}

func mapBillingPeriods(periods []model.BillingPeriod) []statsdto.BillingPeriodResponse {
	out := make([]statsdto.BillingPeriodResponse, 0, len(periods))
	for _, p := range periods {
		out = append(out, statsdto.BillingPeriodResponse{
			StartISO:  p.StartISO,
			StartUnix: p.StartUnix,
			EndUnix:   p.EndUnix,
		})
	}
	return out
}

func (s *Service) GetTasks(ctx context.Context, query statsdto.TasksQuery) (*statsdto.TasksResponse, error) {
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
	if status != "all" && status != "in_progress" && status != "parked" && status != "completed" {
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

	tasks := computeTasks(events, members, time.Now())
	spend.ApplyCompletedSpendSharing(tasks, s.usageHours(members, events), s.repo.LoadUsageSamples())
	filtered := filterTasks(tasks, model.TaskQuery{
		Status: status,
		Alias:  strings.TrimSpace(query.Alias),
		Kind:   kind,
	})
	page := pageTasks(filtered, query.Limit, query.Offset)
	out := make([]statsdto.TaskResponse, 0, len(page))
	for _, t := range page {
		title := t.Title
		if title == "" {
			title = s.repo.TaskTitle(ctx, t.TaskID)
		}
		out = append(out, statsdto.TaskResponse{
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
			ClockPaused:     t.ClockPaused,
			CostUSD:         t.CostUSD,
			BudgetUSD:       t.BudgetUSD,
			OnDemandUSD:     t.OnDemandUSD,
			CursorModelsPct: t.CursorModelsPct,
			OtherModelsPct:  t.OtherModelsPct,
			SpendKind:       t.SpendKind,
			SpendShared:     t.SpendShared,
			RelatedTasks:    t.RelatedTasks,
			RelatedDocs:     t.RelatedDocs,
		})
	}
	return &statsdto.TasksResponse{Tasks: out, Total: len(filtered), Offset: query.Offset, Limit: query.Limit}, nil
}

func (s *Service) GetEvents(ctx context.Context, query eventdto.EventsQuery) (*eventdto.EventsResponse, error) {
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

	out := make([]eventdto.EventResponse, 0, len(events))
	for _, ev := range events {
		out = append(out, eventdto.EventResponse{
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
			Summary:         ev.Summary,
			Findings:        ev.Findings,
			AgentID:         ev.AgentID,
			ToAlias:         ev.ToAlias,
			ToAgent:         ev.ToAgent,
		})
	}

	return &eventdto.EventsResponse{Events: out, Total: total, Offset: query.Offset, Limit: query.Limit}, nil
}

func mapMembers(members []model.Member) []pulsedto.MemberResponse {
	out := make([]pulsedto.MemberResponse, 0, len(members))
	for _, m := range members {
		focus := m.Focus
		if focus == nil {
			focus = []string{}
		}
		services := m.Services
		if services == nil {
			services = []string{}
		}
		crews := m.Crews
		if crews == nil {
			crews = []string{}
		}
		item := pulsedto.MemberResponse{
			Alias:           m.Alias,
			Name:            m.Name,
			Role:            m.Role,
			Access:          m.Access,
			Focus:           focus,
			Crews:           crews,
			PulseScope:      m.PulseScope,
			Services:        services,
			Status:          m.Status,
			TaskID:          m.TaskID,
			TaskTitle:       m.TaskTitle,
			TaskDoc:         m.TaskDoc,
			TaskSummary:     m.TaskSummary,
			Branch:          m.Branch,
			UpdatedAt:       m.UpdatedAt,
			Repos:           mapRepos(m.Repos),
			CostUSD:         m.CostUSD,
			BudgetUSD:       m.BudgetUSD,
			OnDemandUSD:     m.OnDemandUSD,
			CursorModelsPct: m.CursorModelsPct,
			OtherModelsPct:  m.OtherModelsPct,
			SpendKind:       m.SpendKind,
			SpendShared:     m.SpendShared,
		}
		if len(m.Tasks) > 0 {
			item.Tasks = mapMemberTasks(m.Alias, m.Tasks)
			newest := m.Tasks[len(m.Tasks)-1]
			item.DurationSeconds = newest.DurationSeconds
			item.ClockPaused = newest.ClockPaused
		}
		if m.Research != nil {
			item.Research = &pulsedto.ResearchResponse{
				Status:          m.Research.Status,
				Summary:         m.Research.Summary,
				StartedAt:       m.Research.StartedAt,
				DurationSeconds: m.Research.DurationSeconds,
				ClockPaused:     m.Research.ClockPaused,
				CostUSD:         m.Research.CostUSD,
				BudgetUSD:       m.Research.BudgetUSD,
				OnDemandUSD:     m.Research.OnDemandUSD,
				CursorModelsPct: m.Research.CursorModelsPct,
				OtherModelsPct:  m.Research.OtherModelsPct,
				SpendShared:     m.Research.SpendShared,
				Chat:            mapChat(m.Research.Chat),
			}
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

func mapMemberTasks(alias string, tasks []model.MemberTask) []pulsedto.MemberTaskResponse {
	out := make([]pulsedto.MemberTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		services := task.Services
		if services == nil {
			services = []string{}
		}
		out = append(out, pulsedto.MemberTaskResponse{
			TaskID:          task.TaskID,
			TaskTitle:       task.Title,
			TaskDoc:         task.Doc,
			TaskSummary:     task.Summary,
			Branch:          task.Branch,
			Services:        services,
			StartedAt:       task.StartedAt,
			DurationSeconds: task.DurationSeconds,
			ClockPaused:     task.ClockPaused,
			Status:          task.Status,
			Repos:           mapRepos(task.Repos),
			CostUSD:         task.CostUSD,
			BudgetUSD:       task.BudgetUSD,
			OnDemandUSD:     task.OnDemandUSD,
			CursorModelsPct: task.CursorModelsPct,
			OtherModelsPct:  task.OtherModelsPct,
			SpendKind:       task.SpendKind,
			SpendShared:     task.SpendShared,
			Chats:           mapChats(task.Chats),
			Agents:          mapAgents(agentsOfSlot(alias, task)),
			RelatedTasks:    task.RelatedTasks,
			RelatedDocs:     task.RelatedDocs,
		})
	}
	return out
}

func mapAgents(agents []identity.Agent) []pulsedto.AgentResponse {
	if len(agents) == 0 {
		return nil
	}
	out := make([]pulsedto.AgentResponse, 0, len(agents))
	for _, a := range agents {
		if a.IsZero() {
			continue
		}
		out = append(out, pulsedto.AgentResponse{ID: a.ID, Kind: a.Kind, Title: a.Title})
	}
	return out
}

func mapChats(chats []model.ChatTab) []pulsedto.ChatTabResponse {
	if len(chats) == 0 {
		return nil
	}
	out := make([]pulsedto.ChatTabResponse, 0, len(chats))
	for _, chat := range chats {
		out = append(out, pulsedto.ChatTabResponse{Title: chat.Title, SessionID: chat.SessionID})
	}
	return out
}

func mapChat(chat *model.ChatTab) *pulsedto.ChatTabResponse {
	if chat == nil {
		return nil
	}
	return &pulsedto.ChatTabResponse{Title: chat.Title, SessionID: chat.SessionID}
}

func mapRepos(repos []model.RepoWork) []pulsedto.RepoWorkResponse {
	if len(repos) == 0 {
		return nil
	}
	out := make([]pulsedto.RepoWorkResponse, 0, len(repos))
	for _, repo := range repos {
		out = append(out, pulsedto.RepoWorkResponse{
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

func mapConflicts(conflicts []model.Conflict) []pulsedto.ConflictResponse {
	out := make([]pulsedto.ConflictResponse, 0, len(conflicts))
	for _, c := range conflicts {
		out = append(out, pulsedto.ConflictResponse{
			Severity:        c.Severity,
			Title:           c.Title,
			Description:     c.Description,
			AffectedAliases: c.AffectedAliases,
		})
	}
	return out
}

func mapStray(items []model.StrayRepo) []pulsedto.StrayRepoResponse {
	out := make([]pulsedto.StrayRepoResponse, 0, len(items))
	for _, item := range items {
		commits := make([]pulsedto.StrayCommitResponse, 0, len(item.Commits))
		for _, c := range item.Commits {
			commits = append(commits, pulsedto.StrayCommitResponse{Hash: c.Hash, Subject: c.Subject})
		}
		out = append(out, pulsedto.StrayRepoResponse{
			ID:      item.ID,
			Name:    item.Name,
			Repo:    item.Repo,
			Branch:  item.Branch,
			Commits: commits,
		})
	}
	return out
}

func (s *Service) usageHours(members []model.Member, events []model.Event) []model.HourSpend {
	return spend.UsageHours(s.repo.LoadUsageSamples(), members, events, time.Now())
}

func mapHours(hours []model.HourSpend) []statsdto.HourSpendResponse {
	if len(hours) == 0 {
		return nil
	}
	out := make([]statsdto.HourSpendResponse, 0, len(hours))
	for _, hour := range hours {
		parts := make([]statsdto.HourParticipantResponse, 0, len(hour.Participants))
		for _, p := range hour.Participants {
			parts = append(parts, statsdto.HourParticipantResponse{
				Kind:   p.Kind,
				ID:     p.ID,
				Title:  p.Title,
				Alias:  p.Alias,
				Shared: p.Shared,
			})
		}
		out = append(out, statsdto.HourSpendResponse{
			Hour:            hour.Hour,
			Alias:           hour.Alias,
			Aliases:         hour.Aliases,
			CursorModelsPct: hour.CursorModelsPct,
			OtherModelsPct:  hour.OtherModelsPct,
			BudgetUSD:       hour.BudgetUSD,
			OnDemandUSD:     hour.OnDemandUSD,
			Participants:    parts,
		})
	}
	return out
}
