package board

import (
	"testing"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

func TestToFENInitialPosition(t *testing.T) {
	b := InitStandard()
	state := NewGameState()

	fen := b.ToFEN(state)
	want := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if fen != want {
		t.Fatalf("initial FEN mismatch:\n got: %s\nwant: %s", fen, want)
	}
}

func TestToFENWithEnPassantAndCounts(t *testing.T) {
	b := NewBoard()
	state := &GameState{
		Turn:           pieces.BLACK,
		CastlingRights: "Kq",
		HalfmoveClock:  3,
		FullmoveNumber: 7,
	}

	whitePawn := pieces.NewPawn(pieces.WHITE)
	blackKing := pieces.NewKing(pieces.BLACK)
	whiteKing := pieces.NewKing(pieces.WHITE)

	b.SetPiece(address.MakeAddr(3, 4), whitePawn) // e4 with ep target on e5
	b.SetPiece(address.MakeAddr(0, 4), whiteKing) // e1
	b.SetPiece(address.MakeAddr(7, 4), blackKing) // e8

	ep := address.MakeAddr(4, 4) // e5
	state.EnPassant = &ep

	fen := b.ToFEN(state)
	want := "4k3/8/8/8/4P3/8/8/4K3 b Kq e5 3 7"
	if fen != want {
		t.Fatalf("FEN with state mismatch:\n got: %s\nwant: %s", fen, want)
	}
}

func TestFromFENRoundTrip(t *testing.T) {
	original := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	b, state, err := FromFEN(original)
	if err != nil {
		t.Fatalf("FromFEN error: %v", err)
	}
	roundTrip := b.ToFEN(state)
	if roundTrip != original {
		t.Fatalf("round trip mismatch:\n got: %s\nwant: %s", roundTrip, original)
	}
}
