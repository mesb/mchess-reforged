// --- pieces/utils.go ---

package pieces

// CloneWithColor creates a new piece of the same kind but with a new color.
func CloneWithColor(p Piece, color int) Piece {
	switch t := p.(type) {
	case *Rook:
		return NewRook(color)
	case *Knight:
		return NewKnight(color)
	case *Bishop:
		return NewBishop(color)
	case *Queen:
		return NewQueen(color)
	case *King:
		return NewKing(color)
	case *Pawn:
		return NewPawn(color)
	default:
		_ = t
		return nil
	}
}

// FromChar returns a new piece instance based on a character code (q, r, b, n).
func FromChar(c rune, color int) Piece {
	switch c {
	case 'r', 'R':
		return NewRook(color)
	case 'b', 'B':
		return NewBishop(color)
	case 'n', 'N':
		return NewKnight(color)
	default:
		return NewQueen(color)
	}
}
