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

	// Castling: add two-square jumps if castling rights and empty path.
	if state != nil {
		rights := state.GetCastlingRights()
		rank := 0
		if k.color == BLACK {
			rank = 7
		}

		// Kingside
		if hasRight(rights, k.color, true) {
			moves = append(moves, address.MakeAddr(address.Rank(rank), address.File(6)))
		}
		// Queenside
		if hasRight(rights, k.color, false) {
			moves = append(moves, address.MakeAddr(address.Rank(rank), address.File(2)))
		}
	}

	return moves
}

func hasRight(rights string, color int, kingSide bool) bool {
	if color == WHITE {
		if kingSide {
			return contains(rights, "K")
		}
		return contains(rights, "Q")
	}
	if kingSide {
		return contains(rights, "k")
	}
	return contains(rights, "q")
}

func contains(s, sub string) bool {
	for _, r := range s {
		if string(r) == sub {
			return true
		}
	}
	return false
}
