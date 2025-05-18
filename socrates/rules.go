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
	Turn  int
	Log   *Log
}

// New creates a new rule engine with the given board and a fresh move log.
func New(b *board.Board) *RuleEngine {
	return &RuleEngine{
		Board: b,
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

	// Record move before applying
	if r.Log != nil {
		r.Log.Record(from, to, moving, target)
	}

	r.Board.SetPiece(to, moving)
	r.Board.Clear(from)
	r.Turn = 1 - r.Turn // toggle turn
	return true
}

// GetTurn returns the current player's color.
func (r *RuleEngine) GetTurn() int {
	return r.Turn
}

// IsInCheck returns true if the player with the given color is currently in check.
func (r *RuleEngine) IsInCheck(color int) bool {
	kingPos := findKing(r.Board, color)
	if kingPos == nil {
		return false
	}

	for _, pos := range allActive(r.Board, 1-color) {
		enemy := r.Board.PieceAt(pos)
		if enemy == nil {
			continue
		}
		moves := enemy.ValidMoves(pos, r.Board)
		for _, m := range moves {
			if m.Equals(*kingPos) {
				return true
			}
		}
	}
	return false
}

// WouldBeInCheck simulates a move and returns true if it exposes your own king.
func (r *RuleEngine) WouldBeInCheck(from, to address.Addr) bool {
	copy := *r.Board
	moving := copy.PieceAt(from)
	copy.SetPiece(to, moving)
	copy.Clear(from)

	shadow := &RuleEngine{Board: &copy}
	return shadow.IsInCheck(moving.Color())
}

// IsLegalMove determines whether a move from → to is valid under current rules.
func (r *RuleEngine) IsLegalMove(from, to address.Addr) bool {
	piece := r.Board.PieceAt(from)
	if piece == nil || piece.Color() != r.Turn {
		return false
	}

	legalMoves := piece.ValidMoves(from, r.Board)
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

// allActive returns the addresses of all active pieces of a color.
func allActive(b *board.Board, color int) []address.Addr {
	result := []address.Addr{}
	for pos, p := range b.All() {
		if p.Color() == color {
			result = append(result, pos)
		}
	}
	return result
}
