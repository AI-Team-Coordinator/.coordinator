package repository

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"coordinator/domain/claims"
	slotpkg "coordinator/domain/slot"
	"coordinator/model"
)

func resolveInspectBranch(repoPath, ownerAlias, slotBranch string) string {
	for _, b := range slotpkg.BranchCandidates(ownerAlias, slotBranch) {
		if gitRefExists(repoPath, "refs/heads/"+b) || gitRefExists(repoPath, "refs/remotes/origin/"+b) {
			return b
		}
	}
	return strings.TrimSpace(slotBranch)
}

func (r *FileRepository) attachRepoWork(members []model.Member) {
	profile, err := r.GetProject(context.Background())
	if err != nil || profile == nil {
		return
	}
	workspace := r.WorkspaceDir()
	for i := range members {
		slots := members[i].Slots()
		if len(slots) == 0 {
			continue
		}
		all := make([]model.RepoWork, 0)
		for si := range slots {
			repos := inspectTaskRepos(workspace, profile.Services, slots[si].Services, slots[si].Branch, slots[si].TaskID, members[i].Alias)
			for ri := range repos {
				repos[ri].TaskID = slots[si].TaskID
			}
			slots[si].Repos = repos
			all = append(all, repos...)
		}
		members[i].Tasks = slots
		members[i].Repos = all
	}
}

func (r *FileRepository) GetStrayWork(_ context.Context, members []model.Member) ([]model.StrayRepo, error) {
	profile, err := r.GetProject(context.Background())
	if err != nil || profile == nil {
		return []model.StrayRepo{}, err
	}
	claimed := claimedBranches(members)
	workspace := r.WorkspaceDir()
	out := make([]model.StrayRepo, 0)
	seen := make(map[string]struct{})
	for _, svc := range profile.Services {
		if svc.Group == "workspace" || svc.Kind == "workspace" {
			continue
		}
		folder := strings.TrimSpace(svc.Repo)
		if folder == "" || folder == "." {
			continue
		}
		repoPath := filepath.Join(workspace, folder)
		if !isGitRepo(repoPath) {
			continue
		}
		for _, branch := range topicBranches(repoPath) {
			if _, taken := claimed[branch]; taken {
				continue
			}
			key := folder + "\x00" + branch
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			commits := branchCommits(repoPath, branch)
			if len(commits) == 0 {
				continue
			}
			out = append(out, model.StrayRepo{
				ID:      svc.ID,
				Name:    svc.Name,
				Repo:    folder,
				Branch:  branch,
				Commits: commits,
			})
		}
	}
	return out, nil
}

func claimedBranches(members []model.Member) map[string]struct{} {
	out := make(map[string]struct{})
	for _, m := range members {
		for _, slot := range m.Slots() {
			for _, b := range slotpkg.BranchCandidates(m.Alias, slot.Branch) {
				if b != "" {
					out[b] = struct{}{}
				}
			}
		}
	}
	return out
}

func inspectTaskRepos(workspace string, services []model.ServiceNode, claimed []string, branch, taskID, ownerAlias string) []model.RepoWork {
	out := make([]model.RepoWork, 0)
	seen := make(map[string]struct{})
	for _, svc := range reposToInspect(services, claimed, branch) {
		folder := strings.TrimSpace(svc.Repo)
		if folder == "" || folder == "." {
			continue
		}
		if _, dup := seen[folder]; dup {
			continue
		}
		repoPath := filepath.Join(workspace, folder)
		if !isGitRepo(repoPath) {
			continue
		}
		work, ok := inspectOneRepo(repoPath, svc, branch, taskID, ownerAlias)
		if !ok {
			continue
		}
		seen[folder] = struct{}{}
		out = append(out, work)
	}
	return out
}

func reposToInspect(services []model.ServiceNode, claimed []string, branch string) []model.ServiceNode {
	matched := matchClaimedServices(services, claimed)
	if len(matched) > 0 {
		return matched
	}
	if isTrunkBranch(branch) {
		return nil
	}
	out := make([]model.ServiceNode, 0)
	for _, svc := range services {
		if isWorkspaceService(svc) {
			continue
		}
		out = append(out, svc)
	}
	return out
}

func matchClaimedServices(services []model.ServiceNode, claimed []string) []model.ServiceNode {
	out := make([]model.ServiceNode, 0)
	seen := make(map[string]struct{})
	for _, label := range claimed {
		for _, svc := range services {
			if !serviceLabelMatches(label, svc) {
				continue
			}
			key := strings.TrimSpace(svc.Repo)
			if key == "" {
				key = svc.ID
			}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, svc)
		}
	}
	return out
}

func serviceLabelMatches(label string, svc model.ServiceNode) bool {
	want := claims.Parse(label)
	if want.IsZero() {
		return false
	}
	for _, key := range []string{svc.ID, svc.Name, svc.Repo, svc.GitHubRepo} {
		got := claims.Parse(key)
		if got.Key == want.Key {
			return true
		}
	}
	return false
}

func isWorkspaceService(svc model.ServiceNode) bool {
	return svc.Group == "workspace" || svc.Kind == "workspace"
}

func isTrunkBranch(branch string) bool {
	b := strings.ToLower(strings.TrimSpace(branch))
	return b == "main" || b == "master"
}

func inspectOneRepo(repoPath string, svc model.ServiceNode, branch, taskID, ownerAlias string) (model.RepoWork, bool) {
	gitBranch := resolveInspectBranch(repoPath, ownerAlias, branch)
	if isWorkspaceService(svc) {
		return inspectWorkspaceRepo(repoPath, svc, gitBranch, ownerAlias, branch)
	}
	if isTrunkBranch(branch) {
		if taskOnMain(repoPath, taskID) {
			return productRepoWork(svc, "merged", 0), true
		}
		return model.RepoWork{}, false
	}

	hasLocal := gitRefExists(repoPath, "refs/heads/"+gitBranch)
	hasRemote := gitRefExists(repoPath, "refs/remotes/origin/"+gitBranch)
	if !hasLocal && !hasRemote {
		if taskOnMain(repoPath, taskID) {
			return productRepoWork(svc, "merged", 0), true
		}
		return model.RepoWork{}, false
	}

	tip := gitBranch
	if !hasLocal && hasRemote {
		tip = "origin/" + gitBranch
	}
	ahead := countCommits(repoPath, "origin/main.."+tip)
	unpushed := 0
	if hasRemote && hasLocal {
		unpushed = countCommits(repoPath, "origin/"+gitBranch+".."+gitBranch)
	}
	state, shownAhead := classifyProductRepo(ahead, taskOnMain(repoPath, taskID), hasRemote, unpushed)
	work := productRepoWork(svc, state, shownAhead)
	work.Dirty = repoDirtyOnBranch(repoPath, ownerAlias, branch)
	work.HasLocalBranch = hasLocal
	return work, true
}

// classifyProductRepo maps live git facts to a pulse state.
// ahead==0 is not merged by itself: merged needs the task id already on origin/main.
func classifyProductRepo(ahead int, onMain, hasRemote bool, unpushed int) (state string, shownAhead int) {
	if ahead == 0 {
		if onMain {
			return "merged", 0
		}
		return "local", 0
	}
	if hasRemote && unpushed == 0 {
		return "pushed", 0
	}
	if hasRemote {
		return "local", unpushed
	}
	return "local", ahead
}

func repoDirtyOnBranch(repoPath, ownerAlias, slotBranch string) bool {
	cur := currentBranch(repoPath)
	if !slotpkg.BranchMatches(ownerAlias, slotBranch, cur) {
		return false
	}
	return repoDirty(repoPath)
}

func currentBranch(dir string) string {
	out, err := gitOutput(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func repoDirty(dir string) bool {
	out, err := gitOutput(dir, "status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

func inspectWorkspaceRepo(repoPath string, svc model.ServiceNode, gitBranch, _ownerAlias, slotBranch string) (model.RepoWork, bool) {
	ahead := 0
	if isTrunkBranch(slotBranch) || strings.TrimSpace(slotBranch) == "" {
		ahead = countCommits(repoPath, "origin/main..HEAD")
	} else {
		tip := gitBranch
		if tip == "" {
			tip = slotBranch
		}
		if !gitRefExists(repoPath, "refs/heads/"+tip) && gitRefExists(repoPath, "refs/remotes/origin/"+tip) {
			tip = "origin/" + tip
		}
		ahead = countCommits(repoPath, "origin/main.."+tip)
	}
	state := "infra"
	if ahead > 0 {
		state = "local"
	}
	return model.RepoWork{
		ID:    svc.ID,
		Name:  svc.Name,
		Repo:  svc.Repo,
		Kind:  "workspace",
		State: state,
		Ahead: ahead,
	}, true
}

func productRepoWork(svc model.ServiceNode, state string, ahead int) model.RepoWork {
	return model.RepoWork{
		ID:    svc.ID,
		Name:  svc.Name,
		Repo:  svc.Repo,
		State: state,
		Ahead: ahead,
	}
}

func isGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && (info.IsDir() || info.Mode().IsRegular())
}

func gitRefExists(dir, ref string) bool {
	_, err := gitOutput(dir, "show-ref", "--verify", "--quiet", ref)
	return err == nil
}

func taskOnMain(dir, taskID string) bool {
	if taskID == "" {
		return false
	}
	out, err := gitOutput(dir, "log", "origin/main", "--grep="+taskID, "-1", "--pretty=%h")
	return err == nil && strings.TrimSpace(out) != ""
}

func countCommits(dir, spec string) int {
	out, err := gitOutput(dir, "rev-list", "--count", spec)
	if err != nil {
		return 0
	}
	n, convErr := strconv.Atoi(strings.TrimSpace(out))
	if convErr != nil || n < 0 {
		return 0
	}
	return n
}

func topicBranches(dir string) []string {
	out, err := gitOutput(dir, "for-each-ref", "--format=%(refname:short)", "refs/heads/feat/", "refs/heads/fix/")
	if err != nil {
		return nil
	}
	branches := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches
}

func branchCommits(dir, branch string) []model.StrayCommit {
	out, err := gitOutput(dir, "log", "origin/main.."+branch, "--pretty=%h\t%s")
	if err != nil {
		return nil
	}
	commits := make([]model.StrayCommit, 0)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		hash, subject, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		commits = append(commits, model.StrayCommit{Hash: hash, Subject: subject})
		if len(commits) >= 8 {
			break
		}
	}
	return commits
}

func gitOutput(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
