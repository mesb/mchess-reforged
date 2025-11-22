// --- board/state.go ---

// GameState tracks the meta-properties of a match such as turn, castling rights, and move clocks.
package board

import "github.com/mesb/mchess/address"

type GameState struct {
	Turn           int           // Current player's color (WHITE or BLACK)
	CastlingRights string        // e.g., "KQkq"
	EnPassant      *address.Addr // Target square if en passant is available
	HalfmoveClock  int           // For 50-move rule
	FullmoveNumber int           // Starts at 1, incremented after Black's move
}

func NewGameState() *GameState {
	return &GameState{
		Turn:           0,
		CastlingRights: "KQkq",
		EnPassant:      nil,
		HalfmoveClock:  0,
		FullmoveNumber: 1,
	}
}

func (s *GameState) SwitchTurn() {
	s.Turn = 1 - s.Turn
	if s.Turn == 0 {
		s.FullmoveNumber++
	}
}

// GetEnPassant returns the current En Passant target square (satisfies pieces.GameStateView).
func (s *GameState) GetEnPassant() *address.Addr {
	return s.EnPassant
}

// SetEnPassant updates the En Passant target square.
func (s *GameState) SetEnPassant(target *address.Addr) {
	s.EnPassant = target
}
