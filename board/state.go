// --- board/state.go ---

package board

import (
	"strings"

	"github.com/mesb/mchess/address"
)

// GameState tracks the meta-properties of a match.
type GameState struct {
	Turn           int           // Current player's color (WHITE or BLACK)
	CastlingRights string        // e.g., "KQkq", "Kq", "-"
	EnPassant      *address.Addr // Target square if en passant is available
	HalfmoveClock  int           // For 50-move rule
	FullmoveNumber int           // Starts at 1, incremented after Black's move
}

func NewGameState() *GameState {
	return &GameState{
		Turn:           0, // WHITE
		CastlingRights: "KQkq",
		EnPassant:      nil,
		HalfmoveClock:  0,
		FullmoveNumber: 1,
	}
}

// GetEnPassant returns the current En Passant target square.
func (s *GameState) GetEnPassant() *address.Addr {
	return s.EnPassant
}

// SetEnPassant updates the En Passant target square.
func (s *GameState) SetEnPassant(target *address.Addr) {
	s.EnPassant = target
}

// RevokeCastling removes all castling rights for a specific color.
// color 0 (White) removes "KQ", color 1 (Black) removes "kq".
func (s *GameState) RevokeCastling(color int) {
	toRemove := "KQ"
	if color == 1 {
		toRemove = "kq"
	}
	for _, char := range toRemove {
		s.CastlingRights = strings.ReplaceAll(s.CastlingRights, string(char), "")
	}
	if s.CastlingRights == "" {
		s.CastlingRights = "-"
	}
}

// RevokeSide removes a specific castling right (e.g., "Q" or "k").
func (s *GameState) RevokeSide(right string) {
	s.CastlingRights = strings.ReplaceAll(s.CastlingRights, right, "")
	if s.CastlingRights == "" {
		s.CastlingRights = "-"
	}
}

// IncrementClock handles the 50-move rule logic.
// Resets on pawn moves or captures, increments otherwise.
func (s *GameState) IncrementClock(isPawnMove bool, isCapture bool) {
	if isPawnMove || isCapture {
		s.HalfmoveClock = 0
	} else {
		s.HalfmoveClock++
	}
}
