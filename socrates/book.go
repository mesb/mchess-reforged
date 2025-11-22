package socrates

import (
	"math/rand"
	"time"

	"github.com/mesb/mchess/address"
)

// Minimal opening book keyed by FEN (piece placement + turn only for simplicity).
var miniBook = map[string][]string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w":       {"e2e4", "d2d4", "c2c4", "g1f3"},
	"rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w":   {"g1f3", "d2d4"},
	"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b":     {"c7c5", "e7e5", "e7e6", "c7c6"},
	"rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b":     {"d7d5", "g8f6", "e7e6"},
	"rnbqkbnr/pppppp1p/6p1/8/4P3/5N2/PPPP1PPP/RNBQKB1R b": {"d7d6", "c7c5"},
	"rnbqkbnr/pp1ppppp/2p5/8/4P3/5N2/PPPP1PPP/RNBQKB1R b": {"d7d5", "g8f6"},
	"rnbqkbnr/pppppppp/8/8/4P3/4P3/PPPP2PP/RNBQKBNR b":    {"d7d5", "g8f6"},
	"rnbqkbnr/ppp1pppp/3p4/8/3PP3/8/PPP2PPP/RNBQKBNR b":   {"g8f6", "c7c5"},
	"rnbqkbnr/pppppppp/8/8/2P5/8/PP1PPPPP/RNBQKBNR b":     {"e7e5", "c7c5"},
	"rnbqkbnr/pppppppp/8/8/4P3/8/PPPPQPPP/RNB1KBNR b":     {"d7d5", "c7c5", "g8f6"},
	"rnbqkbnr/pp1ppppp/2p5/8/2P5/8/PP1PPPPP/RNBQKBNR w":   {"d2d4", "g1f3"},
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
