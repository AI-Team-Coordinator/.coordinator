package coordinator

import (
	"fmt"
	"regexp"
	"strings"

	"coordinator/model"
)

var (
	aliasPattern   = regexp.MustCompile(`^[A-Z]{2,4}$`)
	rolePattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
	slugIDPattern  = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	pulseScopeAllowed = map[string]struct{}{
		"member":    {},
		"executive": {},
	}
)

func normalizeAndValidateTeam(members []model.TeamPerson, serviceIDs map[string]struct{}, crewIDs map[string]struct{}) ([]model.TeamPerson, error) {
	if len(members) == 0 {
		return nil, &ValidationError{Msg: "team must have at least one member"}
	}

	out := make([]model.TeamPerson, 0, len(members))
	seen := make(map[string]struct{}, len(members))

	for i, raw := range members {
		alias := strings.ToUpper(strings.TrimSpace(raw.Alias))
		name := strings.TrimSpace(raw.Name)
		role := strings.ToLower(strings.TrimSpace(raw.Role))

		if !aliasPattern.MatchString(alias) {
			return nil, &ValidationError{Msg: fmt.Sprintf("member %d: alias must be 2–4 Latin letters", i+1)}
		}
		if name == "" {
			return nil, &ValidationError{Msg: fmt.Sprintf("member %s: name is required", alias)}
		}
		if len(name) > 80 {
			return nil, &ValidationError{Msg: fmt.Sprintf("member %s: name is too long", alias)}
		}
		if role != "" && !rolePattern.MatchString(role) {
			return nil, &ValidationError{Msg: fmt.Sprintf("member %s: invalid role", alias)}
		}
		access := strings.ToLower(strings.TrimSpace(raw.Access))
		if access != "" && access != "admin" && access != "member" {
			return nil, &ValidationError{Msg: fmt.Sprintf("member %s: access must be admin or member", alias)}
		}
		if _, dup := seen[alias]; dup {
			return nil, &ValidationError{Msg: fmt.Sprintf("duplicate alias %s", alias)}
		}
		seen[alias] = struct{}{}

		focus := make([]string, 0, len(raw.Focus))
		focusSeen := make(map[string]struct{})
		for _, id := range raw.Focus {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := serviceIDs[id]; !ok && len(serviceIDs) > 0 {
				return nil, &ValidationError{Msg: fmt.Sprintf("member %s: unknown service %s", alias, id)}
			}
			if _, ok := focusSeen[id]; ok {
				continue
			}
			focusSeen[id] = struct{}{}
			focus = append(focus, id)
		}

		crews := make([]string, 0, len(raw.Crews))
		crewSeen := make(map[string]struct{})
		for _, id := range raw.Crews {
			id = strings.ToLower(strings.TrimSpace(id))
			if id == "" {
				continue
			}
			if crewIDs != nil {
				if _, ok := crewIDs[id]; !ok {
					return nil, &ValidationError{Msg: fmt.Sprintf("member %s: unknown crew %s", alias, id)}
				}
			}
			if _, ok := crewSeen[id]; ok {
				continue
			}
			crewSeen[id] = struct{}{}
			crews = append(crews, id)
		}

		primaryCrew := strings.ToLower(strings.TrimSpace(raw.PrimaryCrew))
		if primaryCrew != "" {
			if crewIDs != nil {
				if _, ok := crewIDs[primaryCrew]; !ok {
					return nil, &ValidationError{Msg: fmt.Sprintf("member %s: unknown primary_crew %s", alias, primaryCrew)}
				}
			}
			if _, ok := crewSeen[primaryCrew]; !ok {
				crews = append(crews, primaryCrew)
			}
		}

		pulseScope := strings.ToLower(strings.TrimSpace(raw.PulseScope))
		if pulseScope == "" {
			pulseScope = "member"
		}
		if _, ok := pulseScopeAllowed[pulseScope]; !ok {
			return nil, &ValidationError{Msg: fmt.Sprintf("member %s: pulse_scope must be member or executive", alias)}
		}

		out = append(out, model.TeamPerson{
			Alias:       alias,
			Name:        name,
			Role:        role,
			Access:      access,
			Focus:       focus,
			Crews:       crews,
			PrimaryCrew: primaryCrew,
			PulseScope:  pulseScope,
		})
	}

	return out, nil
}

func teamAccessByAlias(members []model.TeamPerson) map[string]string {
	out := make(map[string]string, len(members))
	for _, person := range members {
		if person.Alias == "" || person.Access == "" {
			continue
		}
		out[person.Alias] = person.Access
	}
	return out
}

func preserveTeamAccess(members []model.TeamPerson, previous map[string]string) []model.TeamPerson {
	if len(previous) == 0 {
		return members
	}
	for i := range members {
		if members[i].Access != "" {
			continue
		}
		members[i].Access = previous[members[i].Alias]
	}
	return members
}
