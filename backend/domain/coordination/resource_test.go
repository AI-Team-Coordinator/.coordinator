package coordination

import "testing"

func TestParseServiceAndPath(t *testing.T) {
	r := Parse("LLM/stt")
	if r.Key != "llm" || r.Path != "stt" || r.Kind != KindProduct || r.Label != "LLM/stt" {
		t.Fatalf("got %+v", r)
	}
}

func TestParseWorkspaceDot(t *testing.T) {
	r := Parse(".cursor")
	if r.Key != "cursor" || r.Kind != KindWorkspace {
		t.Fatalf("got %+v", r)
	}
}

func TestParseBus(t *testing.T) {
	r := Parse("Common")
	if r.Kind != KindBus || r.Key != "common" {
		t.Fatalf("got %+v", r)
	}
}

func TestRelatesWholeServiceCoversPath(t *testing.T) {
	a := Parse("billing")
	b := Parse("billing/invoices")
	if !a.Relates(b) || !b.Relates(a) {
		t.Fatal("expected prefix overlap")
	}
}

func TestRelatesDifferentServices(t *testing.T) {
	if Parse("billing").Relates(Parse("invoices")) {
		t.Fatal("unrelated keys")
	}
}

func TestRelatesNestedPaths(t *testing.T) {
	a := Parse("payments/invoices")
	b := Parse("payments/invoices/vat")
	if !a.Relates(b) {
		t.Fatal("expected nested path overlap")
	}
	if Parse("payments/invoices").Relates(Parse("payments/payouts")) {
		t.Fatal("sibling paths")
	}
}
