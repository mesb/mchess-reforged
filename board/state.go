// --- board/state.go ---

package board

import (
	"strings"

	"github.com/mesb/mchess/address"
)

type GameState struct {
	Turn           int
	CastlingRights string
	EnPassant      *address.Addr
	HalfmoveClock  int
	FullmoveNumber int
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

func (s *GameState) GetEnPassant() *address.Addr  { return s.EnPassant }
func (s *GameState) SetEnPassant(t *address.Addr) { s.EnPassant = t }
func (s *GameState) GetCastlingRights() string    { return s.CastlingRights }

// RevokeCastling removes rights for a specific color (0=White "KQ", 1=Black "kq").
func (s *GameState) RevokeCastling(color int) {
	toRemove := "KQ"
	if color == 1 {
		toRemove = "kq"
	}
	for _, r := range toRemove {
		s.CastlingRights = strings.ReplaceAll(s.CastlingRights, string(r), "")
	}
	if s.CastlingRights == "" {
		s.CastlingRights = "-"
	}
}

// RevokeSide removes a specific right (e.g. "Q" or "k").
func (s *GameState) RevokeSide(right string) {
	s.CastlingRights = strings.ReplaceAll(s.CastlingRights, right, "")
	if s.CastlingRights == "" {
		s.CastlingRights = "-"
	}
}

// IncrementClock updates the halfmove clock for the 50-move rule.
func (s *GameState) IncrementClock(isPawnMove, isCapture bool) {
	if isPawnMove || isCapture {
		s.HalfmoveClock = 0
	} else {
		s.HalfmoveClock++
	}
}
