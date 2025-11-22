// --- board/fen.go ---

package board

import (
	"fmt"
	"strings"

	"github.com/mesb/mchess/address"
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

// FromFEN hydrates a board and game state from a FEN string.
func FromFEN(fen string) (*Board, *GameState, error) {
	parts := strings.Fields(fen)
	if len(parts) != 6 {
		return nil, nil, fmt.Errorf("invalid FEN: expected 6 fields, got %d", len(parts))
	}

	board := NewBoard()
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return nil, nil, fmt.Errorf("invalid piece placement")
	}

	for fenRank, row := range ranks {
		boardRank := 7 - fenRank // FEN lists 8->1
		file := 0
		for _, c := range row {
			if c >= '1' && c <= '8' {
				file += int(c - '0')
				continue
			}
			p, err := pieceFromFENChar(c)
			if err != nil {
				return nil, nil, err
			}
			if file >= 8 {
				return nil, nil, fmt.Errorf("file overflow in rank %d", boardRank)
			}
			addr := address.MakeAddr(address.Rank(boardRank), address.File(file))
			board.SetPiece(addr, p)
			file++
		}
		if file != 8 {
			return nil, nil, fmt.Errorf("rank %d does not have 8 files", boardRank)
		}
	}

	state := &GameState{}

	// Active color
	switch parts[1] {
	case "w":
		state.Turn = pieces.WHITE
	case "b":
		state.Turn = pieces.BLACK
	default:
		return nil, nil, fmt.Errorf("invalid active color")
	}

	state.CastlingRights = parts[2]
	if state.CastlingRights == "" {
		state.CastlingRights = "-"
	}

	// En passant
	if parts[3] != "-" {
		if len(parts[3]) != 2 {
			return nil, nil, fmt.Errorf("invalid en passant square")
		}
		ep := address.MakeAddr(address.Rank(parts[3][1]-'1'), address.File(parts[3][0]-'a'))
		state.EnPassant = &ep
	}

	// Halfmove and fullmove
	if _, err := fmt.Sscanf(parts[4], "%d", &state.HalfmoveClock); err != nil {
		return nil, nil, fmt.Errorf("invalid halfmove clock")
	}
	if _, err := fmt.Sscanf(parts[5], "%d", &state.FullmoveNumber); err != nil {
		return nil, nil, fmt.Errorf("invalid fullmove number")
	}

	return board, state, nil
}

func pieceFromFENChar(c rune) (pieces.Piece, error) {
	color := pieces.WHITE
	if c >= 'a' && c <= 'z' {
		color = pieces.BLACK
		c -= 32
	}
	switch c {
	case 'P':
		return pieces.NewPawn(color), nil
	case 'R':
		return pieces.NewRook(color), nil
	case 'N':
		return pieces.NewKnight(color), nil
	case 'B':
		return pieces.NewBishop(color), nil
	case 'Q':
		return pieces.NewQueen(color), nil
	case 'K':
		return pieces.NewKing(color), nil
	default:
		return nil, fmt.Errorf("invalid piece char %c", c)
	}
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
