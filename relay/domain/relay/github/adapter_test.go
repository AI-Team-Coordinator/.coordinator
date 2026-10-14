package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"

	"relay/domain/adapter"
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
	got, err := Adapter{WebhookSecret: "s"}.Normalize(t.Context(), adapter.Delivery{
		Event: "pull_request_review_comment",
		Body:  body,
		Sig:   sign("s", body),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "review.commented" || got[0].GitHubLogin != "ek" || got[0].Branch != "feat/PROJ-3" {
		t.Fatalf("%+v", got)
	}
}

func TestBadSignature(t *testing.T) {
	_, err := Adapter{WebhookSecret: "s"}.Normalize(t.Context(), adapter.Delivery{
		Event: "pull_request_review_comment",
		Body:  []byte("{}"),
		Sig:   "sha256=00",
	})
	if !errors.Is(err, ErrBadSignature) {
		t.Fatalf("err=%v", err)
	}
}
