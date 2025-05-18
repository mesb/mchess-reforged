// --- pieces/knight.go ---

package pieces

import "github.com/mesb/mchess/address"

type Knight struct {
	color int
}

func NewKnight(color int) *Knight {
	return &Knight{color: color}
}

func (k *Knight) Color() int {
	return k.color
}

func (k *Knight) String() string {
	if k.color == WHITE {
		return "♘"
	}
	return "♞"
}

func (k *Knight) ValidMoves(from address.Addr, board BoardView) []address.Addr {
	deltas := [][2]int{
		{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2},
		{1, -2}, {1, 2}, {2, -1}, {2, 1},
	}
	moves := []address.Addr{}
	for _, d := range deltas {
		if to, ok := from.Shift(d[0], d[1]); ok {
			if board.IsEmpty(to) || board.PieceAt(to).Color() != k.color {
				moves = append(moves, to)
			}
		}
	}
	return moves
}
