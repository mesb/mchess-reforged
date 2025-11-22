package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

// HashTogglePiece XORs the hash with the key for a piece at a square.
func HashTogglePiece(h uint64, p pieces.Piece, a address.Addr) uint64 {
	idx := pieceIndex(p)
	if idx < 0 {
		return h
	}
	return h ^ pieceKeys[p.Color()][idx][a.Index()]
}

// HashToggleCastling XORs the hash with the castling rights key.
func HashToggleCastling(h uint64, rights string) uint64 {
	return h ^ castleKeys[castleIndex(rights)]
}

// HashToggleEP XORs the hash with the en-passant file key.
func HashToggleEP(h uint64, ep *address.Addr) uint64 {
	if ep == nil {
		return h
	}
	return h ^ epKeys[int(ep.File)]
}

// HashToggleTurn XORs the side-to-move key.
func HashToggleTurn(h uint64) uint64 {
	return h ^ turnKey
}
