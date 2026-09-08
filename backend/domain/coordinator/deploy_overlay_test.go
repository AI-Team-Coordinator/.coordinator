package coordinator

import (
	"testing"

	"coordinator/model"
)

func TestServiceMatchesRepo(t *testing.T) {
	core := model.RepoWork{ID: "core", Name: "Core", Repo: "Core"}
	inbox := model.RepoWork{ID: "inbox_panel_web", Name: "InboxPanelWeb", Repo: "InboxPanelWeb"}
	mobile := model.RepoWork{ID: "inbox_panel_mobile", Name: "InboxPanelMobile", Repo: "InboxPanelMobile"}

	if !serviceMatchesRepo("core", core) {
		t.Fatal("core")
	}
	if !serviceMatchesRepo("inbox_panel", inbox) {
		t.Fatal("inbox_panel -> web")
	}
	if serviceMatchesRepo("inbox_panel", mobile) {
		t.Fatal("inbox_panel must not match mobile")
	}
	if !serviceMatchesRepo("InboxPanelWeb", inbox) {
		t.Fatal("name match")
	}
}

func TestApplyMergeEventsPromotesEmptyLocal(t *testing.T) {
	members := []model.Member{{
		Status: "in_progress",
		TaskID: "T1",
		Repos:  []model.RepoWork{{ID: "core", Name: "Core", Repo: "Core", State: "local"}},
	}}
	applyMergeEvents(members, []model.Event{{Event: "repo_merged", TaskID: "T1", Service: "Core"}})
	if members[0].Repos[0].State != "merged" {
		t.Fatalf("state=%s", members[0].Repos[0].State)
	}
}

func TestApplyMergeEventsKeepsLocalAhead(t *testing.T) {
	members := []model.Member{{
		Status: "in_progress",
		TaskID: "T1",
		Repos:  []model.RepoWork{{ID: "core", Name: "Core", Repo: "Core", State: "local", Ahead: 2}},
	}}
	applyMergeEvents(members, []model.Event{{Event: "repo_merged", TaskID: "T1", Service: "Core"}})
	if members[0].Repos[0].State != "local" {
		t.Fatalf("state=%s", members[0].Repos[0].State)
	}
}

func TestApplyMergeEventsKeepsPushed(t *testing.T) {
	members := []model.Member{{
		Status: "in_progress",
		TaskID: "T1",
		Repos:  []model.RepoWork{{ID: "core", Name: "Core", Repo: "Core", State: "pushed"}},
	}}
	applyMergeEvents(members, []model.Event{{Event: "repo_merged", TaskID: "T1", Service: "Core"}})
	if members[0].Repos[0].State != "pushed" {
		t.Fatalf("state=%s", members[0].Repos[0].State)
	}
}

func TestAllReposDeployed(t *testing.T) {
	m := model.Member{
		Status: "in_progress",
		Repos: []model.RepoWork{
			{State: "merged", Deployed: true},
			{State: "merged", Deployed: true},
		},
	}
	if !allReposDeployed(m) {
		t.Fatal("expected deployed")
	}
	m.Repos[1].Deployed = false
	if allReposDeployed(m) {
		t.Fatal("expected not deployed")
	}

	infra := model.Member{
		Status: "in_progress",
		Repos:  []model.RepoWork{{Kind: "workspace", State: "infra"}},
	}
	if allReposDeployed(infra) {
		t.Fatal("workspace-only task is not a product deploy")
	}
}
