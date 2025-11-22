// --- pieces/bishop.go ---
package pieces

import "github.com/mesb/mchess/address"

type Bishop struct {
	color int
}

func NewBishop(color int) *Bishop {
	return &Bishop{color: color}
}

func (b *Bishop) Color() int {
	return b.color
}

func (b *Bishop) String() string {
	if b.color == WHITE {
		return "♗"
	}
	return "♝"
}

func (b *Bishop) ValidMoves(from address.Addr, board BoardView, state GameStateView) []address.Addr {
	dirs := [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	var moves []address.Addr
	for _, dir := range dirs {
		for step := 1; step < 8; step++ {
			if to, ok := from.Shift(dir[0]*step, dir[1]*step); ok {
				if board.IsEmpty(to) {
					moves = append(moves, to)
				} else {
					if board.PieceAt(to).Color() != b.color {
						moves = append(moves, to)
					}
					break
				}
			} else {
				break
			}
		}
	}
	return moves
}
