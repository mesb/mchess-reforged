// --- socrates/log.go ---

// This file implements a simple move log and undo mechanism for mchess.
package socrates

import (
	"fmt"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

// Move represents a single action taken on the board.
type Move struct {
	From   address.Addr
	To     address.Addr
	Piece  pieces.Piece
	Target pieces.Piece // may be nil
}

// Log tracks the full history of moves for a game.
type Log struct {
	moves []Move
}

// Record stores a move after it has occurred.
func (l *Log) Record(from, to address.Addr, moving, captured pieces.Piece) {
	l.moves = append(l.moves, Move{
		From:   from,
		To:     to,
		Piece:  moving,
		Target: captured,
	})
}

// UndoMove reverts the last move.
func (r *RuleEngine) UndoMove() bool {
	if r.Log == nil || len(r.Log.moves) == 0 {
		return false
	}

	// Pop last move
	latest := r.Log.moves[len(r.Log.moves)-1]
	r.Log.moves = r.Log.moves[:len(r.Log.moves)-1]

	r.Board.SetPiece(latest.From, latest.Piece)
	if latest.Target != nil {
		r.Board.SetPiece(latest.To, latest.Target)
	} else {
		r.Board.Clear(latest.To)
	}

	// Switch turn back
	r.Turn = 1 - r.Turn

	return true
}

// PrintLog prints the move history in human-readable format.
func (l *Log) PrintLog() {
	fmt.Println("\nMove History:")
	for i, m := range l.moves {
		fmt.Printf("%2d. %s → %s (%s)\n", i+1, m.From, m.To, m.Piece.String())
	}
}

// Moves returns the recorded move list (read-only access).
func (l *Log) Moves() []Move {
	return l.moves
}
