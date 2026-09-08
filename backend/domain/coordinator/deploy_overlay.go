package coordinator

import (
	"strings"

	"coordinator/model"
)

func applyMergeEvents(members []model.Member, events []model.Event) {
	for i := range members {
		if members[i].Status != "in_progress" || members[i].TaskID == "" {
			continue
		}
		for j := range members[i].Repos {
			if members[i].Repos[j].Kind == "workspace" {
				continue
			}
			if members[i].Repos[j].State == "merged" {
				continue
			}
			if members[i].Repos[j].State == "pushed" {
				continue
			}
			if members[i].Repos[j].State == "local" && members[i].Repos[j].Ahead > 0 {
				continue
			}
			if repoMergedEvent(members[i].Repos[j], members[i].TaskID, events) {
				members[i].Repos[j].State = "merged"
			}
		}
	}
}

func repoMergedEvent(repo model.RepoWork, taskID string, events []model.Event) bool {
	for _, ev := range events {
		if ev.Event != "repo_merged" {
			continue
		}
		if ev.TaskID != "" && ev.TaskID != taskID {
			continue
		}
		if serviceMatchesRepo(ev.Service, repo) || serviceMatchesRepo(ev.Repo, repo) {
			return true
		}
	}
	return false
}

func applyDeployEvents(members []model.Member, events []model.Event) {
	for i := range members {
		if members[i].Status != "in_progress" || members[i].TaskID == "" {
			continue
		}
		for j := range members[i].Repos {
			if members[i].Repos[j].Kind == "workspace" {
				continue
			}
			if members[i].Repos[j].State != "merged" {
				continue
			}
			if repoDeployed(members[i].Repos[j], members[i].TaskID, events) {
				members[i].Repos[j].Deployed = true
			}
		}
	}
}

func repoDeployed(repo model.RepoWork, taskID string, events []model.Event) bool {
	for _, ev := range events {
		if ev.Event != "deploy_finished" {
			continue
		}
		if ev.TaskID != "" && ev.TaskID != taskID {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(ev.Status))
		if status != "" && status != "finished" {
			continue
		}
		if serviceMatchesRepo(ev.Service, repo) {
			return true
		}
	}
	return false
}

func serviceMatchesRepo(label string, repo model.RepoWork) bool {
	n := normService(label)
	if n == "" {
		return false
	}
	for _, key := range []string{repo.ID, repo.Name, repo.Repo} {
		k := normService(key)
		if k == "" {
			continue
		}
		if n == k {
			return true
		}
		if strings.HasSuffix(k, "_web") && strings.TrimSuffix(k, "_web") == n {
			return true
		}
	}
	return false
}

func normService(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func allReposDeployed(member model.Member) bool {
	if member.Status != "in_progress" || len(member.Repos) == 0 {
		return false
	}
	product := 0
	for _, repo := range member.Repos {
		if repo.Kind == "workspace" {
			continue
		}
		product++
		if repo.State != "merged" || !repo.Deployed {
			return false
		}
	}
	if product == 0 {
		return false
	}
	return true
}
