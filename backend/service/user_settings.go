package service

import (
	"context"
	"coordinator/domain/validate"
	"os"
	"strings"

	"coordinator/domain/org"
	teamrules "coordinator/domain/team"
	"coordinator/model"
	settingsdto "coordinator/service/settings/dto"
)

type userSettingsFiles interface {
	GetLocale(ctx context.Context) (*model.LocaleFile, error)
	SaveLocale(ctx context.Context, loc *model.LocaleFile) error
	UserAvatarExists(alias string) bool
	ReadUserAvatar(alias string) ([]byte, error)
	WriteUserAvatar(alias string, data []byte) error
	DeleteUserAvatar(alias string) error
}

func (s *Service) GetUserSettings(ctx context.Context) (*settingsdto.UserSettingsResponse, error) {
	author, err := s.repo.CurrentAuthor(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(author) == "" {
		return nil, &validate.Error{Msg: "set Common/data/.current_author before opening settings"}
	}

	out := &settingsdto.UserSettingsResponse{
		Alias:    strings.ToUpper(strings.TrimSpace(author)),
		Language: "en",
	}
	if files, ok := s.repo.(userSettingsFiles); ok {
		if loc, err := files.GetLocale(ctx); err == nil && loc != nil {
			out.Language = normalizeSetupLanguage(loc.ChatLanguage)
		}
		if files.UserAvatarExists(out.Alias) {
			out.AvatarURL = "/api/v1/me/avatar"
		}
	}

	team, err := s.repo.GetTeam(ctx)
	if err != nil {
		return nil, err
	}
	for _, person := range team {
		if strings.EqualFold(person.Alias, out.Alias) {
			out.Name = person.Name
			out.Role = person.Role
			out.Access = person.Access
			break
		}
	}
	if out.Name == "" {
		out.Name = out.Alias
	}
	return out, nil
}

func (s *Service) SaveUserSettings(ctx context.Context, req settingsdto.SaveUserSettingsRequest) (*settingsdto.UserSettingsResponse, error) {
	current, err := s.GetUserSettings(ctx)
	if err != nil {
		return nil, err
	}

	files, ok := s.repo.(userSettingsFiles)
	if !ok {
		return nil, &validate.Error{Msg: "user settings are not available on this store"}
	}

	language := normalizeSetupLanguage(req.Language)
	if language == "" {
		language = current.Language
	}
	if language != current.Language {
		if err := files.SaveLocale(ctx, &model.LocaleFile{
			Version:      1,
			ChatLanguage: language,
			DocsLanguage: language,
		}); err != nil {
			return nil, err
		}
	}

	name := strings.TrimSpace(req.Name)
	if name != "" && name != current.Name {
		team, err := s.repo.GetTeam(ctx)
		if err != nil {
			return nil, err
		}
		previousAccess := teamrules.AccessByAlias(team)
		found := false
		for i, person := range team {
			if strings.EqualFold(person.Alias, current.Alias) {
				team[i].Name = name
				found = true
				break
			}
		}
		if !found {
			team = append(team, model.TeamPerson{Alias: current.Alias, Name: name, Role: current.Role})
		}
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
		topology, err := s.repo.GetOrgTopology(ctx)
		if err != nil {
			return nil, err
		}
		members, err := teamrules.Normalize(team, serviceIDs, org.CrewIDs(topology))
		if err != nil {
			return nil, err
		}
		members = teamrules.PreserveAccess(members, previousAccess)
		if err := s.repo.SaveTeam(ctx, members); err != nil {
			return nil, err
		}
		_ = s.settingsPushWarning(ctx)
	}

	return s.GetUserSettings(ctx)
}

func (s *Service) SaveUserAvatar(ctx context.Context, data []byte) (*settingsdto.UserSettingsResponse, error) {
	if len(data) == 0 {
		return nil, &validate.Error{Msg: "empty avatar"}
	}
	if len(data) > 512<<10 {
		return nil, &validate.Error{Msg: "avatar too large (max 512 KB)"}
	}
	current, err := s.GetUserSettings(ctx)
	if err != nil {
		return nil, err
	}
	files, ok := s.repo.(userSettingsFiles)
	if !ok {
		return nil, &validate.Error{Msg: "avatar upload is not available on this store"}
	}
	if err := files.WriteUserAvatar(current.Alias, data); err != nil {
		return nil, err
	}
	return s.GetUserSettings(ctx)
}

func (s *Service) ReadUserAvatar(ctx context.Context) ([]byte, error) {
	current, err := s.GetUserSettings(ctx)
	if err != nil {
		return nil, err
	}
	files, ok := s.repo.(userSettingsFiles)
	if !ok || !files.UserAvatarExists(current.Alias) {
		return nil, os.ErrNotExist
	}
	return files.ReadUserAvatar(current.Alias)
}

func (s *Service) DeleteUserAvatar(ctx context.Context) (*settingsdto.UserSettingsResponse, error) {
	current, err := s.GetUserSettings(ctx)
	if err != nil {
		return nil, err
	}
	files, ok := s.repo.(userSettingsFiles)
	if !ok {
		return nil, &validate.Error{Msg: "avatar upload is not available on this store"}
	}
	if err := files.DeleteUserAvatar(current.Alias); err != nil {
		return nil, err
	}
	return s.GetUserSettings(ctx)
}
