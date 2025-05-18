// --- pieces/pawn.go ---

package pieces

import "github.com/mesb/mchess/address"

type Pawn struct {
	color int
}

func NewPawn(color int) *Pawn {
	return &Pawn{color: color}
}

func (p *Pawn) Color() int {
	return p.color
}

func (p *Pawn) String() string {
	if p.color == WHITE {
		return "♙"
	}
	return "♟"
}

func (p *Pawn) ValidMoves(from address.Addr, b BoardView) []address.Addr {
	var moves []address.Addr
	dir := 1
	startRank := 1
	if p.color == BLACK {
		dir = -1
		startRank = 6
	}

	// Forward 1 square
	if m1, ok := from.Shift(dir, 0); ok && b.IsEmpty(m1) {
		moves = append(moves, m1)

		// Forward 2 squares from initial rank
		if int(from.Rank) == startRank {
			if m2, ok := from.Shift(2*dir, 0); ok && b.IsEmpty(m2) {
				moves = append(moves, m2)
			}
		}
	}

	// Diagonal captures
	for _, df := range []int{-1, 1} {
		if diag, ok := from.Shift(dir, df); ok {
			if target := b.PieceAt(diag); target != nil && target.Color() != p.color {
				moves = append(moves, diag)
			}
		}
	}

	return moves
}
