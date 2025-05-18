// --- pieces/utils.go ---

package pieces

// CloneWithColor creates a new piece of the same kind but with a new color.
// This is used during board initialization or pawn promotion logic.
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
		_ = t // prevent unused variable warning
		return nil
	}
}
