// --- board/fen.go ---

package board

import (
	"fmt"
	"strings"

	"github.com/mesb/mchess/pieces"
)

// ToFEN serializes the entire game state into a standard FEN string.
// It leverages the internal 1D array layout for maximum efficiency.
func (b *Board) ToFEN(state *GameState) string {
	var fen strings.Builder
	// Pre-allocate approx capacity to minimize re-allocations (FEN max length ~90)
	fen.Grow(90)

	// 1. Piece Placement
	// FEN starts at Rank 8 (indices 56-63) and moves down to Rank 1 (indices 0-7)
	for r := 7; r >= 0; r-- {
		emptyCount := 0
		rowStart := r * 8 // Direct offset into 1D array

		for f := 0; f < 8; f++ {
			// Direct array access: O(1) with no address calculation overhead
			p := b.squares[rowStart+f]

			if p == nil {
				emptyCount++
			} else {
				if emptyCount > 0 {
					// Write accumulated empty squares
					fen.WriteByte(byte('0' + emptyCount)) // Fast int-to-byte conversion
					emptyCount = 0
				}
				fen.WriteRune(pieceToFenChar(p))
			}
		}

		if emptyCount > 0 {
			fen.WriteByte(byte('0' + emptyCount))
		}

		if r > 0 {
			fen.WriteByte('/')
		}
	}

	// 2. Active Color
	fen.WriteByte(' ')
	if state.Turn == pieces.BLACK {
		fen.WriteByte('b')
	} else {
		fen.WriteByte('w')
	}

	// 3. Castling Rights
	fen.WriteByte(' ')
	if state.CastlingRights == "" {
		fen.WriteByte('-')
	} else {
		fen.WriteString(state.CastlingRights)
	}

	// 4. En Passant Target
	fen.WriteByte(' ')
	if state.EnPassant != nil {
		// We only need the string representation here
		// Optimization: Get string directly or compute algebraic safely
		epStr := state.EnPassant.String()
		if len(epStr) >= 2 {
			fen.WriteString(epStr[:2])
		} else {
			fen.WriteByte('-')
		}
	} else {
		fen.WriteByte('-')
	}

	// 5. Halfmove Clock & Fullmove Number
	// Using fmt.Fprintf is cleaner here than manual string building for integers
	fmt.Fprintf(&fen, " %d %d", state.HalfmoveClock, state.FullmoveNumber)

	return fen.String()
}

// pieceToFenChar converts a piece to its FEN character (e.g., 'P', 'n', 'q').
// White is uppercase, Black is lowercase.
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
	default:
		return '?'
	}

	if p.Color() == pieces.WHITE {
		return c - 32 // ASCII optimization: 'a' (97) - 32 = 'A' (65)
	}
	return c
}
