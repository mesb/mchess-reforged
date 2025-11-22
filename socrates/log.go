// --- socrates/log.go ---

// This file implements a simple move log and undo mechanism for mchess.
package socrates

import (
	"fmt"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

// Move represents a single action taken on the board.
type Move struct {
	From   address.Addr
	To     address.Addr
	Piece  pieces.Piece
	Target pieces.Piece // may be nil

	// TargetPos records where Target was removed from (needed for en passant).
	TargetPos *address.Addr

	// RookMove records the rook displacement during castling.
	RookMove *CastleMove

	// PrevState restores the full game state (turn, clocks, EP, castling).
	PrevState StateSnapshot
}

// CastleMove describes the rook motion in a castle.
type CastleMove struct {
	From address.Addr
	To   address.Addr
}

// StateSnapshot is a copy of GameState, including a safe clone of EnPassant.
type StateSnapshot struct {
	Turn           int
	CastlingRights string
	EnPassant      *address.Addr
	HalfmoveClock  int
	FullmoveNumber int
}

func snapshotState(s *board.GameState) StateSnapshot {
	var ep *address.Addr
	if s.EnPassant != nil {
		addrCopy := *s.EnPassant
		ep = &addrCopy
	}
	return StateSnapshot{
		Turn:           s.Turn,
		CastlingRights: s.CastlingRights,
		EnPassant:      ep,
		HalfmoveClock:  s.HalfmoveClock,
		FullmoveNumber: s.FullmoveNumber,
	}
}

func applySnapshot(s *board.GameState, snap StateSnapshot) {
	s.Turn = snap.Turn
	s.CastlingRights = snap.CastlingRights
	s.HalfmoveClock = snap.HalfmoveClock
	s.FullmoveNumber = snap.FullmoveNumber
	if snap.EnPassant != nil {
		addrCopy := *snap.EnPassant
		s.EnPassant = &addrCopy
	} else {
		s.EnPassant = nil
	}
}

// Log tracks the full history of moves for a game.
type Log struct {
	moves []Move
}

// Record stores a move after it has occurred.
func (l *Log) Record(m Move) {
	l.moves = append(l.moves, m)
}

// UndoMove reverts the last move.
func (r *RuleEngine) UndoMove() bool {
	if r.Log == nil || len(r.Log.moves) == 0 {
		return false
	}

	// Pop last move
	latest := r.Log.moves[len(r.Log.moves)-1]
	r.Log.moves = r.Log.moves[:len(r.Log.moves)-1]

	applySnapshot(r.State, latest.PrevState)
	r.Turn = r.State.Turn

	// Undo castling rook move first so the king restore does not overwrite it.
	if latest.RookMove != nil {
		rook := r.Board.PieceAt(latest.RookMove.To)
		r.Board.SetPiece(latest.RookMove.From, rook)
		r.Board.Clear(latest.RookMove.To)
	}

	r.Board.SetPiece(latest.From, latest.Piece)

	if latest.Target != nil {
		restoreTo := latest.To
		if latest.TargetPos != nil {
			restoreTo = *latest.TargetPos
		}
		r.Board.SetPiece(restoreTo, latest.Target)
	} else {
		r.Board.Clear(latest.To)
	}

	if len(r.hashHistory) > 1 {
		r.hashHistory = r.hashHistory[:len(r.hashHistory)-1]
		r.hash = r.hashHistory[len(r.hashHistory)-1]
	} else if len(r.hashHistory) == 1 {
		r.hash = r.hashHistory[0]
	}

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
