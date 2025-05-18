// --- shell/renderer.go ---

package shell

import (
	"fmt"
	"strings"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

// Renderer defines a pluggable visual output interface.
type Renderer interface {
	Render(*board.Board)
	Prompt(turn int)
	Message(msg string)
	ShowCaptured(map[int][]pieces.Piece)
}

// TerminalRenderer implements Renderer using ANSI CLI output.
type TerminalRenderer struct{}

func (TerminalRenderer) Render(b *board.Board) {
	fmt.Print("\033[2J\033[H")
	for r := 7; r >= 0; r-- {
		fmt.Printf("%d  ", r+1)
		for f := 0; f < 8; f++ {
			a := address.MakeAddr(address.Rank(r), address.File(f))
			p := b.PieceAt(a)
			if p == nil {
				fmt.Print(" -- ")
			} else {
				fmt.Printf(" %2s ", p.String())
			}
		}
		fmt.Println()
	}
	fmt.Println("    a   b   c   d   e   f   g   h")
	fmt.Println()
}

func (TerminalRenderer) Prompt(turn int) {
	if turn == pieces.WHITE {
		fmt.Print("♙ White to move ~> ")
	} else {
		fmt.Print("♟ Black to move ~> ")
	}
}

func (TerminalRenderer) Message(msg string) {
	lines := strings.Split(msg, "\n")
	for _, l := range lines {
		fmt.Println("  ", l)
	}
	fmt.Println()
}

func (TerminalRenderer) ShowCaptured(captured map[int][]pieces.Piece) {
	fmt.Print("Captured ♟: ")
	for _, p := range captured[pieces.WHITE] {
		fmt.Print(p.String(), " ")
	}
	fmt.Println()

	fmt.Print("Captured ♙: ")
	for _, p := range captured[pieces.BLACK] {
		fmt.Print(p.String(), " ")
	}
	fmt.Println()
}
