package socrates

import (
	"math/rand"
	"time"

	"github.com/mesb/mchess/address"
)

// Minimal opening book keyed by FEN (piece placement + turn only for simplicity).
var miniBook = map[string][]string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w": {"e2e4", "d2d4", "c2c4", "g1f3"},
}

// BookMove returns a legal book move if available.
func (r *RuleEngine) BookMove() *SimpleMove {
	fen := r.Board.ToFEN(r.State)
	key := keyFromFEN(fen)
	cands, ok := miniBook[key]
	if !ok || len(cands) == 0 {
		return nil
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Shuffle(len(cands), func(i, j int) { cands[i], cands[j] = cands[j], cands[i] })

	for _, mv := range cands {
		if len(mv) < 4 {
			continue
		}
		from := parseCoord(mv[:2])
		to := parseCoord(mv[2:4])
		if from == nil || to == nil {
			continue
		}
		promo := rune(0)
		if len(mv) == 5 {
			promo = rune(mv[4])
		}
		if r.IsLegalMove(*from, *to) {
			return &SimpleMove{From: *from, To: *to, Promo: promo}
		}
	}
	return nil
}

func keyFromFEN(f string) string {
	parts := []rune(f)
	spaces := 0
	for i, c := range parts {
		if c == ' ' {
			spaces++
			if spaces == 2 {
				return string(parts[:i])
			}
		}
	}
	return f
}

func parseCoord(s string) *address.Addr {
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
