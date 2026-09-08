package coordinator

import (
	"net/url"
	"strings"

	"coordinator/model"
)

func githubHTMLHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return "github.com"
	}
	return host
}

func githubOrgHTMLURL(binding model.GitHubBinding) string {
	org := strings.TrimSpace(binding.Org)
	if org == "" {
		return ""
	}
	return "https://" + githubHTMLHost(binding.Host) + "/" + org
}

func githubRepoHTMLURL(binding model.GitHubBinding, slug string) string {
	orgURL := githubOrgHTMLURL(binding)
	slug = strings.TrimSpace(slug)
	if orgURL == "" || slug == "" {
		return ""
	}
	return orgURL + "/" + slug
}

func parseGitHubRemote(raw string) (org, slug string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	raw = strings.TrimSuffix(raw, ".git")

	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", ""
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 2 {
			return "", ""
		}
		return parts[0], strings.Join(parts[1:], "/")
	}

	if at := strings.Index(raw, ":"); at >= 0 && !strings.Contains(raw[:at], "/") {
		path := strings.TrimPrefix(raw[at+1:], "/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			return "", ""
		}
		return parts[0], strings.Join(parts[1:], "/")
	}
	return "", ""
}
