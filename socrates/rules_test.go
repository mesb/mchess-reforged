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

func TestCastlingKingside(t *testing.T) {
	b := board.NewBoard()
	e := &RuleEngine{
		Board: b,
		State: board.NewGameState(),
		Turn:  pieces.WHITE,
		Log:   &Log{},
	}

	kingPos := address.MakeAddr(0, 4) // e1
	rookPos := address.MakeAddr(0, 7) // h1
	b.SetPiece(kingPos, pieces.NewKing(pieces.WHITE))
	b.SetPiece(rookPos, pieces.NewRook(pieces.WHITE))

	if !e.MakeMove(kingPos, address.MakeAddr(0, 6), 0) {
		t.Fatal("Expected kingside castling to succeed")
	}

	if _, ok := b.PieceAt(address.MakeAddr(0, 6)).(*pieces.King); !ok {
		t.Fatalf("King not on g1 after castling")
	}
	if _, ok := b.PieceAt(address.MakeAddr(0, 5)).(*pieces.Rook); !ok {
		t.Fatalf("Rook not on f1 after castling")
	}
}

func TestCastlingBlockedByAttack(t *testing.T) {
	b := board.NewBoard()
	e := &RuleEngine{
		Board: b,
		State: board.NewGameState(),
		Turn:  pieces.WHITE,
		Log:   &Log{},
	}

	kingPos := address.MakeAddr(0, 4) // e1
	rookPos := address.MakeAddr(0, 7) // h1
	b.SetPiece(kingPos, pieces.NewKing(pieces.WHITE))
	b.SetPiece(rookPos, pieces.NewRook(pieces.WHITE))

	// Black rook attacking f1, which should invalidate castling.
	b.SetPiece(address.MakeAddr(7, 5), pieces.NewRook(pieces.BLACK)) // f8 -> attacks f1

	if e.MakeMove(kingPos, address.MakeAddr(0, 6), 0) {
		t.Fatal("Castling through attack should fail")
	}
}

func TestEnPassantCannotExposeKing(t *testing.T) {
	b := board.NewBoard()
	e := &RuleEngine{
		Board: b,
		State: board.NewGameState(),
		Turn:  pieces.WHITE,
		Log:   &Log{},
	}

	whiteKing := pieces.NewKing(pieces.WHITE)
	whitePawn := pieces.NewPawn(pieces.WHITE)
	blackPawn := pieces.NewPawn(pieces.BLACK)
	blackRook := pieces.NewRook(pieces.BLACK)

	b.SetPiece(address.MakeAddr(0, 4), whiteKing) // e1
	b.SetPiece(address.MakeAddr(4, 4), whitePawn) // e5
	b.SetPiece(address.MakeAddr(4, 3), blackPawn) // d5 (just double-moved)
	b.SetPiece(address.MakeAddr(7, 4), blackRook) // e8

	epTarget := address.MakeAddr(5, 3) // d6
	e.State.SetEnPassant(&epTarget)

	if e.MakeMove(address.MakeAddr(4, 4), epTarget, 0) {
		t.Fatal("En passant that exposes king to rook should be illegal")
	}
}

func TestUndoRestoresState(t *testing.T) {
	b := board.InitStandard()
	e := New(b)

	from := address.MakeAddr(1, 4) // e2
	to := address.MakeAddr(3, 4)   // e4

	if !e.MakeMove(from, to, 0) {
		t.Fatal("Initial pawn double move failed")
	}
	if e.State.GetEnPassant() == nil {
		t.Fatal("Expected en passant target after double move")
	}

	if !e.UndoMove() {
		t.Fatal("UndoMove failed")
	}

	if e.State.GetEnPassant() != nil {
		t.Fatal("En passant target not cleared after undo")
	}
	if e.State.CastlingRights != "KQkq" {
		t.Fatalf("Castling rights not restored, got %s", e.State.CastlingRights)
	}
	if e.State.FullmoveNumber != 1 {
		t.Fatalf("FullmoveNumber not restored, got %d", e.State.FullmoveNumber)
	}
	if e.State.Turn != pieces.WHITE || e.Turn != pieces.WHITE {
		t.Fatal("Turn not restored after undo")
	}
	if b.PieceAt(from) == nil || !b.IsEmpty(to) {
		t.Fatal("Board not restored after undo")
	}
}

func TestStalemateScenario(t *testing.T) {
	// Classic stalemate: Black king on h8, White king on g6, White queen on f7.
	fen := "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1"
	b, state, err := board.FromFEN(fen)
	if err != nil {
		t.Fatalf("Failed to parse FEN: %v", err)
	}

	engine := New(b)
	engine.State = state
	engine.Turn = pieces.BLACK
	engine.ResetHashHistory()

	moves := engine.GenerateLegalMoves()
	if len(moves) != 0 {
		t.Errorf("Expected 0 legal moves (stalemate), found %d", len(moves))
		for _, m := range moves {
			t.Logf("Legal move found: %v -> %v", m.From, m.To)
		}
	}

	if engine.IsInCheck(pieces.BLACK) {
		t.Fatal("Black incorrectly reported in check in stalemate position")
	}

	result := engine.Search(1)
	if result.Score != 0 {
		t.Fatalf("Expected draw evaluation (0) in stalemate, got %d", result.Score)
	}
}
