// --- socrates/rules.go ---

// This file defines the RuleEngine, the primary logic engine for enforcing legal chess moves.
package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

// RuleEngine represents the core chess logic engine.
type RuleEngine struct {
	Board *board.Board
	State *board.GameState // Added GameState to track metadata like En Passant
	Turn  int
	Log   *Log
}

// New creates a new rule engine with the given board and a fresh move log.
func New(b *board.Board) *RuleEngine {
	return &RuleEngine{
		Board: b,
		State: board.NewGameState(), // Initialize default game state
		Turn:  pieces.WHITE,
		Log:   &Log{},
	}
}

// MakeMove executes a move if legal, records it, and switches turns.
func (r *RuleEngine) MakeMove(from, to address.Addr) bool {
	if !r.IsLegalMove(from, to) {
		return false
	}

	moving := r.Board.PieceAt(from)
	target := r.Board.PieceAt(to)

	// Handle En Passant Capture Logic
	// If a pawn moves to the En Passant target square, it's a capture,
	// but the target square itself is empty on the board.
	if p, ok := moving.(*pieces.Pawn); ok {
		if ep := r.State.GetEnPassant(); ep != nil && to.Equals(*ep) {
			// Capture the pawn BEHIND the moving pawn
			// White moves Up (positive rank), so victim is Down (Shift -1 rank)
			// Black moves Down (negative rank), so victim is Up (Shift +1 rank)
			captureRankDir := -1
			if p.Color() == pieces.BLACK {
				captureRankDir = 1
			}

			// The victim is at the 'to' file, but 'from' rank (conceptually adjacent to 'to' on previous rank)
			if victimPos, ok := to.Shift(captureRankDir, 0); ok {
				target = r.Board.PieceAt(victimPos) // Record for log
				r.Board.Clear(victimPos)            // Remove victim from board
			}
		}
	}

	// Record move before applying
	if r.Log != nil {
		r.Log.Record(from, to, moving, target)
	}

	// Execute Move
	r.Board.SetPiece(to, moving)
	r.Board.Clear(from)

	// Handle En Passant State Update (Must happen before turn switch)
	r.updateEnPassantState(moving, from, to)

	// Check for Promotion (Auto-Queen)
	// Note: A full implementation would ask the user, but this is a safe default.
	if p, ok := moving.(*pieces.Pawn); ok {
		rank := to.Rank
		if (p.Color() == pieces.WHITE && rank == 7) || (p.Color() == pieces.BLACK && rank == 0) {
			r.Board.SetPiece(to, pieces.NewQueen(p.Color()))
		}
	}

	r.Turn = 1 - r.Turn   // toggle turn
	r.State.Turn = r.Turn // Sync state
	return true
}

// updateEnPassantState checks if a pawn moved 2 squares and sets the flag
func (r *RuleEngine) updateEnPassantState(p pieces.Piece, from, to address.Addr) {
	r.State.SetEnPassant(nil) // Default: clear it

	if _, ok := p.(*pieces.Pawn); ok {
		dr, _ := address.Delta(from, to)
		if dr == 2 || dr == -2 {
			// Set target to the square passed over
			midRank := (int(from.Rank) + int(to.Rank)) / 2
			target := address.MakeAddr(address.Rank(midRank), from.File)
			r.State.SetEnPassant(&target)
		}
	}
}

// GetTurn returns the current player's color.
func (r *RuleEngine) GetTurn() int {
	return r.Turn
}

// IsInCheck determines if the King of the given color is under attack.
// OPTIMIZATION: Looks outward from the King rather than iterating all enemy pieces.
func (r *RuleEngine) IsInCheck(color int) bool {
	kingPos := findKing(r.Board, color)
	if kingPos == nil {
		return false // Should theoretically not happen in a valid game
	}
	k := *kingPos

	enemyColor := 1 - color

	// 1. Check for Knights
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

	// 2. Check for Pawns
	pawnDir := 1 // incoming attack direction depends on our color
	if color == pieces.BLACK {
		pawnDir = -1
	}
	// Pawns attack from diagonals
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

	// 3. Check Sliding Pieces (Rook, Bishop, Queen)
	// Orthogonal (Rook/Queen)
	orthoDirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	if scanDirs(r.Board, k, orthoDirs, enemyColor, true, false) {
		return true
	}

	// Diagonal (Bishop/Queen)
	diagDirs := [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	if scanDirs(r.Board, k, diagDirs, enemyColor, false, true) {
		return true
	}

	return false
}

// scanDirs sends a "ray" out from the King to see if a slider is aiming at it.
func scanDirs(b *board.Board, start address.Addr, dirs [][2]int, enemyColor int, checkRook, checkBishop bool) bool {
	for _, d := range dirs {
		for i := 1; i < 8; i++ {
			pos, ok := start.Shift(d[0]*i, d[1]*i)
			if !ok {
				break // Off board
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
				break // Blocked by any piece (friend or foe)
			}
		}
	}
	return false
}

// WouldBeInCheck simulates a move without copying the entire board.
// OPTIMIZATION: Uses "Make-Unmake" pattern.
func (r *RuleEngine) WouldBeInCheck(from, to address.Addr) bool {
	moving := r.Board.PieceAt(from)
	captured := r.Board.PieceAt(to)

	// 1. Make the move
	r.Board.SetPiece(to, moving)
	r.Board.Clear(from)

	// 2. Check legality
	inCheck := r.IsInCheck(moving.Color())

	// 3. Unmake the move (restore state)
	r.Board.SetPiece(from, moving)
	r.Board.SetPiece(to, captured) // captured is nil if empty, so this works safely

	return inCheck
}

// IsLegalMove determines whether a move from → to is valid under current rules.
func (r *RuleEngine) IsLegalMove(from, to address.Addr) bool {
	piece := r.Board.PieceAt(from)
	if piece == nil || piece.Color() != r.Turn {
		return false
	}

	// Note: ValidMoves now accepts GameState to check en passant/castling
	legalMoves := piece.ValidMoves(from, r.Board, r.State)
	for _, move := range legalMoves {
		if move.Equals(to) {
			return !r.WouldBeInCheck(from, to)
		}
	}
	return false
}

// findKing locates the king of the given color on the board.
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
