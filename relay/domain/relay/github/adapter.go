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

	"relay/domain/adapter"
	"relay/model"
)

const originID = "github"

var (
	ErrBadSignature = errors.New("github webhook signature mismatch")
	ErrIgnored      = errors.New("github event ignored")
	ErrNotReady     = errors.New("github read is not implemented")
)

// Adapter is the first Relay adapter: GitHub review and CI webhooks.
type Adapter struct {
	WebhookSecret string
}

func (a Adapter) ID() string { return originID }

func (a Adapter) Read(context.Context, string) (model.ExternalEvent, error) {
	return model.ExternalEvent{}, ErrNotReady
}

func (a Adapter) Normalize(_ context.Context, d adapter.Delivery) ([]model.ExternalEvent, error) {
	if err := verify(a.WebhookSecret, d.Body, d.Sig); err != nil {
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
			return nil, ErrIgnored
		}
		return []model.ExternalEvent{ev}, nil
	default:
		return nil, ErrIgnored
	}
}

func verify(secret string, body []byte, header string) error {
	if secret == "" {
		return ErrBadSignature
	}
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return ErrBadSignature
	}
	got, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return ErrBadSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	if !hmac.Equal(mac.Sum(nil), got) {
		return ErrBadSignature
	}
	return nil
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
		return model.ExternalEvent{}, ErrIgnored
	}
	summary := oneLine(p.Comment.Body, 240)
	title := p.Comment.Path
	if title == "" {
		title = "review comment"
	}
	return model.ExternalEvent{
		ID:          fmt.Sprintf("github:review_comment:%d", p.Comment.ID),
		Source:      originID,
		OriginID:    originID,
		ExternalID:  fmt.Sprintf("%d", p.Comment.ID),
		Kind:        "review.commented",
		Title:       title,
		Summary:     summary,
		URL:         firstNonEmpty(p.Comment.HTMLURL, p.PullRequest.HTMLURL),
		Repo:        p.Repository.FullName,
		Branch:      p.PullRequest.Head.Ref,
		GitHubLogin: p.PullRequest.User.Login,
		Tags:        []string{"github", "review"},
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
		return model.ExternalEvent{}, ErrIgnored
	}
	return model.ExternalEvent{
		ID:          fmt.Sprintf("github:review:%d", p.Review.ID),
		Source:      originID,
		OriginID:    originID,
		ExternalID:  fmt.Sprintf("%d", p.Review.ID),
		Kind:        kind,
		Title:       p.Review.State,
		Summary:     oneLine(p.Review.Body, 240),
		URL:         firstNonEmpty(p.Review.HTMLURL, p.PullRequest.HTMLURL),
		Repo:        p.Repository.FullName,
		Branch:      p.PullRequest.Head.Ref,
		GitHubLogin: p.PullRequest.User.Login,
		Tags:        []string{"github", "review"},
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
		return model.ExternalEvent{}, ErrIgnored
	}
	kind := ""
	switch strings.ToLower(p.WorkflowRun.Conclusion) {
	case "failure", "timed_out", "startup_failure":
		kind = "ci.failed"
	case "success":
		kind = "ci.passed"
	default:
		return model.ExternalEvent{}, ErrIgnored
	}
	return model.ExternalEvent{
		ID:         fmt.Sprintf("github:workflow_run:%d", p.WorkflowRun.ID),
		Source:     originID,
		OriginID:   originID,
		ExternalID: fmt.Sprintf("%d", p.WorkflowRun.ID),
		Kind:       kind,
		Title:      p.WorkflowRun.Name,
		Summary:    p.WorkflowRun.HeadSHA,
		URL:        p.WorkflowRun.HTMLURL,
		Repo:       p.Repository.FullName,
		Branch:     p.WorkflowRun.HeadBranch,
		Tags:       []string{"github", "ci"},
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
