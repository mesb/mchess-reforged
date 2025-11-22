// --- shell/session.go ---

package shell

import (
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
	"github.com/mesb/mchess/socrates"
	"sync"
)

// GameSession represents an active chess session with state and utilities.
type GameSession struct {
	Mu       sync.RWMutex
	Engine   *socrates.RuleEngine
	Captured map[int][]pieces.Piece
	Log      *socrates.Log
	Renderer Renderer
}

// NewSession creates a fully initialized session with default board and state.
func NewSession(renderer Renderer) *GameSession {
	log := &socrates.Log{}

	// Initialize the board and the game state (turn, castling, en passant)
	b := board.InitStandard()
	state := board.NewGameState()

	engine := &socrates.RuleEngine{
		Board: b,
		State: state, // Inject the State here
		Turn:  pieces.WHITE,
		Log:   log,
	}
	engine.ResetHashHistory()

	return &GameSession{
		Engine:   engine,
		Captured: map[int][]pieces.Piece{},
		Log:      log,
		Renderer: renderer,
	}
}

// UpdateCaptured rebuilds captured-piece tracking from the move log.
func (s *GameSession) UpdateCaptured() {
	s.Captured = map[int][]pieces.Piece{
		pieces.WHITE: {},
		pieces.BLACK: {},
	}
	if s.Log == nil {
		return
	}
	for _, m := range s.Log.Moves() {
		if m.Target != nil {
			c := m.Target.Color()
			s.Captured[c] = append(s.Captured[c], m.Target)
		}
	}
}
