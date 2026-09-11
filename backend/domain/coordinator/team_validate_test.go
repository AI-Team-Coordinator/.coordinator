package coordinator

import (
	"testing"

	"coordinator/model"
)

func TestNormalizeAndValidateTeam(t *testing.T) {
	services := map[string]struct{}{"core": {}, "web_chat": {}}

	members, err := normalizeAndValidateTeam([]model.TeamPerson{
		{Alias: "ek", Name: " Evgeny ", Role: "Founder", Access: "admin", Focus: []string{"core", "core", "web_chat"}},
	}, services)
	if err != nil {
		t.Fatal(err)
	}
	if members[0].Alias != "EK" || members[0].Role != "founder" || members[0].Access != "admin" || len(members[0].Focus) != 2 {
		t.Fatalf("unexpected member: %+v", members[0])
	}

	if _, err := normalizeAndValidateTeam(nil, services); err == nil {
		t.Fatal("expected error for empty team")
	}

	if _, err := normalizeAndValidateTeam([]model.TeamPerson{{Alias: "TOOLONG", Name: "X"}}, services); err == nil {
		t.Fatal("expected alias error")
	}

	if _, err := normalizeAndValidateTeam([]model.TeamPerson{
		{Alias: "EK", Name: "A", Focus: []string{"missing"}},
	}, services); err == nil {
		t.Fatal("expected unknown service error")
	}

	if _, err := normalizeAndValidateTeam([]model.TeamPerson{
		{Alias: "EK", Name: "A", Access: "root"},
	}, services); err == nil {
		t.Fatal("expected access error")
	}

	if _, err := normalizeAndValidateTeam([]model.TeamPerson{
		{Alias: "EK", Name: "A"},
		{Alias: "ek", Name: "B"},
	}, services); err == nil {
		t.Fatal("expected duplicate alias error")
	}
}
