package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"relay/domain/connector"
	"relay/model"
)

var ErrNotReady = errors.New("github read is not implemented")

// Adapter is the stock GitHub connector: review and CI webhooks.
type Adapter struct {
	WebhookSecret string
	Secret        func(context.Context) string
}

func (a Adapter) webhookSecret(ctx context.Context) string {
	if a.Secret != nil {
		if s := strings.TrimSpace(a.Secret(ctx)); s != "" {
			return s
		}
	}
	return strings.TrimSpace(a.WebhookSecret)
}

func (a Adapter) Descriptor() connector.Descriptor {
	enabled := strings.TrimSpace(a.WebhookSecret) != ""
	return connector.Descriptor{
		ID:      SourceID,
		Name:    "GitHub",
		Detail:  "Pull requests, reviews, and Actions (read-only).",
		Enabled: enabled,
	}
}

func (a Adapter) Read(context.Context, string) (model.ExternalEvent, error) {
	return model.ExternalEvent{}, ErrNotReady
}

func (a Adapter) Normalize(ctx context.Context, d connector.Delivery) ([]model.ExternalEvent, error) {
	if err := verify(a.webhookSecret(ctx), d.Body, d.Sig); err != nil {
		return nil, err
	}
	switch d.Event {
	case "pull_request_review_comment":
		ev, err := reviewComment(d.Body)
		if err != nil {
			return nil, err
		}
		return []model.ExternalEvent{ev}, nil
	case "pull_request_review":
		ev, err := review(d.Body)
		if err != nil {
			return nil, err
		}
		return []model.ExternalEvent{ev}, nil
	case "workflow_run":
		ev, err := workflow(d.Body)
		if err != nil {
			return nil, err
		}
		if ev.Kind == "" {
			return nil, connector.ErrIgnored
		}
		return []model.ExternalEvent{ev}, nil
	default:
		return nil, connector.ErrIgnored
	}
}

func verify(secret string, body []byte, header string) error {
	if secret == "" {
		return connector.ErrBadSignature
	}
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return connector.ErrBadSignature
	}
	got, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return connector.ErrBadSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	if !hmac.Equal(mac.Sum(nil), got) {
		return connector.ErrBadSignature
	}
	return nil
}

func classForKind(kind string) string {
	switch kind {
	case "review.changes_requested", "ci.failed":
		return model.ClassWarning
	default:
		return model.ClassInfo
	}
}

func githubData(repo, branch, prAuthor string, extra map[string]any) map[string]any {
	out := map[string]any{
		"repo":   repo,
		"branch": branch,
	}
	if prAuthor != "" {
		out["pr_author_login"] = prAuthor
	}
	for k, v := range extra {
		if v != nil && v != "" {
			out[k] = v
		}
	}
	return out
}

func reviewComment(body []byte) (model.ExternalEvent, error) {
	var p struct {
		Action  string `json:"action"`
		Comment struct {
			ID      int64  `json:"id"`
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
			Path    string `json:"path"`
			User    struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"comment"`
		PullRequest struct {
			HTMLURL string `json:"html_url"`
			Head    struct {
				Ref string `json:"ref"`
			} `json:"head"`
			User struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"pull_request"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return model.ExternalEvent{}, err
	}
	if p.Action != "" && p.Action != "created" {
		return model.ExternalEvent{}, connector.ErrIgnored
	}
	summary := oneLine(p.Comment.Body, 240)
	title := p.Comment.Path
	if title == "" {
		title = "review comment"
	}
	kind := "review.commented"
	return model.ExternalEvent{
		ID:         fmt.Sprintf("github:review_comment:%d", p.Comment.ID),
		Source:     SourceID,
		ExternalID: fmt.Sprintf("%d", p.Comment.ID),
		Class:      classForKind(kind),
		Kind:       kind,
		Title:      title,
		Summary:    summary,
		URL:        firstNonEmpty(p.Comment.HTMLURL, p.PullRequest.HTMLURL),
		Tags:       []string{"github", "review"},
		Data: githubData(
			p.Repository.FullName,
			p.PullRequest.Head.Ref,
			p.PullRequest.User.Login,
			map[string]any{
				"path":            p.Comment.Path,
				"commenter_login": p.Comment.User.Login,
			},
		),
	}, nil
}

func review(body []byte) (model.ExternalEvent, error) {
	var p struct {
		Action string `json:"action"`
		Review struct {
			ID      int64  `json:"id"`
			State   string `json:"state"`
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
		} `json:"review"`
		PullRequest struct {
			HTMLURL string `json:"html_url"`
			Head    struct {
				Ref string `json:"ref"`
			} `json:"head"`
			User struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"pull_request"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return model.ExternalEvent{}, err
	}
	kind := ""
	switch p.Action {
	case "submitted":
		switch strings.ToLower(p.Review.State) {
		case "changes_requested":
			kind = "review.changes_requested"
		case "approved":
			kind = "review.approved"
		default:
			kind = "review.submitted"
		}
	default:
		return model.ExternalEvent{}, connector.ErrIgnored
	}
	return model.ExternalEvent{
		ID:         fmt.Sprintf("github:review:%d", p.Review.ID),
		Source:     SourceID,
		ExternalID: fmt.Sprintf("%d", p.Review.ID),
		Class:      classForKind(kind),
		Kind:       kind,
		Title:      p.Review.State,
		Summary:    oneLine(p.Review.Body, 240),
		URL:        firstNonEmpty(p.Review.HTMLURL, p.PullRequest.HTMLURL),
		Tags:       []string{"github", "review"},
		Data:       githubData(p.Repository.FullName, p.PullRequest.Head.Ref, p.PullRequest.User.Login, nil),
	}, nil
}

func workflow(body []byte) (model.ExternalEvent, error) {
	var p struct {
		Action      string `json:"action"`
		WorkflowRun struct {
			ID         int64  `json:"id"`
			Name       string `json:"name"`
			Conclusion string `json:"conclusion"`
			HTMLURL    string `json:"html_url"`
			HeadBranch string `json:"head_branch"`
			HeadSHA    string `json:"head_sha"`
		} `json:"workflow_run"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return model.ExternalEvent{}, err
	}
	if p.Action != "completed" {
		return model.ExternalEvent{}, connector.ErrIgnored
	}
	kind := ""
	switch strings.ToLower(p.WorkflowRun.Conclusion) {
	case "failure", "timed_out", "startup_failure":
		kind = "ci.failed"
	case "success":
		kind = "ci.passed"
	default:
		return model.ExternalEvent{}, connector.ErrIgnored
	}
	return model.ExternalEvent{
		ID:         fmt.Sprintf("github:workflow_run:%d", p.WorkflowRun.ID),
		Source:     SourceID,
		ExternalID: fmt.Sprintf("%d", p.WorkflowRun.ID),
		Class:      classForKind(kind),
		Kind:       kind,
		Title:      p.WorkflowRun.Name,
		Summary:    p.WorkflowRun.HeadSHA,
		URL:        p.WorkflowRun.HTMLURL,
		Tags:       []string{"github", "ci"},
		Data: githubData(
			p.Repository.FullName,
			p.WorkflowRun.HeadBranch,
			"",
			map[string]any{"sha": p.WorkflowRun.HeadSHA, "conclusion": p.WorkflowRun.Conclusion},
		),
	}, nil
}

func oneLine(s string, max int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
