package repository

import (
	"sort"
	"strings"
	"time"

	"coordinator/model"
)

func overlayReportedRepos(live []model.RepoWork, report *model.GitReport) []model.RepoWork {
	liveByKey := make(map[string]model.RepoWork)
	workspace := make([]model.RepoWork, 0)
	for _, row := range live {
		if row.Kind == "workspace" {
			workspace = append(workspace, row)
			continue
		}
		liveByKey[reportRepoKey(row.ID, row.Repo)] = row
	}

	out := append([]model.RepoWork{}, workspace...)
	if report == nil || len(report.Repos) == 0 {
		for _, row := range live {
			if row.Kind == "workspace" {
				continue
			}
			row.Dirty = false
			out = append(out, row)
		}
		return out
	}

	used := make(map[string]struct{})
	for _, rr := range report.Repos {
		key := reportRepoKey(rr.ID, rr.Repo)
		used[key] = struct{}{}
		row, ok := liveByKey[key]
		if !ok {
			row = model.RepoWork{
				ID:    pickReportStr(rr.ID, rr.Repo),
				Name:  pickReportStr(rr.Name, rr.Repo, rr.ID),
				Repo:  pickReportStr(rr.Repo, rr.ID),
				State: "local",
			}
		}
		if row.Name == "" {
			row.Name = pickReportStr(rr.Name, rr.Repo, row.ID)
		}
		if row.ID == "" {
			row.ID = rr.ID
		}
		switch {
		case row.State == "merged" || row.State == "deployed":
			row.Dirty = false
		case row.State == "pushed" && rr.Unpushed == 0:
			row.Dirty = rr.Dirty
			row.Ahead = 0
		case rr.HasLocalBranch || rr.Dirty || rr.Unpushed > 0:
			row.State = "local"
			row.Ahead = rr.Unpushed
			row.Dirty = rr.Dirty
		default:
			row.Dirty = false
		}
		out = append(out, row)
	}
	for key, row := range liveByKey {
		if _, ok := used[key]; ok {
			continue
		}
		row.Dirty = false
		out = append(out, row)
	}
	return out
}

func buildGitReportFromRepos(repos []model.RepoWork, reportedAt time.Time) *model.GitReport {
	items := make([]model.GitReportRepo, 0)
	for _, repo := range repos {
		if repo.Kind == "workspace" {
			continue
		}
		unpushed := 0
		if repo.State == "local" {
			unpushed = repo.Ahead
		}
		items = append(items, model.GitReportRepo{
			ID:             repo.ID,
			Repo:           repo.Repo,
			Name:           repo.Name,
			HasLocalBranch: repo.HasLocalBranch,
			Dirty:          repo.Dirty,
			Unpushed:       unpushed,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return reportRepoKey(items[i].ID, items[i].Repo) < reportRepoKey(items[j].ID, items[j].Repo)
	})
	return &model.GitReport{ReportedAt: reportedAt.UTC(), Repos: items}
}

func gitReportEqual(a, b *model.GitReport) bool {
	ra := gitReportItems(a)
	rb := gitReportItems(b)
	if len(ra) != len(rb) {
		return false
	}
	byKey := make(map[string]model.GitReportRepo, len(ra))
	for _, item := range ra {
		byKey[reportRepoKey(item.ID, item.Repo)] = item
	}
	for _, item := range rb {
		prev, ok := byKey[reportRepoKey(item.ID, item.Repo)]
		if !ok || prev.Dirty != item.Dirty || prev.Unpushed != item.Unpushed || prev.HasLocalBranch != item.HasLocalBranch {
			return false
		}
	}
	return true
}

func gitReportItems(report *model.GitReport) []model.GitReportRepo {
	if report == nil {
		return nil
	}
	return report.Repos
}

func reportRepoKey(id, repo string) string {
	k := normServiceLabel(id)
	if k == "" {
		k = normServiceLabel(repo)
	}
	return k
}

func pickReportStr(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
