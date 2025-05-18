// --- address/address.go ---

// Package address provides the core spatial abstraction for mchess.
// It defines the addressing scheme and motion mechanics on the 8x8 grid.
package address

import (
	"fmt"
)

const (
	BoardSize  = 8  // Number of squares per rank or file
	NumSquares = 64 // Total number of squares on the board
	MaxIndex   = 63 // Maximum valid square index
	MinIndex   = 0  // Minimum valid square index
)

// Rank represents vertical position (1 through 8)
type Rank uint

// File represents horizontal position (a through h)
type File uint

// Addr represents a chess square using rank and file coordinates
// and supports translation to and from linear indices.
type Addr struct {
	Rank
	File
}

// MakeAddr constructs a new address from rank and file.
// It provides a declarative alternative to manually building Addr structs.
func MakeAddr(r Rank, f File) Addr {
	return Addr{Rank: r, File: f}
}

// Index returns the linear index [0–63] of the address.
// It follows row-major order: index = rank * 8 + file
func (a Addr) Index() int {
	return int(a.Rank)*BoardSize + int(a.File)
}

// Equals returns true if two addresses refer to the same square.
func (a Addr) Equals(b Addr) bool {
	return a.File == b.File && a.Rank == b.Rank
}

// Shift returns a new Addr offset by (dr, df).
// It safely bounds the result and returns ok = false if off-board.
func (a Addr) Shift(dr, df int) (Addr, bool) {
	newR := int(a.Rank) + dr
	newF := int(a.File) + df
	if newR < 0 || newR >= BoardSize || newF < 0 || newF >= BoardSize {
		return Addr{}, false
	}
	return Addr{Rank(newR), File(newF)}, true
}

// TranslateIndex converts a linear index back to a 2D address.
func TranslateIndex(index int) Addr {
	rank := Rank(index / BoardSize)
	file := File(index % BoardSize)
	return Addr{rank, file}
}

// Delta returns the rank and file deltas between two addresses.
func Delta(from, to Addr) (int, int) {
	return int(to.Rank) - int(from.Rank), int(to.File) - int(from.File)
}

// Char returns the character representation of the file (a–h).
func (f File) Char() rune {
	return rune('a' + f)
}

// String formats an Addr into user-readable notation (e.g. "e2::12").
func (a Addr) String() string {
	return fmt.Sprintf("%c%v::%2d", a.File.Char(), a.Rank+1, a.Index())
}
