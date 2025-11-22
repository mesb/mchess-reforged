// --- uci/uci.go ---

package uci

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/socrates"
)

// Run starts the UCI loop, listening to Stdin and writing to Stdout.
func Run() {
	scanner := bufio.NewScanner(os.Stdin)
	// Initialize with standard start position
	eng := socrates.New(board.InitStandard())

	for scanner.Scan() {
		line := scanner.Text()
		cmd := strings.Fields(line)
		if len(cmd) == 0 {
			continue
		}

		switch cmd[0] {
		case "uci":
			fmt.Println("id name MCHESS Dragon")
			fmt.Println("id author Hexa")
			fmt.Println("uciok")

		case "isready":
			fmt.Println("readyok")

		case "ucinewgame":
			eng = socrates.New(board.InitStandard())

		case "position":
			handlePosition(eng, cmd)

		case "go":
			handleGo(eng, cmd)

		case "quit":
			return
		}
	}
}

// handlePosition parses "position startpos moves e2e4..." or "position fen ... moves ..."
func handlePosition(eng *socrates.RuleEngine, args []string) {
	if len(args) < 2 {
		return
	}

	moveIdx := -1
	// 1. Reset Board
	if args[1] == "startpos" {
		*eng = *socrates.New(board.InitStandard())
		moveIdx = 2
	} else if args[1] == "fen" {
		// Join tokens until "moves" keyword
		var fenParts []string
		for i := 2; i < len(args); i++ {
			if args[i] == "moves" {
				moveIdx = i
				break
			}
			fenParts = append(fenParts, args[i])
		}
		fen := strings.Join(fenParts, " ")
		b, s, err := board.FromFEN(fen)
		if err == nil {
			eng.Board = b
			eng.State = s
			eng.Turn = s.Turn
			eng.ResetHashHistory()
		}
	}

	// 2. Apply Moves (if any)
	if moveIdx != -1 && moveIdx < len(args) && args[moveIdx] == "moves" {
		for i := moveIdx + 1; i < len(args); i++ {
			mvStr := args[i]
			from, to, promo, err := socrates.ParseMove(mvStr)
			if err == nil {
				eng.MakeMove(*from, *to, promo)
			}
		}
	}
}

// handleGo starts the search. Currently supports fixed depth or simple time management.
func handleGo(eng *socrates.RuleEngine, args []string) {
	// Default search parameters
	depth := 5

	// TODO: Implement iterative deepening and time management parsing here.
	_ = args

	start := time.Now()
	res := eng.Search(depth)
	elapsed := time.Since(start)

	// Output info string for the GUI
	// cp = centipawns score
	fmt.Printf("info depth %d score cp %d nodes %d time %d\n",
		depth, res.Score, res.Nodes, elapsed.Milliseconds())

	// Output the best move found
	fmt.Printf("bestmove %s%s\n", squareString(res.From), squareString(res.To))
}

func squareString(a address.Addr) string {
	return fmt.Sprintf("%c%d", a.File.Char(), int(a.Rank)+1)
}
