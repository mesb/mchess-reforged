package pieces

import (
	"testing"

	"github.com/mesb/mchess/address"
)

func TestCloneWithColor(t *testing.T) {
	q := NewQueen(WHITE)
	clone := CloneWithColor(q, BLACK)
	if clone == nil || clone.Color() != BLACK {
		t.Fatalf("clone color mismatch: %#v", clone)
	}
	if clone.String() != "♛" {
		t.Fatalf("expected black queen symbol, got %s", clone.String())
	}
}

func TestFromChar(t *testing.T) {
	if _, ok := FromChar('n', WHITE).(*Knight); !ok {
		t.Fatalf("expected knight from 'n'")
	}
	if _, ok := FromChar('B', BLACK).(*Bishop); !ok {
		t.Fatalf("expected bishop from 'B'")
	}
	if _, ok := FromChar('x', WHITE).(*Queen); !ok {
		t.Fatalf("default should produce queen")
	}
}

func TestKnightMoves(t *testing.T) {
	b := NewBoardStub()
	k := NewKnight(WHITE)
	from := address.MakeAddr(4, 4) // e5
	moves := k.ValidMoves(from, b, nil)
	if len(moves) != 8 {
		t.Fatalf("expected 8 knight moves from center, got %d", len(moves))
	}
}

// NewBoardStub provides an empty board implementing BoardView.
func NewBoardStub() BoardView {
	return &stubBoard{}
}

type stubBoard struct{}

func (s *stubBoard) IsEmpty(address.Addr) bool { return true }
func (s *stubBoard) PieceAt(address.Addr) Piece {
	return nil
}
