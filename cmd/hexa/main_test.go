package main

import "testing"

func TestInitBoardLayout(t *testing.T) {
	b := initBoard()
	if len(b) != total {
		t.Fatalf("expected %d cells, got %d", total, len(b))
	}
	if b[0] != "♜" || b[total-1] != "♖" {
		t.Fatalf("unexpected corner pieces: %s ... %s", b[0], b[total-1])
	}
}
