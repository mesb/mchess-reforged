// --- socrates/rules.go ---

package socrates

import (
	"strings"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pieces"
)

type RuleEngine struct {
	Board *board.Board
	State *board.GameState
	Turn  int
	Log   *Log

	hash        uint64
	hashHistory []uint64
}

func New(b *board.Board) *RuleEngine {
	r := &RuleEngine{
		Board: b,
		State: board.NewGameState(),
		Turn:  pieces.WHITE,
		Log:   &Log{},
	}
	r.resetHashHistory()
	return r
}

// MakeMove executes a move. promoChar is optional (e.g., 'q', 'n').
func (r *RuleEngine) MakeMove(from, to address.Addr, promoChar rune) bool {
	stateBefore := snapshotState(r.State)

	if !r.IsLegalMove(from, to) {
		return false
	}

	moving := r.Board.PieceAt(from)
	target := r.Board.PieceAt(to)
	targetPos := to
	rookMove := (*CastleMove)(nil)

	// 50-Move Rule Tracking
	_, isPawn := moving.(*pieces.Pawn)
	isCapture := target != nil // [FIX] target is now used here

	// --- Logic: En Passant Capture ---
	if isPawn {
		if ep := r.State.GetEnPassant(); ep != nil && to.Equals(*ep) {
			captureRankDir := -1
			if moving.Color() == pieces.BLACK {
				captureRankDir = 1
			}
			if victimPos, ok := to.Shift(captureRankDir, 0); ok {
				target = r.Board.PieceAt(victimPos) // [FIX] target updated
				r.Board.Clear(victimPos)
				isCapture = true
				targetPos = victimPos
			}
		}
	}

	// --- Logic: Castling Execution ---
	if _, isKing := moving.(*pieces.King); isKing {
		df := int(to.File) - int(from.File)
		if df == 2 || df == -2 {
			rank := from.Rank
			var rookFrom, rookTo address.Addr
			if df == 2 { // Kingside
				rookFrom = address.MakeAddr(rank, 7)
				rookTo = address.MakeAddr(rank, 5)
			} else { // Queenside
				rookFrom = address.MakeAddr(rank, 0)
				rookTo = address.MakeAddr(rank, 3)
			}
			// Move Rook
			rook := r.Board.PieceAt(rookFrom)
			r.Board.SetPiece(rookTo, rook)
			r.Board.Clear(rookFrom)
			rookMove = &CastleMove{From: rookFrom, To: rookTo}
		}
	}

	// Record Move (Must happen before board update destroys 'from' state)
	if r.Log != nil {
		r.Log.Record(Move{
			From:      from,
			To:        to,
			Piece:     moving,
			Target:    target,
			TargetPos: &targetPos,
			RookMove:  rookMove,
			PrevState: stateBefore,
		})
	}

	// Apply Main Move
	r.Board.SetPiece(to, moving)
	r.Board.Clear(from)

	// --- State Updates ---
	r.updateEnPassantState(moving, from, to)
	r.updateCastlingRights(moving, from, target, targetPos)
	r.State.IncrementClock(isPawn, isCapture)

	// --- Logic: Promotion ---
	if isPawn {
		rank := to.Rank
		if (moving.Color() == pieces.WHITE && rank == 7) || (moving.Color() == pieces.BLACK && rank == 0) {
			// Promote based on input char, default to Queen
			newPiece := pieces.FromChar(promoChar, moving.Color())
			r.Board.SetPiece(to, newPiece)
		}
	}

	r.Turn = 1 - r.Turn
	r.State.Turn = r.Turn
	if moving.Color() == pieces.BLACK {
		r.State.FullmoveNumber++
	}

	r.refreshHashHistory()

	return true
}

func (r *RuleEngine) updateEnPassantState(p pieces.Piece, from, to address.Addr) {
	r.State.SetEnPassant(nil)
	if _, ok := p.(*pieces.Pawn); ok {
		dr, _ := address.Delta(from, to)
		if dr == 2 || dr == -2 {
			midRank := (int(from.Rank) + int(to.Rank)) / 2
			target := address.MakeAddr(address.Rank(midRank), from.File)
			r.State.SetEnPassant(&target)
		}
	}
}

func (r *RuleEngine) updateCastlingRights(p pieces.Piece, from address.Addr, captured pieces.Piece, capturePos address.Addr) {
	if _, ok := p.(*pieces.King); ok {
		r.State.RevokeCastling(p.Color())
		return
	}
	if _, ok := p.(*pieces.Rook); ok {
		if from.Equals(address.MakeAddr(0, 7)) {
			r.State.RevokeSide("K")
		}
		if from.Equals(address.MakeAddr(0, 0)) {
			r.State.RevokeSide("Q")
		}
		if from.Equals(address.MakeAddr(7, 7)) {
			r.State.RevokeSide("k")
		}
		if from.Equals(address.MakeAddr(7, 0)) {
			r.State.RevokeSide("q")
		}
	}
	if captured != nil {
		if _, ok := captured.(*pieces.Rook); ok {
			// If you captured an enemy rook on its home square, revoke that side.
			if capturePos.Equals(address.MakeAddr(0, 7)) {
				r.State.RevokeSide("K")
			}
			if capturePos.Equals(address.MakeAddr(0, 0)) {
				r.State.RevokeSide("Q")
			}
			if capturePos.Equals(address.MakeAddr(7, 7)) {
				r.State.RevokeSide("k")
			}
			if capturePos.Equals(address.MakeAddr(7, 0)) {
				r.State.RevokeSide("q")
			}
		}
	}
}

func (r *RuleEngine) IsLegalMove(from, to address.Addr) bool {
	piece := r.Board.PieceAt(from)
	if piece == nil || piece.Color() != r.Turn {
		return false
	}

	if _, isKing := piece.(*pieces.King); isKing {
		df := int(to.File) - int(from.File)
		if from.Rank == to.Rank && (df == 2 || df == -2) {
			if r.canCastle(piece.Color(), df == 2, from, to) {
				return true
			}
			return false
		}
	}

	legalMoves := piece.ValidMoves(from, r.Board, r.State)
	for _, move := range legalMoves {
		if move.Equals(to) {
			if _, isKing := piece.(*pieces.King); isKing {
				df := int(to.File) - int(from.File)
				if df == 2 || df == -2 {
					if r.IsInCheck(r.Turn) {
						return false
					}
					midFile := int(from.File) + (df / 2)
					midSquare := address.MakeAddr(from.Rank, address.File(midFile))
					if r.WouldBeInCheck(from, midSquare) {
						return false
					}
				}
			}
			return !r.WouldBeInCheck(from, to)
		}
	}
	return false
}

func (r *RuleEngine) GetTurn() int { return r.Turn }

func (r *RuleEngine) canCastle(color int, kingSide bool, from, to address.Addr) bool {
	rights := r.State.CastlingRights
	if kingSide {
		if (color == pieces.WHITE && !strings.Contains(rights, "K")) ||
			(color == pieces.BLACK && !strings.Contains(rights, "k")) {
			return false
		}
	} else {
		if (color == pieces.WHITE && !strings.Contains(rights, "Q")) ||
			(color == pieces.BLACK && !strings.Contains(rights, "q")) {
			return false
		}
	}

	rookFile := 7
	if !kingSide {
		rookFile = 0
	}
	rookPos := address.MakeAddr(from.Rank, address.File(rookFile))
	rook := r.Board.PieceAt(rookPos)
	if rook == nil {
		return false
	}
	if _, ok := rook.(*pieces.Rook); !ok || rook.Color() != color {
		return false
	}

	// Squares between king and rook must be empty.
	step := 1
	if !kingSide {
		step = -1
	}
	for f := int(from.File) + step; f != rookFile; f += step {
		pos := address.MakeAddr(from.Rank, address.File(f))
		if !r.Board.IsEmpty(pos) {
			return false
		}
	}

	if r.IsInCheck(color) {
		return false
	}

	// King cannot pass through or land on attacked squares.
	midFile := int(from.File) + step
	midSquare := address.MakeAddr(from.Rank, address.File(midFile))
	if r.WouldBeInCheck(from, midSquare) || r.WouldBeInCheck(from, to) {
		return false
	}

	return true
}

func (r *RuleEngine) IsInCheck(color int) bool {
	kingPos := findKing(r.Board, color)
	if kingPos == nil {
		return false
	}
	k := *kingPos
	enemyColor := 1 - color

	// Knights
	knightDeltas := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}
	for _, d := range knightDeltas {
		if pos, ok := k.Shift(d[0], d[1]); ok {
			p := r.Board.PieceAt(pos)
			if p != nil && p.Color() == enemyColor {
				if _, isKnight := p.(*pieces.Knight); isKnight {
					return true
				}
			}
		}
	}
	// Pawns
	pawnDir := 1
	if color == pieces.BLACK {
		pawnDir = -1
	}
	for _, dx := range []int{-1, 1} {
		if pos, ok := k.Shift(pawnDir, dx); ok {
			p := r.Board.PieceAt(pos)
			if p != nil && p.Color() == enemyColor {
				if _, isPawn := p.(*pieces.Pawn); isPawn {
					return true
				}
			}
		}
	}
	// Sliders
	if scanDirs(r.Board, k, [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}, enemyColor, true, false) {
		return true
	}
	if scanDirs(r.Board, k, [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}, enemyColor, false, true) {
		return true
	}
	return false
}

func scanDirs(b *board.Board, start address.Addr, dirs [][2]int, enemyColor int, checkRook, checkBishop bool) bool {
	for _, d := range dirs {
		for i := 1; i < 8; i++ {
			pos, ok := start.Shift(d[0]*i, d[1]*i)
			if !ok {
				break
			}
			p := b.PieceAt(pos)
			if p != nil {
				if p.Color() == enemyColor {
					switch p.(type) {
					case *pieces.Queen:
						return true
					case *pieces.Rook:
						if checkRook {
							return true
						}
					case *pieces.Bishop:
						if checkBishop {
							return true
						}
					}
				}
				break
			}
		}
	}
	return false
}

func (r *RuleEngine) WouldBeInCheck(from, to address.Addr) bool {
	moving := r.Board.PieceAt(from)
	captured := r.Board.PieceAt(to)

	// Handle en passant when simulating checks: remove the pawn that would be captured.
	if pawn, ok := moving.(*pieces.Pawn); ok {
		if ep := r.State.GetEnPassant(); ep != nil && to.Equals(*ep) && captured == nil && from.File != to.File {
			captureRankDir := -1
			if pawn.Color() == pieces.BLACK {
				captureRankDir = 1
			}
			if victimPos, ok := to.Shift(captureRankDir, 0); ok {
				captured = r.Board.PieceAt(victimPos)
				r.Board.Clear(victimPos)
				defer r.Board.SetPiece(victimPos, captured)
			}
		}
	}

	r.Board.SetPiece(to, moving)
	r.Board.Clear(from)
	inCheck := r.IsInCheck(moving.Color())
	r.Board.SetPiece(from, moving)
	r.Board.SetPiece(to, captured)
	return inCheck
}

func findKing(b *board.Board, color int) *address.Addr {
	return b.FindKing(color)
}

func (r *RuleEngine) resetHashHistory() {
	r.hash = computeHash(r.Board, r.State, r.Turn)
	r.hashHistory = []uint64{r.hash}
}

func (r *RuleEngine) ResetHashHistory() {
	r.resetHashHistory()
}

func (r *RuleEngine) refreshHashHistory() {
	r.hash = computeHash(r.Board, r.State, r.Turn)
	r.hashHistory = append(r.hashHistory, r.hash)
}

func (r *RuleEngine) isRepetition() bool {
	if len(r.hashHistory) < 3 {
		return false
	}
	current := r.hashHistory[len(r.hashHistory)-1]
	count := 0
	for _, h := range r.hashHistory {
		if h == current {
			count++
		}
	}
	return count >= 3
}
