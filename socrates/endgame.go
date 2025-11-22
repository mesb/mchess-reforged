// --- socrates/endgame.go ---

// Provides checkmate and stalemate detection based on current game state.
package socrates

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

// hasAnyLegalMove checks if the current player has at least one legal move.
func (r *RuleEngine) hasAnyLegalMove() bool {
	for from, p := range r.Board.All() {
		if p.Color() != r.Turn {
			continue
		}
		// Fix: Pass r.State to ValidMoves
		moves := p.ValidMoves(from, r.Board, r.State)
		for _, to := range moves {
			if !r.WouldBeInCheck(from, to) {
				return true
			}
		}
	}
	return false
}
