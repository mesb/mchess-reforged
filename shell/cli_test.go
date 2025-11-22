package shell

import "testing"

func TestNormalizeInput(t *testing.T) {
	if got := normalizeInput("e2e4"); got != "m e2e4" {
		t.Fatalf("expected move normalization, got %s", got)
	}
	if got := normalizeInput("m e2e4"); got != "m e2e4" {
		t.Fatalf("expected passthrough for explicit move, got %s", got)
	}
	if got := normalizeInput("noop"); got != "noop" {
		t.Fatalf("expected noop passthrough, got %s", got)
	}
}

func TestColorName(t *testing.T) {
	if colorName(0) != "White" || colorName(1) != "Black" {
		t.Fatalf("unexpected color names")
	}
}
