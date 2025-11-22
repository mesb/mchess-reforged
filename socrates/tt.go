package socrates

// Transposition table entry.
type ttEntry struct {
	hash  uint64
	depth int
	score int
	flag  int
	move  SimpleMove
	gen   int
}

const (
	ttExact = iota
	ttLower
	ttUpper
)

const (
	TTSize = 1 << 20 // 1M entries; ~32-40MB
	TTMask = TTSize - 1
)

func toTTScore(score, ply int) int {
	if score > MateScore-1000 {
		return score + ply
	}
	if score < -MateScore+1000 {
		return score - ply
	}
	return score
}

func fromTTScore(score, ply int) int {
	if score > MateScore-1000 {
		return score - ply
	}
	if score < -MateScore+1000 {
		return score + ply
	}
	return score
}
