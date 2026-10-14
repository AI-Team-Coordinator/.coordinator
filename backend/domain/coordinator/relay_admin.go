package coordinator

import (
	"context"
	"strings"
)

func (s *Service) requireRelayAdmin(ctx context.Context) (adminAlias string, err error) {
	author, err := s.repo.CurrentAuthor(ctx)
	if err != nil {
		return "", err
	}
	adminAlias = strings.ToUpper(strings.TrimSpace(author))
	if adminAlias == "" {
		return "", &ValidationError{Msg: "set Common/data/.current_author before managing relay tokens"}
	}
	team, err := s.repo.GetTeam(ctx)
	if err != nil {
		return "", err
	}
	for _, person := range team {
		if strings.EqualFold(person.Alias, adminAlias) {
			if strings.EqualFold(strings.TrimSpace(person.Access), "admin") {
				return adminAlias, nil
			}
			return "", &ValidationError{Msg: "relay token management requires admin access"}
		}
	}
	return "", &ValidationError{Msg: "relay token management requires admin access"}
}
