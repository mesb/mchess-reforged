// --- pieces/rook.go ---

package pieces

import "github.com/mesb/mchess/address"

type Rook struct {
	color int
}

func NewRook(color int) *Rook {
	return &Rook{color: color}
}

func (r *Rook) Color() int {
	return r.color
}

func (r *Rook) String() string {
	if r.color == WHITE {
		return "♖"
	}
	return "♜"
}

func (r *Rook) ValidMoves(from address.Addr, board BoardView, state GameStateView) []address.Addr {
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	var moves []address.Addr
	for _, dir := range dirs {
		for step := 1; step < 8; step++ {
			if to, ok := from.Shift(dir[0]*step, dir[1]*step); ok {
				if board.IsEmpty(to) {
					moves = append(moves, to)
				} else {
					if board.PieceAt(to).Color() != r.color {
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
