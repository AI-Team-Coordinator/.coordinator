package coordinator

import (
	"testing"

	"coordinator/model"
)

func TestNormalizeAndValidateOrgTopology(t *testing.T) {
	topology, err := normalizeAndValidateOrgTopology(model.OrgTopology{
		Crews: []model.Crew{
			{ID: "qa", Label: model.LocalizedText{EN: "QA"}, Kind: "internal"},
			{ID: "core-dev", Label: model.LocalizedText{RU: "Core"}, Kind: "internal"},
		},
		CrewConnectors: []model.CrewConnector{
			{
				ID:         "qa-watches-core",
				FromCrewID: "qa",
				ToCrewID:   "core-dev",
				Interests:  []model.CrewConnectorItem{{Title: "Releases"}},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(topology.Crews) != 2 || len(topology.CrewConnectors) != 1 {
		t.Fatalf("unexpected topology: %+v", topology)
	}

	if _, err := normalizeAndValidateOrgTopology(model.OrgTopology{
		CrewConnectors: []model.CrewConnector{
			{ID: "x", FromCrewID: "a", ToCrewID: "b", Interests: []model.CrewConnectorItem{{Title: "t"}}},
		},
	}); err == nil {
		t.Fatal("expected unknown crew error")
	}
}

func TestValidateTopologyAgainstTeam(t *testing.T) {
	topology := model.OrgTopology{
		Crews: []model.Crew{{ID: "qa", Label: model.LocalizedText{EN: "QA"}, Kind: "internal"}},
	}
	if err := validateTopologyAgainstTeam(topology, []model.TeamPerson{
		{Alias: "EK", Crews: []string{"missing"}},
	}); err == nil {
		t.Fatal("expected member crew error")
	}
}
