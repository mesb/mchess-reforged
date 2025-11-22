package socrates

import (
	"testing"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

func TestEvaluateInitialPositionIsZero(t *testing.T) {
	b := board.InitStandard()
	if got := Evaluate(b); got != 0 {
		t.Fatalf("expected initial eval 0, got %d", got)
	}
}

func TestEvaluateMaterialLead(t *testing.T) {
	b := board.NewBoard()
	queen := pieces.NewQueen(pieces.WHITE)
	b.SetPiece(address.MakeAddr(3, 3), queen) // d4

	score := Evaluate(b)
	if score <= ValueQueen {
		t.Fatalf("expected strong white advantage, got %d", score)
	}
}

func TestEvaluateSymmetry(t *testing.T) {
	b := board.NewBoard()
	b.SetPiece(address.MakeAddr(1, 1), pieces.NewKnight(pieces.WHITE)) // b2
	b.SetPiece(address.MakeAddr(6, 6), pieces.NewKnight(pieces.BLACK)) // g7 (mirrored)

	if got := Evaluate(b); got != 0 {
		t.Fatalf("expected symmetrical knights to cancel, got %d", got)
	}
}
