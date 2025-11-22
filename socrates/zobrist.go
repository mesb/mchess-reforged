package socrates

import (
	"math/rand"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

var (
	pieceKeys   [2][6][64]uint64 // color, piece index, square
	castleKeys  [16]uint64
	epKeys      [8]uint64
	turnKey     uint64
	zobristInit bool
)

func initZobrist() {
	if zobristInit {
		return
	}
	rnd := rand.New(rand.NewSource(42))
	for c := 0; c < 2; c++ {
		for p := 0; p < 6; p++ {
			for sq := 0; sq < 64; sq++ {
				pieceKeys[c][p][sq] = rnd.Uint64()
			}
		}
	}
	for i := 0; i < 16; i++ {
		castleKeys[i] = rnd.Uint64()
	}
	for i := 0; i < 8; i++ {
		epKeys[i] = rnd.Uint64()
	}
	turnKey = rnd.Uint64()
	zobristInit = true
}

func computeHash(b *board.Board, state *board.GameState, turn int) uint64 {
	initZobrist()
	var h uint64

	b.ForEachPiece(func(a address.Addr, p pieces.Piece) {
		idx := pieceIndex(p)
		if idx >= 0 {
			h ^= pieceKeys[p.Color()][idx][a.Index()]
		}
	})

	h ^= castleKeys[castleIndex(state.CastlingRights)]

	if state.EnPassant != nil {
		h ^= epKeys[int(state.EnPassant.File)]
	}

	if turn == pieces.BLACK {
		h ^= turnKey
	}

	return h
}

func pieceIndex(p pieces.Piece) int {
	switch p.(type) {
	case *pieces.Pawn:
		return 0
	case *pieces.Knight:
		return 1
	case *pieces.Bishop:
		return 2
	case *pieces.Rook:
		return 3
	case *pieces.Queen:
		return 4
	case *pieces.King:
		return 5
	default:
		return -1
	}
}

func castleIndex(rights string) int {
	idx := 0
	for _, r := range rights {
		switch r {
		case 'K':
			idx |= 1 << 0
		case 'Q':
			idx |= 1 << 1
		case 'k':
			idx |= 1 << 2
		case 'q':
			idx |= 1 << 3
		}
	}
	return idx
}
