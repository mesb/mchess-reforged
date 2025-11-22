// --- socrates/socrates.go ---

package socrates

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mesb/mchess/address"
)

// Dialog parses and performs a move command, e.g., "m e2e4"
func Dialog(input string, engine *RuleEngine) error {
	fields := strings.Fields(input)
	if len(fields) != 2 {
		return errors.New("invalid command: use 'm e2e4'")
	}

	notation := fields[1]
	if len(notation) < 4 || len(notation) > 5 {
		return errors.New("invalid notation: expected format 'e2e4' or 'a7a8q'")
	}

	from := parseSquare(notation[:2])
	to := parseSquare(notation[2:4])
	if from == nil || to == nil {
		return errors.New("invalid square coordinates")
	}

	var promoChar rune
	if len(notation) == 5 {
		promoChar = rune(notation[4])
	}

	if !engine.MakeMove(*from, *to, promoChar) {
		return fmt.Errorf("illegal move %s", notation)
	}

	return nil
}

// parseSquare converts "e2" to an address.Addr
func parseSquare(s string) *address.Addr {
	if len(s) != 2 {
		return nil
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return nil
	}
	a := address.MakeAddr(address.Rank(rank), address.File(file))
	return &a
}
