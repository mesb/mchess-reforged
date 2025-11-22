// --- pgn/pgn.go ---

package pgn

import (
	"fmt"
	"os"
	"strings"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/socrates"
)

// Export converts the game log into a PGN string.
func Export(log *socrates.Log) string {
	var builder strings.Builder
	builder.WriteString("[Event \"MCHESS Game\"]\n")
	builder.WriteString("[Site \"MCHESS Server\"]\n")
	builder.WriteString("\n")

	for i, m := range log.Moves() {
		if i%2 == 0 {
			builder.WriteString(fmt.Sprintf("%d. ", i/2+1))
		}
		builder.WriteString(fmt.Sprintf("%s%s ", formatSquare(m.From), formatSquare(m.To)))
	}
	builder.WriteString("*")
	return builder.String()
}

// Import parses a PGN string and replays it on the engine.
func Import(engine *socrates.RuleEngine, data string) error {
	words := strings.Fields(data)
	for _, word := range words {
		if strings.Contains(word, ".") || strings.Contains(word, "[") || strings.Contains(word, "]") {
			continue
		}
		if len(word) < 4 {
			continue
		}

		// Handle standard moves (e2e4) and promotions (a7a8q)
		// Simplified parsing for PGN replay
		clean := strings.TrimSuffix(word, "+") // Remove check indicator if present
		clean = strings.TrimSuffix(clean, "#") // Remove mate indicator

		from := parseSquare(clean[:2])
		to := parseSquare(clean[2:4])

		var promo rune
		if len(clean) == 5 {
			promo = rune(clean[4])
		}

		if !engine.MakeMove(*from, *to, promo) {
			return fmt.Errorf("illegal move in PGN: %s", word)
		}
	}
	return nil
}

// --- File Wrappers (Backward Compatibility) ---

func Save(log *socrates.Log, filename string) error {
	data := Export(log)
	return os.WriteFile(filename, []byte(data), 0644)
}

func Load(engine *socrates.RuleEngine, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return Import(engine, string(data))
}

// --- Helpers ---

func formatSquare(a address.Addr) string {
	return fmt.Sprintf("%c%d", a.File.Char(), int(a.Rank)+1)
}

func parseSquare(s string) *address.Addr {
	if len(s) < 2 {
		return nil
	}
	f := int(s[0] - 'a')
	r := int(s[1] - '1')
	a := address.MakeAddr(address.Rank(r), address.File(f))
	return &a
}
