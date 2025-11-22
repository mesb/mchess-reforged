// --- pgn/pgn.go ---

// Package pgn provides a simple encoder/decoder for storing and loading games
// in Portable Game Notation (PGN), a standard format used by chess engines.
package pgn

import (
	"fmt"
	"os"
	"strings"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/socrates"
)

// Save writes the game log as a PGN file.
func Save(log *socrates.Log, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var builder strings.Builder
	// Add standard headers
	builder.WriteString("[Event \"MCHESS Game\"]\n")
	builder.WriteString("[Site \"Local CLI\"]\n")
	builder.WriteString("\n")

	for i, m := range log.Moves() {
		if i%2 == 0 {
			builder.WriteString(fmt.Sprintf("%d. ", i/2+1))
		}
		builder.WriteString(fmt.Sprintf("%s%s ", formatSquare(m.From), formatSquare(m.To)))
	}
	// Add result terminator
	builder.WriteString("*")

	_, err = file.WriteString(builder.String())
	return err
}

// Load parses a PGN file and replays it on the given rule engine.
func Load(engine *socrates.RuleEngine, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	words := strings.Fields(string(data))
	for _, word := range words {
		if strings.Contains(word, ".") {
			continue
		}
		if len(word) != 4 {
			continue
		}
		from := parseSquare(word[:2])
		to := parseSquare(word[2:])
		// Inside pgn.Load loop:
		if !engine.MakeMove(from, to, 0) { // Pass 0 for no specific promotion preference
			return fmt.Errorf("illegal move in PGN: %s", word)
		}
	}
	return nil
}

// formatSquare turns an Addr into PGN string like "e2"
func formatSquare(a address.Addr) string {
	return fmt.Sprintf("%c%d", a.File.Char(), int(a.Rank)+1)
}

// parseSquare reads a square like "e2" into an Addr
func parseSquare(s string) address.Addr {
	f := int(s[0] - 'a')
	r := int(s[1] - '1')
	return address.MakeAddr(address.Rank(r), address.File(f))
}
