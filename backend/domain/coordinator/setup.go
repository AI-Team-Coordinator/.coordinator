package coordinator

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"

	"coordinator/domain/coordinator/dto"
	"coordinator/domain/coordinator/repository"
	"coordinator/model"
)

type setupFiles interface {
	WriteCurrentAuthor(ctx context.Context, alias string) error
	GetLocale(ctx context.Context) (*model.LocaleFile, error)
	SaveLocale(ctx context.Context, loc *model.LocaleFile) error
	CoordinatorFile(ctx context.Context) model.CoordinatorFile
	SaveCoordinatorFile(ctx context.Context, cfg model.CoordinatorFile) error
	Collaboration() string
	SetupLayout() repository.SetupLayout
	ApplySetupFolders(docsRel, dataRel string) error
	GuessProjectName(ctx context.Context) string
	CoordinatorName(ctx context.Context) string
}

func needsSetup(profile *model.ProjectProfile) bool {
	if profile == nil {
		return true
	}
	if profile.Setup.Completed {
		return false
	}
	// Existing product maps (AlinaAssist and similar) skip the confirm screen.
	if strings.TrimSpace(profile.Project.Name) != "" && len(profile.Services) > 0 {
		return false
	}
	return true
}

func (s *Service) GetSetup(ctx context.Context) (*dto.SetupStateResponse, error) {
	profile, err := s.repo.GetProject(ctx)
	if err != nil {
		return nil, err
	}
	team, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	author, err := s.repo.CurrentAuthor(ctx)
	if err != nil {
		return nil, err
	}

	out := &dto.SetupStateResponse{
		Needed:          needsSetup(profile),
		CoordinatorName: repository.DefaultCoordinatorName(""),
		ProjectName:     strings.TrimSpace(profile.Project.Name),
		Language:        "en",
		Layout:          "existing",
		Local:           true,
		Collaboration:   model.CollaborationSolo,
	}
	if files, ok := s.repo.(setupFiles); ok {
		out.CoordinatorName = files.CoordinatorName(ctx)
		layout := files.SetupLayout()
		out.Layout = layout.Kind
		out.DocsDir = layout.DocsRel
		out.DataDir = layout.DataRel
		if out.ProjectName == "" {
			out.ProjectName = files.GuessProjectName(ctx)
		}
		if loc, err := files.GetLocale(ctx); err == nil && loc != nil {
			out.Language = normalizeSetupLanguage(loc.ChatLanguage)
		}
		if out.Needed {
			out.Collaboration = model.CollaborationSolo
		} else {
			out.Collaboration = files.Collaboration()
		}
	}
	if person := setupPrefillMember(team, author); person != nil {
		out.Alias = person.Alias
		out.Name = person.Name
		out.Role = person.Role
	} else if author != "" {
		out.Alias = strings.ToUpper(strings.TrimSpace(author))
	}
	if out.Role == "" {
		out.Role = "founder"
	}
	return out, nil
}

func (s *Service) CompleteSetup(ctx context.Context, req dto.CompleteSetupRequest) (*dto.SetupStateResponse, error) {
	profile, err := s.repo.GetProject(ctx)
	if err != nil {
		return nil, err
	}
	if !needsSetup(profile) {
		return s.GetSetup(ctx)
	}

	files, ok := s.repo.(setupFiles)
	if !ok {
		return nil, &ValidationError{Msg: "setup is not available on this store"}
	}

	layout := files.SetupLayout()
	docsRel := strings.TrimSpace(req.DocsDir)
	if docsRel == "" {
		docsRel = layout.DocsRel
	}
	dataRel := strings.TrimSpace(req.DataDir)
	if dataRel == "" {
		dataRel = layout.DataRel
	}
	if err := files.ApplySetupFolders(docsRel, dataRel); err != nil {
		return nil, &ValidationError{Msg: err.Error()}
	}

	projectName := strings.TrimSpace(req.ProjectName)
	if projectName == "" {
		return nil, &ValidationError{Msg: "project name is required"}
	}
	if len(projectName) > 80 {
		return nil, &ValidationError{Msg: "project name is too long"}
	}

	language := normalizeSetupLanguage(req.Language)
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "" {
		role = "founder"
	}

	alias := strings.ToUpper(strings.TrimSpace(req.Alias))
	name := strings.TrimSpace(req.Name)
	installer := model.TeamPerson{
		Alias:  alias,
		Name:   name,
		Role:   role,
		Access: "admin",
	}
	members, err := normalizeAndValidateTeam([]model.TeamPerson{installer}, nil)
	if err != nil {
		return nil, err
	}
	members[0].Access = "admin"

	existing, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	merged := mergeInstallerIntoTeam(existing, members[0])
	merged, err = normalizeAndValidateTeam(merged, nil)
	if err != nil {
		return nil, err
	}
	for i := range merged {
		if merged[i].Alias == members[0].Alias {
			merged[i].Access = "admin"
		}
	}

	if err := files.SaveLocale(ctx, &model.LocaleFile{
		Version:      1,
		ChatLanguage: language,
		DocsLanguage: language,
	}); err != nil {
		return nil, err
	}
	coordName := strings.TrimSpace(req.CoordinatorName)
	if coordName == "" {
		coordName = repository.DefaultCoordinatorName(language)
	}
	if utf8.RuneCountInString(coordName) > 40 {
		return nil, &ValidationError{Msg: "coordinator name is too long"}
	}
	cfg := files.CoordinatorFile(ctx)
	cfg.Name = coordName
	cfg.Collaboration = model.NormalizeCollaboration(req.Collaboration, model.CollaborationSolo)
	if err := files.SaveCoordinatorFile(ctx, cfg); err != nil {
		return nil, err
	}
	if err := s.repo.SaveTeam(ctx, merged); err != nil {
		return nil, err
	}
	if err := files.WriteCurrentAuthor(ctx, members[0].Alias); err != nil {
		return nil, err
	}

	profile.Project.Name = projectName
	if strings.TrimSpace(profile.Project.ID) == "" {
		profile.Project.ID = slugifyProjectID(projectName)
	}
	profile.Setup.Completed = true
	if profile.Version == 0 {
		profile.Version = 2
	}
	if err := s.repo.SaveProject(ctx, profile); err != nil {
		return nil, err
	}

	return s.GetSetup(ctx)
}

func setupPrefillMember(team []model.TeamPerson, author string) *model.TeamPerson {
	author = strings.ToUpper(strings.TrimSpace(author))
	if author != "" {
		for i := range team {
			if team[i].Alias == author {
				return &team[i]
			}
		}
	}
	if len(team) == 1 {
		return &team[0]
	}
	return nil
}

func mergeInstallerIntoTeam(existing []model.TeamPerson, installer model.TeamPerson) []model.TeamPerson {
	out := make([]model.TeamPerson, 0, len(existing)+1)
	found := false
	for _, person := range existing {
		if person.Alias == installer.Alias {
			person.Name = installer.Name
			person.Role = installer.Role
			person.Access = "admin"
			found = true
		}
		out = append(out, person)
	}
	if !found {
		out = append([]model.TeamPerson{installer}, out...)
	}
	return out
}

func normalizeSetupLanguage(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if strings.HasPrefix(raw, "ru") {
		return "ru"
	}
	return "en"
}

func slugifyProjectID(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if b.Len() > 0 && !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "project"
	}
	if len(out) > 48 {
		out = out[:48]
		out = strings.Trim(out, "-")
	}
	return out
}
