package coordinator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"coordinator/domain/coordinator/dto"
	"coordinator/model"
)

var (
	serviceIDPattern   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,47}$`)
	serviceKindPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
	githubSlugPattern  = regexp.MustCompile(`^[A-Za-z0-9._][A-Za-z0-9._-]{0,99}$`)
	localFolderPattern = regexp.MustCompile(`^[A-Za-z0-9._][A-Za-z0-9._-]{0,79}$`)
)

func normalizeCreateService(profile *model.ProjectProfile, req dto.CreateServiceRequest) (model.ServiceNode, error) {
	if profile == nil {
		return model.ServiceNode{}, &ValidationError{Msg: "project profile is missing"}
	}
	if strings.TrimSpace(profile.GitHub.Org) == "" {
		return model.ServiceNode{}, &ValidationError{Msg: "GitHub organization is not bound"}
	}

	id := strings.TrimSpace(req.ID)
	name := strings.TrimSpace(req.Name)
	group := strings.TrimSpace(req.Group)
	kind := strings.ToLower(strings.TrimSpace(req.Kind))
	folder := strings.TrimSpace(req.Repo)
	slug := strings.TrimSpace(req.GitHubRepo)
	purposeEN := strings.TrimSpace(req.Purpose.EN)
	purposeRU := strings.TrimSpace(req.Purpose.RU)

	if !serviceIDPattern.MatchString(id) {
		return model.ServiceNode{}, &ValidationError{Msg: "id must be snake_case, start with a letter"}
	}
	if name == "" {
		return model.ServiceNode{}, &ValidationError{Msg: "name is required"}
	}
	if utf8.RuneCountInString(name) > 80 {
		return model.ServiceNode{}, &ValidationError{Msg: "name is too long"}
	}
	if kind == "" {
		kind = "worker"
	}
	if !serviceKindPattern.MatchString(kind) {
		return model.ServiceNode{}, &ValidationError{Msg: "invalid kind"}
	}
	if !githubSlugPattern.MatchString(slug) || slug == "." || slug == ".." {
		return model.ServiceNode{}, &ValidationError{Msg: "invalid GitHub repository name"}
	}
	if strings.Contains(folder, "/") || strings.Contains(folder, `\`) || !localFolderPattern.MatchString(folder) || folder == "." || folder == ".." {
		return model.ServiceNode{}, &ValidationError{Msg: "local folder must be a single directory name"}
	}
	if utf8.RuneCountInString(purposeEN) > 240 || utf8.RuneCountInString(purposeRU) > 240 {
		return model.ServiceNode{}, &ValidationError{Msg: "purpose is too long"}
	}

	groupOK := false
	for _, g := range profile.Groups {
		if g.ID == group {
			groupOK = true
			break
		}
	}
	if !groupOK {
		return model.ServiceNode{}, &ValidationError{Msg: "unknown group"}
	}

	for _, svc := range profile.Services {
		if svc.ID == id {
			return model.ServiceNode{}, &ValidationError{Msg: "a unit with this id already exists"}
		}
		if strings.EqualFold(svc.Repo, folder) {
			return model.ServiceNode{}, &ValidationError{Msg: "a unit with this folder already exists"}
		}
		if strings.EqualFold(svc.GitHubRepo, slug) {
			return model.ServiceNode{}, &ValidationError{Msg: "a unit with this GitHub repository already exists"}
		}
	}

	return model.ServiceNode{
		ID:         id,
		Name:       name,
		Group:      group,
		Kind:       kind,
		Repo:       folder,
		GitHubRepo: slug,
		Purpose:    model.LocalizedText{EN: purposeEN, RU: purposeRU},
	}, nil
}

func serviceDescription(node model.ServiceNode) string {
	if node.Purpose.EN != "" {
		return node.Purpose.EN
	}
	if node.Purpose.RU != "" {
		return node.Purpose.RU
	}
	return node.Name
}

func cloneDestinationExists(workspace, folder string) (bool, error) {
	dest := filepath.Join(workspace, folder)
	if rel, err := filepath.Rel(workspace, dest); err != nil || strings.HasPrefix(rel, "..") {
		return false, fmt.Errorf("invalid folder")
	}
	_, err := os.Stat(dest)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
