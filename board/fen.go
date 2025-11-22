package board

import (
	"fmt"
	"strings"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

// ToFEN serializes the entire game state into a standard FEN string.
func (b *Board) ToFEN(state *GameState) string {
	var fen strings.Builder

	// 1. Piece Placement
	for r := 7; r >= 0; r-- {
		emptyCount := 0
		for f := 0; f < 8; f++ {
			p := b.PieceAt(address.MakeAddr(address.Rank(r), address.File(f)))
			if p == nil {
				emptyCount++
			} else {
				if emptyCount > 0 {
					fen.WriteString(fmt.Sprintf("%d", emptyCount))
					emptyCount = 0
				}
				fen.WriteRune(pieceToFenChar(p))
			}
		}
		if emptyCount > 0 {
			fen.WriteString(fmt.Sprintf("%d", emptyCount))
		}
		if r > 0 {
			fen.WriteRune('/')
		}
	}

	// 2. Active Color
	turn := "w"
	if state.Turn == pieces.BLACK {
		turn = "b"
	}
	fen.WriteString(" " + turn)

	// 3. Castling Rights
	rights := state.CastlingRights
	if rights == "" {
		rights = "-"
	}
	fen.WriteString(" " + rights)

	// 4. En Passant Target
	ep := "-"
	if state.EnPassant != nil {
		ep = state.EnPassant.String()[:2] // "e3" from "e3::20"
	}
	fen.WriteString(" " + ep)

	// 5. Halfmove Clock & Fullmove Number
	fen.WriteString(fmt.Sprintf(" %d %d", state.HalfmoveClock, state.FullmoveNumber))

	return fen.String()
}

func pieceToFenChar(p pieces.Piece) rune {
	var c rune
	switch p.(type) {
	case *pieces.Pawn:
		c = 'p'
	case *pieces.Rook:
		c = 'r'
	case *pieces.Knight:
		c = 'n'
	case *pieces.Bishop:
		c = 'b'
	case *pieces.Queen:
		c = 'q'
	case *pieces.King:
		c = 'k'
	}

	if p.Color() == pieces.WHITE {
		return c - 32 // Convert to Uppercase (ASCII math)
	}
	return c
}
