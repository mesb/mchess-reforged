// --- socrates/search.go ---

package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

const (
	MaxScore  = 30000
	MinScore  = -30000
	MateScore = 20000
)

// SearchResult holds the best move found and its evaluation.
type SearchResult struct {
	From  address.Addr
	To    address.Addr
	Score int
	Nodes int // how many positions were analyzed
}

// Search runs the Alpha-Beta Negamax algorithm to a fixed depth.
func (r *RuleEngine) Search(depth int) SearchResult {
	alpha := MinScore
	beta := MaxScore

	bestMove := SearchResult{Score: MinScore}

	moves := r.GenerateLegalMoves()

	totalNodes := 0

	for _, m := range moves {
		r.MakeMove(m.From, m.To, m.Promo)

		score, visited := r.negamax(depth-1, -beta, -alpha)
		score = -score
		totalNodes += visited + 1

		r.UndoMove()

		if score > alpha {
			alpha = score
			bestMove.From = m.From
			bestMove.To = m.To
			bestMove.Score = score
		}
	}

	bestMove.Nodes = totalNodes
	return bestMove
}

// negamax returns the score relative to the side to move and nodes visited.
func (r *RuleEngine) negamax(depth, alpha, beta int) (int, int) {
	nodes := 1 // count this node

	// 1. Leaf Node: Return Static Evaluation
	if depth == 0 {
		return r.evaluateRelative(), nodes
	}

	// 2. Generate Moves
	moves := r.GenerateLegalMoves()

	// 3. Game Over Detection
	if len(moves) == 0 {
		if r.IsInCheck(r.Turn) {
			return -MateScore + (100 - depth), nodes
		}
		return 0, nodes // Stalemate
	}

	// 4. Recursion
	for _, m := range moves {
		r.MakeMove(m.From, m.To, m.Promo)
		score, childNodes := r.negamax(depth-1, -beta, -alpha)
		nodes += childNodes
		r.UndoMove()

		score = -score

		if score >= beta {
			return beta, nodes // Pruning
		}
		if score > alpha {
			alpha = score
		}
	}
	return alpha, nodes
}

// evaluateRelative adapts the static evaluation to the current turn.
// Evaluate() returns White - Black.
// If it's Black's turn, we want Black - White (which is -(White - Black)).
func (r *RuleEngine) evaluateRelative() int {
	score := Evaluate(r.Board)
	if r.Turn == pieces.BLACK {
		return -score
	}
	return score
}

// SimpleMove is a lightweight struct for move generation.
type SimpleMove struct {
	From, To address.Addr
	Promo    rune
}

// GenerateLegalMoves aggregates all valid moves for the current turn.
func (r *RuleEngine) GenerateLegalMoves() []SimpleMove {
	moves := make([]SimpleMove, 0, 40)

	var captures []SimpleMove
	var quiets []SimpleMove

	r.Board.ForEachPiece(func(from address.Addr, p pieces.Piece) {
		if p.Color() != r.Turn {
			return
		}
		candidates := p.ValidMoves(from, r.Board, r.State)
		for _, to := range candidates {
			if r.IsLegalMove(from, to) {
				if isPromo(p, to) {
					move := SimpleMove{from, to, 'q'}
					if r.Board.IsEmpty(to) {
						quiets = append(quiets, move)
					} else {
						captures = append(captures, move)
					}
				} else {
					move := SimpleMove{from, to, 0}
					if r.Board.IsEmpty(to) {
						quiets = append(quiets, move)
					} else {
						captures = append(captures, move)
					}
				}
			}
		}
	})
	moves = append(moves, captures...)
	moves = append(moves, quiets...)
	return moves
}

func isPromo(p pieces.Piece, to address.Addr) bool {
	if _, ok := p.(*pieces.Pawn); !ok {
		return false
	}
	return to.Rank == 0 || to.Rank == 7
}
