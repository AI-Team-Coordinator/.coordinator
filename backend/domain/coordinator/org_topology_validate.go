package coordinator

import (
	"fmt"
	"strings"

	"coordinator/model"
)

var crewKindAllowed = map[string]struct{}{
	"internal":   {},
	"affiliated": {},
	"external":   {},
}

func normalizeAndValidateOrgTopology(raw model.OrgTopology) (model.OrgTopology, error) {
	out := model.OrgTopology{
		Version:        1,
		Crews:          make([]model.Crew, 0, len(raw.Crews)),
		CrewConnectors: make([]model.CrewConnector, 0, len(raw.CrewConnectors)),
	}

	crewIDs := make(map[string]struct{})
	for i, c := range raw.Crews {
		id := strings.ToLower(strings.TrimSpace(c.ID))
		if !slugIDPattern.MatchString(id) {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew %d: invalid id", i+1)}
		}
		if _, dup := crewIDs[id]; dup {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("duplicate crew id %s", id)}
		}
		crewIDs[id] = struct{}{}

		kind := strings.ToLower(strings.TrimSpace(c.Kind))
		if kind == "" {
			kind = "internal"
		}
		if _, ok := crewKindAllowed[kind]; !ok {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew %s: invalid kind", id)}
		}
		label := normalizeLocalized(c.Label)
		if strings.TrimSpace(label.EN) == "" && strings.TrimSpace(label.RU) == "" {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew %s: label is required", id)}
		}

		out.Crews = append(out.Crews, model.Crew{
			ID:            id,
			Label:         label,
			Kind:          kind,
			InterestBrief: strings.TrimSpace(c.InterestBrief),
			Notes:         strings.TrimSpace(c.Notes),
		})
	}

	connectorIDs := make(map[string]struct{})
	pairSeen := make(map[string]struct{})
	for i, link := range raw.CrewConnectors {
		id := strings.ToLower(strings.TrimSpace(link.ID))
		if !slugIDPattern.MatchString(id) {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %d: invalid id", i+1)}
		}
		if _, dup := connectorIDs[id]; dup {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("duplicate crew_connector id %s", id)}
		}
		connectorIDs[id] = struct{}{}

		from := strings.ToLower(strings.TrimSpace(link.FromCrewID))
		to := strings.ToLower(strings.TrimSpace(link.ToCrewID))
		if from == "" || to == "" {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %s: from_crew_id and to_crew_id required", id)}
		}
		if from == to {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %s: from and to must differ", id)}
		}
		if _, ok := crewIDs[from]; !ok {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %s: unknown from_crew_id %s", id, from)}
		}
		if _, ok := crewIDs[to]; !ok {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %s: unknown to_crew_id %s", id, to)}
		}
		pairKey := from + "->" + to
		if _, dup := pairSeen[pairKey]; dup {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("duplicate link %s to %s", from, to)}
		}
		pairSeen[pairKey] = struct{}{}

		interests := make([]model.CrewConnectorItem, 0, len(link.Interests))
		for j, item := range link.Interests {
			title := strings.TrimSpace(item.Title)
			if title == "" {
				return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %s: interests[%d] title required", id, j)}
			}
			if len(title) > 200 {
				return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %s: interests[%d] title too long", id, j)}
			}
			tags := make([]string, 0, len(item.Tags))
			tagSeen := make(map[string]struct{})
			for _, tag := range item.Tags {
				tag = strings.TrimSpace(tag)
				if tag == "" {
					continue
				}
				if _, ok := tagSeen[tag]; ok {
					continue
				}
				tagSeen[tag] = struct{}{}
				tags = append(tags, tag)
			}
			interests = append(interests, model.CrewConnectorItem{
				Title:  title,
				Detail: strings.TrimSpace(item.Detail),
				Tags:   tags,
			})
		}
		if len(interests) == 0 {
			return model.OrgTopology{}, &ValidationError{Msg: fmt.Sprintf("crew_connector %s: at least one interest required", id)}
		}

		out.CrewConnectors = append(out.CrewConnectors, model.CrewConnector{
			ID:         id,
			FromCrewID: from,
			ToCrewID:   to,
			Summary:    strings.TrimSpace(link.Summary),
			Notes:      strings.TrimSpace(link.Notes),
			Interests:  interests,
		})
	}

	return out, nil
}

func crewIDSet(topology model.OrgTopology) map[string]struct{} {
	out := make(map[string]struct{}, len(topology.Crews))
	for _, c := range topology.Crews {
		if c.ID != "" {
			out[c.ID] = struct{}{}
		}
	}
	return out
}

func normalizeLocalized(l model.LocalizedText) model.LocalizedText {
	return model.LocalizedText{
		EN: strings.TrimSpace(l.EN),
		RU: strings.TrimSpace(l.RU),
	}
}

func validateTopologyAgainstTeam(topology model.OrgTopology, team []model.TeamPerson) error {
	crewIDs := crewIDSet(topology)
	for _, person := range team {
		alias := strings.TrimSpace(person.Alias)
		if alias == "" {
			continue
		}
		for _, id := range person.Crews {
			if _, ok := crewIDs[id]; !ok {
				return &ValidationError{Msg: fmt.Sprintf("member %s references unknown crew %s", alias, id)}
			}
		}
		if pc := strings.TrimSpace(person.PrimaryCrew); pc != "" {
			if _, ok := crewIDs[pc]; !ok {
				return &ValidationError{Msg: fmt.Sprintf("member %s references unknown primary_crew %s", alias, pc)}
			}
		}
	}
	return nil
}
