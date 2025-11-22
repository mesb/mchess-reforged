// --- socrates/rules.go ---

package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

type RuleEngine struct {
	Board *board.Board
	State *board.GameState
	Turn  int
	Log   *Log
}

func New(b *board.Board) *RuleEngine {
	return &RuleEngine{
		Board: b,
		State: board.NewGameState(),
		Turn:  pieces.WHITE,
		Log:   &Log{},
	}
}

func (r *RuleEngine) MakeMove(from, to address.Addr) bool {
	if !r.IsLegalMove(from, to) {
		return false
	}

	moving := r.Board.PieceAt(from)
	target := r.Board.PieceAt(to)

	// Flag for 50-move rule
	_, isPawn := moving.(*pieces.Pawn)
	isCapture := target != nil

	// --- Logic: En Passant Capture ---
	if isPawn {
		if ep := r.State.GetEnPassant(); ep != nil && to.Equals(*ep) {
			captureRankDir := -1
			if moving.Color() == pieces.BLACK {
				captureRankDir = 1
			}
			if victimPos, ok := to.Shift(captureRankDir, 0); ok {
				target = r.Board.PieceAt(victimPos) // Capture the pawn behind
				r.Board.Clear(victimPos)
				isCapture = true
			}
		}
	}

	// --- Logic: Castling Execution ---
	if _, isKing := moving.(*pieces.King); isKing {
		df := int(to.File) - int(from.File)
		if df == 2 || df == -2 {
			rank := from.Rank
			var rookFrom, rookTo address.Addr

			if df == 2 { // Kingside
				rookFrom = address.MakeAddr(rank, 7)
				rookTo = address.MakeAddr(rank, 5)
			} else { // Queenside
				rookFrom = address.MakeAddr(rank, 0)
				rookTo = address.MakeAddr(rank, 3)
			}

			rook := r.Board.PieceAt(rookFrom)
			r.Board.SetPiece(rookTo, rook)
			r.Board.Clear(rookFrom)
		}
	}

	// Record Move
	if r.Log != nil {
		r.Log.Record(from, to, moving, target)
	}

	// Apply Main Move
	r.Board.SetPiece(to, moving)
	r.Board.Clear(from)

	// --- State Updates ---
	r.updateEnPassantState(moving, from, to)
	r.updateCastlingRights(moving, from)

	// Update 50-move clock
	r.State.IncrementClock(isPawn, isCapture)

	// Auto-Queen (Simplified for CLI)
	if isPawn {
		rank := to.Rank
		if (moving.Color() == pieces.WHITE && rank == 7) || (moving.Color() == pieces.BLACK && rank == 0) {
			r.Board.SetPiece(to, pieces.NewQueen(moving.Color()))
		}
	}

	r.Turn = 1 - r.Turn
	r.State.Turn = r.Turn

	// Increment full move number if it was Black's turn
	if moving.Color() == pieces.BLACK {
		r.State.FullmoveNumber++
	}

	return true
}

func (r *RuleEngine) updateEnPassantState(p pieces.Piece, from, to address.Addr) {
	r.State.SetEnPassant(nil)
	if _, ok := p.(*pieces.Pawn); ok {
		dr, _ := address.Delta(from, to)
		if dr == 2 || dr == -2 {
			midRank := (int(from.Rank) + int(to.Rank)) / 2
			target := address.MakeAddr(address.Rank(midRank), from.File)
			r.State.SetEnPassant(&target)
		}
	}
}

func (r *RuleEngine) updateCastlingRights(p pieces.Piece, from address.Addr) {
	if _, ok := p.(*pieces.King); ok {
		r.State.RevokeCastling(p.Color())
		return
	}

	if _, ok := p.(*pieces.Rook); ok {
		if from.Equals(address.MakeAddr(0, 7)) {
			r.State.RevokeSide("K")
		}
		if from.Equals(address.MakeAddr(0, 0)) {
			r.State.RevokeSide("Q")
		}
		if from.Equals(address.MakeAddr(7, 7)) {
			r.State.RevokeSide("k")
		}
		if from.Equals(address.MakeAddr(7, 0)) {
			r.State.RevokeSide("q")
		}
	}
}

func (r *RuleEngine) IsLegalMove(from, to address.Addr) bool {
	piece := r.Board.PieceAt(from)
	if piece == nil || piece.Color() != r.Turn {
		return false
	}

	legalMoves := piece.ValidMoves(from, r.Board, r.State)
	for _, move := range legalMoves {
		if move.Equals(to) {
			// Special check for Castling: Cannot castle out of, through, or into check.
			if _, isKing := piece.(*pieces.King); isKing {
				df := int(to.File) - int(from.File)
				if df == 2 || df == -2 {
					// 1. Cannot castle if currently in check
					if r.IsInCheck(r.Turn) {
						return false
					}
					// 2. Cannot castle through check (check the middle square)
					midFile := int(from.File) + (df / 2)
					midSquare := address.MakeAddr(from.Rank, address.File(midFile))
					if r.WouldBeInCheck(from, midSquare) {
						return false
					}
				}
			}

			return !r.WouldBeInCheck(from, to)
		}
	}
	return false
}

// --- Existing Helper Methods (Required for Compilation) ---

func (r *RuleEngine) GetTurn() int { return r.Turn }

func (r *RuleEngine) IsInCheck(color int) bool {
	kingPos := findKing(r.Board, color)
	if kingPos == nil {
		return false
	}
	k := *kingPos
	enemyColor := 1 - color

	// Knights
	knightDeltas := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	for _, d := range knightDeltas {
		if pos, ok := k.Shift(d[0], d[1]); ok {
			p := r.Board.PieceAt(pos)
			if p != nil && p.Color() == enemyColor {
				if _, isKnight := p.(*pieces.Knight); isKnight {
					return true
				}
			}
		}
	}
	// Pawns
	pawnDir := 1
	if color == pieces.BLACK {
		pawnDir = -1
	}
	for _, dx := range []int{-1, 1} {
		if pos, ok := k.Shift(pawnDir, dx); ok {
			p := r.Board.PieceAt(pos)
			if p != nil && p.Color() == enemyColor {
				if _, isPawn := p.(*pieces.Pawn); isPawn {
					return true
				}
			}
		}
	}
	// Sliders
	if scanDirs(r.Board, k, [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}, enemyColor, true, false) {
		return true
	}
	if scanDirs(r.Board, k, [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}, enemyColor, false, true) {
		return true
	}
	return false
}

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

func (r *RuleEngine) WouldBeInCheck(from, to address.Addr) bool {
	moving := r.Board.PieceAt(from)
	captured := r.Board.PieceAt(to)
	r.Board.SetPiece(to, moving)
	r.Board.Clear(from)
	inCheck := r.IsInCheck(moving.Color())
	r.Board.SetPiece(from, moving)
	r.Board.SetPiece(to, captured)
	return inCheck
}

func findKing(b *board.Board, color int) *address.Addr {
	for pos, p := range b.All() {
		if p.Color() == color {
			if _, ok := p.(*pieces.King); ok {
				ref := pos
				return &ref
			}
		}
	}
	return nil
}
