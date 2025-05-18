// --- pieces/queen.go ---

package pieces

import "github.com/mesb/mchess/address"

type Queen struct {
	color int
}

func NewQueen(color int) *Queen {
	return &Queen{color: color}
}

func (q *Queen) Color() int {
	return q.color
}

func (q *Queen) String() string {
	if q.color == WHITE {
		return "♕"
	}
	return "♛"
}

func (q *Queen) ValidMoves(from address.Addr, board BoardView) []address.Addr {
	dirs := [][2]int{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {-1, -1}, {1, -1}, {-1, 1},
	}
	var moves []address.Addr
	for _, dir := range dirs {
		for step := 1; step < 8; step++ {
			if to, ok := from.Shift(dir[0]*step, dir[1]*step); ok {
				if board.IsEmpty(to) {
					moves = append(moves, to)
				} else {
					if board.PieceAt(to).Color() != q.color {
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
