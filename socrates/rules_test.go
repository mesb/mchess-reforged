// --- socrates/rules_test.go ---

package socrates

import (
	"testing"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

func TestFoolsmate(t *testing.T) {
	b := board.InitStandard()
	e := New(b)

	moves := []string{"f2f3", "e7e5", "g2g4", "d8h4"}

	for _, m := range moves {
		from := parseSquare(m[:2])
		to := parseSquare(m[2:])
		if !e.MakeMove(*from, *to, 0) {
			t.Fatalf("Move %s failed", m)
		}
	}

	if !e.IsCheckmate() {
		t.Error("Expected checkmate")
	}
}

func TestPromotion(t *testing.T) {
	b := board.NewBoard()
	e := New(b)

	// Setup pawn about to promote
	pawn := pieces.NewPawn(pieces.WHITE)
	start := address.MakeAddr(6, 0) // a7
	b.SetPiece(start, pawn)

	end := address.MakeAddr(7, 0) // a8

	// Promote to Knight
	if !e.MakeMove(start, end, 'n') {
		t.Fatal("Promotion move failed")
	}

	promoted := b.PieceAt(end)
	if _, ok := promoted.(*pieces.Knight); !ok {
		t.Errorf("Expected Knight, got %v", promoted)
	}
}
