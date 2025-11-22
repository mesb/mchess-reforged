package socrates

// Transposition table entry.
type ttEntry struct {
	hash  uint64
	depth int
	score int
	flag  int
	move  SimpleMove
}

const (
	ttExact = iota
	ttLower
	ttUpper
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
