package main

import (
	"testing"

	"github.com/mesb/mchess/shell"
)

// Smoke test to ensure the CLI can bootstrap a session without interactive IO.
func TestSessionBootstrap(t *testing.T) {
	session := shell.NewSession(nil)
	if session == nil || session.Engine == nil {
		t.Fatal("expected non-nil session and engine")
	}
	if session.Engine.Turn != 0 {
		t.Fatalf("expected White to start, got %d", session.Engine.Turn)
	}
}
