// --- pieces/king.go ---

package pieces

import "github.com/mesb/mchess/address"

type King struct {
	color int
}

func NewKing(color int) *King {
	return &King{color: color}
}

func (k *King) Color() int {
	return k.color
}

func (k *King) String() string {
	if k.color == WHITE {
		return "♔"
	}
	return "♚"
}

func (k *King) ValidMoves(from address.Addr, board BoardView, state GameStateView) []address.Addr {
	dirs := [][2]int{
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {-1, -1}, {1, -1}, {-1, 1},
	}
	var moves []address.Addr
	for _, dir := range dirs {
		if to, ok := from.Shift(dir[0], dir[1]); ok {
			if board.IsEmpty(to) || board.PieceAt(to).Color() != k.color {
				moves = append(moves, to)
			}
		}
	}
	return moves
}
