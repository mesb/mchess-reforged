// --- socrates/endgame.go ---

package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

// IsCheckmate returns true if the current player is in check and has no legal escape.
func (r *RuleEngine) IsCheckmate() bool {
	if !r.IsInCheck(r.Turn) {
		return false
	}
	return !r.hasAnyLegalMove()
}

// IsStalemate returns true if the player is NOT in check, but has no legal move.
func (r *RuleEngine) IsStalemate() bool {
	if r.IsInCheck(r.Turn) {
		return false
	}
	return !r.hasAnyLegalMove()
}

// IsFiftyMoveRule returns true if 50 moves have passed without capture or pawn advance.
func (r *RuleEngine) IsFiftyMoveRule() bool {
	return r.State.HalfmoveClock >= 100 // 50 full moves = 100 half moves
}

// IsInsufficientMaterial returns true if neither side has enough material to mate.
// Covers: K vs K, K+B vs K, K+N vs K.
func (r *RuleEngine) IsInsufficientMaterial() bool {
	allPieces := r.Board.All()
	if len(allPieces) == 2 {
		return true // King vs King
	}

	if len(allPieces) == 3 {
		for _, p := range allPieces {
			switch p.(type) {
			case *pieces.Bishop, *pieces.Knight:
				return true // King + Minor Piece vs King
			}
		}
	}

	// (Further checks for K+B vs K+B on same color could go here)
	return false
}

func (r *RuleEngine) hasAnyLegalMove() bool {
	found := false
	r.Board.ForEachPiece(func(from address.Addr, p pieces.Piece) {
		if found || p.Color() != r.Turn {
			return
		}
		moves := p.ValidMoves(from, r.Board, r.State)
		for _, to := range moves {
			if r.IsLegalMove(from, to) {
				found = true
				return
			}
		}
	})
	return found
}
