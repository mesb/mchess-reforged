// --- board/board.go ---

// Package board defines the central state structure of the chessboard.
// It stores the 8x8 grid and provides accessors for querying positions.
package board

import (
	"fmt"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

const (
	FIRSTRANK = iota
	SECONDRANK
	THIRDRANK
	FOURTHRANK
	FIFTHRANK
	SIXTHRANK
	SEVENTHRANK
	EIGHTHRANK
)

const (
	A = iota
	B
	C
	D
	E
	F
	G
	H
)

var DEFAULTFIRSTRANK = [8]pieces.Piece{
	pieces.NewRook(pieces.WHITE),
	pieces.NewKnight(pieces.WHITE),
	pieces.NewBishop(pieces.WHITE),
	pieces.NewQueen(pieces.WHITE),
	pieces.NewKing(pieces.WHITE),
	pieces.NewBishop(pieces.WHITE),
	pieces.NewKnight(pieces.WHITE),
	pieces.NewRook(pieces.WHITE),
}

// Board holds the active state of the 64-square chess grid.
type Board struct {
	squares [address.NumSquares]pieces.Piece
}

// Ensure Board satisfies the BoardView interface
var _ pieces.BoardView = (*Board)(nil)

// NewBoard initializes an empty board.
func NewBoard() *Board {
	return &Board{}
}

// InitStandard sets up a default game with both camps populated.
func InitStandard() *Board {
	b := NewBoard()

	for i := 0; i < 8; i++ {
		// White major pieces
		b.SetPiece(address.MakeAddr(address.Rank(FIRSTRANK), address.File(i)), DEFAULTFIRSTRANK[i])
		// White pawns
		b.SetPiece(address.MakeAddr(address.Rank(SECONDRANK), address.File(i)), pieces.NewPawn(pieces.WHITE))
		// Black major pieces
		b.SetPiece(address.MakeAddr(address.Rank(EIGHTHRANK), address.File(i)), pieces.CloneWithColor(DEFAULTFIRSTRANK[i], pieces.BLACK))
		// Black pawns
		b.SetPiece(address.MakeAddr(address.Rank(SEVENTHRANK), address.File(i)), pieces.NewPawn(pieces.BLACK))
	}

	return b
}

// PieceAt returns the piece at a given square, or nil if empty.
func (b *Board) PieceAt(a address.Addr) pieces.Piece {
	return b.squares[a.Index()]
}

// SetPiece places a piece at a given address.
func (b *Board) SetPiece(a address.Addr, p pieces.Piece) {
	b.squares[a.Index()] = p
}

// IsEmpty returns true if a square is unoccupied.
func (b *Board) IsEmpty(a address.Addr) bool {
	return b.PieceAt(a) == nil
}

// Clear removes any piece from the given square.
func (b *Board) Clear(a address.Addr) {
	b.SetPiece(a, nil)
}

// All returns a map of all non-empty squares and their pieces.
func (b *Board) All() map[address.Addr]pieces.Piece {
	result := make(map[address.Addr]pieces.Piece)
	for i := 0; i < address.NumSquares; i++ {
		sq := address.TranslateIndex(i)
		p := b.PieceAt(sq)
		if p != nil {
			result[sq] = p
		}
	}
	return result
}

// ForEachPiece walks the board without allocations.
func (b *Board) ForEachPiece(fn func(address.Addr, pieces.Piece)) {
	for i, p := range b.squares {
		if p == nil {
			continue
		}
		fn(address.TranslateIndex(i), p)
	}
}

// Print renders the board row by row for debugging.
func Print(b *Board) {
	for i := 1; i <= address.NumSquares; i++ {
		j := i - 1
		piece := b.squares[j]
		if piece == nil {
			fmt.Print("-- ")
		} else {
			fmt.Printf("%s ", piece.String())
		}
		if i%8 == 0 {
			fmt.Println()
		}
	}
}

// FindKing returns the position of the king of the given color, if present.
func (b *Board) FindKing(color int) *address.Addr {
	for i, p := range b.squares {
		if p == nil || p.Color() != color {
			continue
		}
		if _, ok := p.(*pieces.King); ok {
			addr := address.TranslateIndex(i)
			return &addr
		}
	}
	return nil
}
