// --- pieces/piece.go ---

package pieces

import "github.com/mesb/mchess/address"

const (
	WHITE = 0
	BLACK = 1
)

// GameStateView defines the subset of state that pieces need to see.
// This prevents circular dependencies between 'pieces' and 'board'.
type GameStateView interface {
	GetEnPassant() *address.Addr
	GetCastlingRights() string
}

// Piece defines the behavior of any chess piece type.
// Pieces are stateless and express only their color and movement rules.
type Piece interface {
	Color() int
	String() string
	// ValidMoves now requires State to handle En Passant and Castling
	ValidMoves(from address.Addr, board BoardView, state GameStateView) []address.Addr
}

// BoardView is a minimal interface to query board state without creating import cycles.
type BoardView interface {
	IsEmpty(address.Addr) bool
	PieceAt(address.Addr) Piece
}
