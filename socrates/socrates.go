// --- socrates/socrates.go ---

package socrates

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mesb/mchess/address"
)

// ParseMove converts a string like "e2e4" or "a7a8q" into usable coordinates.
// This is a pure function: no engine state required.
func ParseMove(notation string) (from, to *address.Addr, promo rune, err error) {
	if len(notation) < 4 || len(notation) > 5 {
		return nil, nil, 0, errors.New("invalid notation: expected format 'e2e4' or 'a7a8q'")
	}

	from = parseSquare(notation[:2])
	to = parseSquare(notation[2:4])
	if from == nil || to == nil {
		return nil, nil, 0, errors.New("invalid square coordinates")
	}

	if len(notation) == 5 {
		promo = rune(notation[4])
	}

	return from, to, promo, nil
}

// Dialog parses and performs a move command for the CLI (e.g., "m e2e4").
func Dialog(input string, engine *RuleEngine) error {
	fields := strings.Fields(input)
	if len(fields) != 2 {
		return errors.New("invalid command: use 'm e2e4'")
	}

	from, to, promo, err := ParseMove(fields[1])
	if err != nil {
		return err
	}

	if !engine.MakeMove(*from, *to, promo) {
		return fmt.Errorf("illegal move %s", fields[1])
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
