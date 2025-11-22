// --- socrates/eval.go ---

package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

// Score constants (centipawns)
const (
	ValuePawn   = 100
	ValueKnight = 320
	ValueBishop = 330
	ValueRook   = 500
	ValueQueen  = 900
	ValueKing   = 20000

	MobilityWeight = 2
)

// Evaluate returns the score of the board from White's perspective.
// Positive = White advantage, Negative = Black advantage.
func Evaluate(b *board.Board) int {
	return EvaluatePosition(b, nil)
}

// EvaluatePosition includes optional state for mobility-aware scoring.
func EvaluatePosition(b *board.Board, state *board.GameState) int {
	score := 0
	b.ForEachPiece(func(sq address.Addr, p pieces.Piece) {
		val := 0
		idx := sq.Index() // 0 = a1, 63 = h8

		// 1. Material & Position Score
		switch p.(type) {
		case *pieces.Pawn:
			val = ValuePawn + pstPawn[mirror(p.Color(), idx)]
		case *pieces.Knight:
			val = ValueKnight + pstKnight[mirror(p.Color(), idx)]
		case *pieces.Bishop:
			val = ValueBishop + pstBishop[mirror(p.Color(), idx)]
		case *pieces.Rook:
			val = ValueRook + pstRook[mirror(p.Color(), idx)]
		case *pieces.Queen:
			val = ValueQueen + pstQueen[mirror(p.Color(), idx)]
		case *pieces.King:
			val = ValueKing + pstKingMid[mirror(p.Color(), idx)]
		}

		// Mobility bonus encourages development and activity.
		if state != nil {
			m := p.ValidMoves(sq, b, state)
			val += len(m) * MobilityWeight
		}

		// 2. Accumulate
		if p.Color() == pieces.WHITE {
			score += val
		} else {
			score -= val
		}
	})
	return score
}

// mirror flips the index for Black so we can use the same PST array.
// White views board from rank 1->8. Black views it effectively 8->1.
func mirror(color, index int) int {
	if color == pieces.WHITE {
		return index // White uses table as-is
	}
	// Flip vertical: rank 0 becomes rank 7, etc.
	// XOR 56 (binary 111000) flips the rank bits (0-7 <-> 56-63)
	return index ^ 56
}

// --- Piece-Square Tables (PST) ---
// Defined from Rank 1 (bottom) to Rank 8 (top), a-h.
// These encourage pieces to move to active squares.

var pstPawn = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var pstKnight = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var pstBishop = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var pstRook = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var pstQueen = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var pstKingMid = [64]int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}
