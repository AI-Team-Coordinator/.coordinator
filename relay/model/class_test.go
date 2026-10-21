package model

import "testing"

func TestValidClass(t *testing.T) {
	if !ValidClass(ClassInfo) || !ValidClass(ClassWarning) || ValidClass("alert") {
		t.Fatal("class validation")
	}
}

func TestExternalEventSeal(t *testing.T) {
	ev := ExternalEvent{Kind: "ci.failed"}
	ev.Seal("2026-10-20T12:00:00Z")
	if ev.SchemaVersion != ExternalEventSchemaVersion || ev.ReceivedAt == "" || ev.Class != ClassInfo {
		t.Fatalf("%+v", ev)
	}
}
