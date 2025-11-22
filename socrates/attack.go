package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

// isSquareAttacked returns true if square a is attacked by given color.
func (r *RuleEngine) isSquareAttacked(a address.Addr, byColor int) bool {
	// Knights
	knightDeltas := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	for _, d := range knightDeltas {
		if pos, ok := a.Shift(d[0], d[1]); ok {
			p := r.Board.PieceAt(pos)
			if p != nil && p.Color() == byColor {
				if _, isKnight := p.(*pieces.Knight); isKnight {
					return true
				}
			}
		}
	}

	// Pawns
	pawnDir := 1
	if byColor == pieces.BLACK {
		pawnDir = -1
	}
	for _, dx := range []int{-1, 1} {
		if pos, ok := a.Shift(pawnDir, dx); ok {
			p := r.Board.PieceAt(pos)
			if p != nil && p.Color() == byColor {
				if _, isPawn := p.(*pieces.Pawn); isPawn {
					return true
				}
			}
		}
	}

	// Enemy king adjacency
	kingDeltas := [][2]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
	for _, d := range kingDeltas {
		if pos, ok := a.Shift(d[0], d[1]); ok {
			p := r.Board.PieceAt(pos)
			if p != nil && p.Color() == byColor {
				if _, isKing := p.(*pieces.King); isKing {
					return true
				}
			}
		}
	}

	// Sliders
	if scanDirs(r.Board, a, [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}, byColor, true, false) {
		return true
	}
	if scanDirs(r.Board, a, [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}, byColor, false, true) {
		return true
	}
	return false
}

// reuse scanDirs
func scanDirs(b *board.Board, start address.Addr, dirs [][2]int, enemyColor int, checkRook, checkBishop bool) bool {
	for _, d := range dirs {
		for i := 1; i < 8; i++ {
			pos, ok := start.Shift(d[0]*i, d[1]*i)
			if !ok {
				break
			}
			p := b.PieceAt(pos)
			if p != nil {
				if p.Color() == enemyColor {
					switch p.(type) {
					case *pieces.Queen:
						return true
					case *pieces.Rook:
						if checkRook {
							return true
						}
					case *pieces.Bishop:
						if checkBishop {
							return true
						}
					}
				}
				break
			}
		}
	}
	return false
}
