// --- shell/session.go ---

package shell

import (
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
	"github.com/mesb/mchess/socrates"
)

// GameSession represents an active chess session with state and utilities.
type GameSession struct {
	Engine   *socrates.RuleEngine
	Captured map[int][]pieces.Piece
	Log      *socrates.Log
	Renderer Renderer
}

// NewSession creates a fully initialized session with default board and state.
func NewSession(renderer Renderer) *GameSession {
	log := &socrates.Log{}
	engine := &socrates.RuleEngine{
		Board: board.InitStandard(),
		Turn:  pieces.WHITE,
		Log:   log,
	}

	return &GameSession{
		Engine:   engine,
		Captured: map[int][]pieces.Piece{},
		Log:      log,
		Renderer: renderer,
	}
}
