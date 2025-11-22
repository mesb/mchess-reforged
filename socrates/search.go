// --- socrates/search.go ---

package socrates

import (
	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/pieces"
)

const (
	// MaxScore/MinScore act as search infinities.
	MaxScore = 32000
	MinScore = -32000
	// MateScore bounds mate scores; distance-to-mate is expressed as MateScore - ply.
	MateScore = 30000
	// EvalClamp keeps quiescence from hallucinating mates.
	EvalClamp = 29000
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

	// No legal moves: return mate/stalemate immediately.
	if len(moves) == 0 {
		score := 0
		if r.IsInCheck(r.Turn) {
			score = -MateScore
		}
		return SearchResult{Score: score, Nodes: 1}
	}

	totalNodes := 0

	for _, m := range moves {
		r.MakeMove(m.From, m.To, m.Promo)

		score, visited := r.negamax(depth-1, 1, -beta, -alpha)
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
func (r *RuleEngine) negamax(depth, ply, alpha, beta int) (int, int) {
	nodes := 1 // count this node
	alphaOrig := alpha

	if entry, ok := r.tt[r.hash]; ok && entry.depth >= depth {
		score := fromTTScore(entry.score, ply)
		switch entry.flag {
		case ttExact:
			return score, nodes
		case ttLower:
			if score >= beta {
				return score, nodes
			}
		case ttUpper:
			if score <= alpha {
				return score, nodes
			}
		}
	}

	if r.isRepetition() {
		return 0, nodes
	}

	// 1. Leaf Node: Return Static Evaluation
	if depth == 0 {
		score, qNodes := r.quiesce(alpha, beta)
		return score, nodes + qNodes
	}

	// 2. Generate Moves
	moves := r.GenerateLegalMoves()

	// 3. Game Over Detection
	if len(moves) == 0 {
		if r.IsInCheck(r.Turn) {
			return -MateScore + ply, nodes
		}
		return 0, nodes // Stalemate
	}

	// 4. Recursion
	for _, m := range moves {
		r.MakeMove(m.From, m.To, m.Promo)
		score, childNodes := r.negamax(depth-1, ply+1, -beta, -alpha)
		nodes += childNodes
		r.UndoMove()

		score = -score

		if score >= beta {
			r.storeTT(r.hash, depth, toTTScore(score, ply), ttLower, m)
			return beta, nodes // Pruning
		}
		if score > alpha {
			alpha = score
		}
	}
	r.storeTT(r.hash, depth, toTTScore(alpha, ply), flagFrom(alpha, beta, alphaOrig), SimpleMove{})
	return alpha, nodes
}

// evaluateRelative adapts the static evaluation to the current turn.
// Evaluate() returns White - Black.
// If it's Black's turn, we want Black - White (which is -(White - Black)).
func (r *RuleEngine) evaluateRelative() int {
	score := EvaluatePosition(r.Board, r.State)
	if r.Turn == pieces.BLACK {
		return -score
	}
	return score
}

// quiesce searches capture sequences to reduce horizon effects.
func (r *RuleEngine) quiesce(alpha, beta int) (int, int) {
	nodes := 1
	score := r.evaluateRelative()
	if score > EvalClamp {
		score = EvalClamp
	}
	if score < -EvalClamp {
		score = -EvalClamp
	}

	if score >= beta {
		return beta, nodes
	}
	if score > alpha {
		alpha = score
	}

	for _, m := range r.GenerateCaptureMoves() {
		// Only explore captures or promotions (as noisy moves).
		r.MakeMove(m.From, m.To, m.Promo)
		childScore, childNodes := r.quiesce(-beta, -alpha)
		nodes += childNodes
		r.UndoMove()

		childScore = -childScore

		if childScore >= beta {
			return beta, nodes
		}
		if childScore > alpha {
			alpha = childScore
		}
	}

	return alpha, nodes
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
				target := r.Board.PieceAt(to)
				isPawn := false
				if _, ok := p.(*pieces.Pawn); ok {
					isPawn = true
				}
				isEnPassantCapture := target == nil && isPawn && from.File != to.File
				capturing := target != nil || isEnPassantCapture

				if isPromo(p, to) {
					move := SimpleMove{from, to, 'q'}
					if capturing {
						captures = append(captures, move)
					} else {
						quiets = append(quiets, move)
					}
				} else {
					move := SimpleMove{from, to, 0}
					if capturing {
						captures = append(captures, move)
					} else {
						quiets = append(quiets, move)
					}
				}
			}
		}
	})
	moves = append(moves, captures...)
	moves = append(moves, quiets...)
	return moves
}

// GenerateCaptureMoves returns only captures/promotions to speed quiescence.
func (r *RuleEngine) GenerateCaptureMoves() []SimpleMove {
	moves := make([]SimpleMove, 0, 32)
	r.Board.ForEachPiece(func(from address.Addr, p pieces.Piece) {
		if p.Color() != r.Turn {
			return
		}
		candidates := p.ValidMoves(from, r.Board, r.State)
		for _, to := range candidates {
			if r.IsLegalMove(from, to) && (r.isCapture(SimpleMove{From: from, To: to}) || isPromo(p, to)) {
				promo := rune(0)
				if isPromo(p, to) {
					promo = 'q'
				}
				moves = append(moves, SimpleMove{From: from, To: to, Promo: promo})
			}
		}
	})
	return moves
}

func isPromo(p pieces.Piece, to address.Addr) bool {
	if _, ok := p.(*pieces.Pawn); !ok {
		return false
	}
	return to.Rank == 0 || to.Rank == 7
}

func (r *RuleEngine) isCapture(m SimpleMove) bool {
	if !r.Board.IsEmpty(m.To) {
		return true
	}
	// En passant detection: pawn moving diagonally into empty square
	if fromPiece := r.Board.PieceAt(m.From); fromPiece != nil {
		if _, ok := fromPiece.(*pieces.Pawn); ok && m.From.File != m.To.File {
			return true
		}
	}
	return false
}

func (r *RuleEngine) storeTT(hash uint64, depth int, score int, flag int, move SimpleMove) {
	r.tt[hash] = ttEntry{
		hash:  hash,
		depth: depth,
		score: score,
		flag:  flag,
		move:  move,
	}
}

func flagFrom(score, beta, alphaOrig int) int {
	if score <= alphaOrig {
		return ttUpper
	}
	if score >= beta {
		return ttLower
	}
	return ttExact
}
