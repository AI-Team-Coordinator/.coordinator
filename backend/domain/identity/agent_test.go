package identity

import "testing"

func TestCursorChatsDedupes(t *testing.T) {
	got := CursorChats("EK", []string{" a ", "", "a", "b"})
	if len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" {
		t.Fatalf("got %+v", got)
	}
	if got[0].Person != "EK" || got[0].Kind != KindCursorChat {
		t.Fatalf("agent %+v", got[0])
	}
}

func TestWithChatTitles(t *testing.T) {
	got := WithChatTitles(CursorChats("EK", []string{"a"}), map[string]string{"a": " Domain model "})
	if len(got) != 1 || got[0].Title != "Domain model" {
		t.Fatalf("got %+v", got)
	}
}
