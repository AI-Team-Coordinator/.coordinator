package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"relay/domain/connector"
	"relay/model"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestReviewComment(t *testing.T) {
	body := []byte(`{
		"action":"created",
		"comment":{"id":9,"body":"rename this","html_url":"https://github.com/acme/app/pull/3#discussion_r9","path":"main.go","user":{"login":"reviewer"}},
		"pull_request":{"html_url":"https://github.com/acme/app/pull/3","head":{"ref":"feat/PROJ-3"},"user":{"login":"ek"}},
		"repository":{"full_name":"acme/app"}
	}`)
	got, err := Adapter{WebhookSecret: "s"}.Normalize(t.Context(), connector.Delivery{
		Event: "pull_request_review_comment",
		Body:  body,
		Sig:   sign("s", body),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	ev := got[0]
	if ev.Kind != "review.commented" || ev.Class != model.ClassInfo || ev.Source != "github" {
		t.Fatalf("%+v", ev)
	}
	if ev.Data["branch"] != "feat/PROJ-3" || ev.Data["pr_author_login"] != "ek" {
		t.Fatalf("data=%v", ev.Data)
	}
}

func TestReviewChangesRequestedIsWarning(t *testing.T) {
	body := []byte(`{
		"action":"submitted",
		"review":{"id":1,"state":"changes_requested","body":"fix","html_url":"https://github.com/acme/app/pull/1#pullrequestreview-1"},
		"pull_request":{"html_url":"https://github.com/acme/app/pull/1","head":{"ref":"feat/x"},"user":{"login":"ek"}},
		"repository":{"full_name":"acme/app"}
	}`)
	got, err := Adapter{WebhookSecret: "s"}.Normalize(t.Context(), connector.Delivery{
		Event: "pull_request_review",
		Body:  body,
		Sig:   sign("s", body),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Class != model.ClassWarning || got[0].Kind != "review.changes_requested" {
		t.Fatalf("%+v", got[0])
	}
}

func TestBadSignature(t *testing.T) {
	_, err := Adapter{WebhookSecret: "s"}.Normalize(t.Context(), connector.Delivery{
		Event: "pull_request_review_comment",
		Body:  []byte("{}"),
		Sig:   "sha256=00",
	})
	if !errors.Is(err, connector.ErrBadSignature) {
		t.Fatalf("err=%v", err)
	}
}
