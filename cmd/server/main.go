// --- cmd/server/main.go ---

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq" // Postgres Driver

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/pgn"
	"github.com/mesb/mchess/pieces"
	ws "github.com/mesb/mchess/server"
	"github.com/mesb/mchess/shell"
	"github.com/mesb/mchess/socrates"
)

// --- Interfaces & DTOs ---

type CreateGameResponse struct {
	ID string `json:"game_id"`
}

type GameStateResponse struct {
	ID       string     `json:"game_id"`
	Turn     string     `json:"turn"`
	IsOver   bool       `json:"is_game_over"`
	Status   string     `json:"status"`
	BoardFEN string     `json:"board_fen"`
	Board    [][]string `json:"board"`
}

type MoveRequest struct {
	Move string `json:"move"`
}

type GameStore interface {
	Create() (string, error)
	Get(id string) (*shell.GameSession, error)
	Save(id string, session *shell.GameSession) error
}

// --- Implementation 1: InMemory (Development) ---

type InMemoryStore struct {
	sync.RWMutex
	games map[string]*shell.GameSession
}

func NewMemoryStore() *InMemoryStore {
	return &InMemoryStore{games: make(map[string]*shell.GameSession)}
}

func (s *InMemoryStore) Create() (string, error) {
	s.Lock()
	defer s.Unlock()
	id := fmt.Sprintf("game_%d", time.Now().UnixNano())
	s.games[id] = shell.NewSession(nil)
	return id, nil
}

func (s *InMemoryStore) Get(id string) (*shell.GameSession, error) {
	s.RLock()
	defer s.RUnlock()
	g, ok := s.games[id]
	if !ok {
		return nil, fmt.Errorf("game not found")
	}
	return g, nil
}

func (s *InMemoryStore) Save(id string, session *shell.GameSession) error {
	// In-memory holds the pointer, so logic automatically "saves" updates.
	return nil
}

// --- Implementation 2: Postgres (Production) ---

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := ensureSchema(db); err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

// ensureSchema creates the games table if needed and adds new columns when upgrading.
func ensureSchema(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS games (id TEXT PRIMARY KEY, fen TEXT, pgn TEXT);`); err != nil {
		return err
	}
	// Migrations for older deployments missing fen or pgn columns.
	if _, err := db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS fen TEXT;`); err != nil {
		return err
	}
	if _, err := db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS pgn TEXT;`); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) Create() (string, error) {
	id := fmt.Sprintf("game_%d", time.Now().UnixNano())
	session := shell.NewSession(nil)
	initialFEN := session.Engine.Board.ToFEN(session.Engine.State)
	_, err := s.db.Exec("INSERT INTO games (id, fen, pgn) VALUES ($1, $2, '')", id, initialFEN)
	return id, err
}

func (s *PostgresStore) Get(id string) (*shell.GameSession, error) {
	var fenData, pgnData sql.NullString
	err := s.db.QueryRow("SELECT fen, pgn FROM games WHERE id = $1", id).Scan(&fenData, &pgnData)
	if err != nil {
		return nil, err
	}

	if fenData.Valid {
		board, state, err := board.FromFEN(fenData.String)
		if err == nil {
			session := shell.NewSession(nil)
			session.Engine.Board = board
			session.Engine.State = state
			session.Engine.Turn = state.Turn
			session.Engine.ResetHashHistory()
			return session, nil
		}
		// fall back to PGN replay if FEN invalid
	}

	session := shell.NewSession(nil)
	if pgnData.Valid && pgnData.String != "" {
		if err := pgn.Import(session.Engine, pgnData.String); err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (s *PostgresStore) Save(id string, session *shell.GameSession) error {
	// Serialize state to PGN
	data := pgn.Export(session.Log)
	fen := session.Engine.Board.ToFEN(session.Engine.State)
	_, err := s.db.Exec("UPDATE games SET fen = $1, pgn = $2 WHERE id = $3", fen, data, id)
	return err
}

// --- Main Server Logic ---

func main() {
	var store GameStore
	var err error

	// Elegant Switch: If DB env is present, use Postgres. Else Memory.
	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		log.Println("🔌 Connecting to PostgreSQL...")
		store, err = NewPostgresStore(dsn)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("⚠️  No Database URL found. Using In-Memory Store.")
		store = NewMemoryStore()
	}

	hub := ws.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		gameID := r.URL.Query().Get("game_id")
		if gameID == "" {
			http.Error(w, "missing game_id", http.StatusBadRequest)
			return
		}
		if _, err := store.Get(gameID); err != nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		ws.ServeWS(hub, gameID, w, r)
	})

	http.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // FIXED
			return
		}
		id, err := store.Create()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // FIXED
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateGameResponse{ID: id})
	})

	http.HandleFunc("/games/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.Error(w, "Invalid path", http.StatusBadRequest) // FIXED
			return
		}
		gameID := parts[1]

		session, err := store.Get(gameID)
		if err != nil {
			http.Error(w, "Game not found", http.StatusNotFound) // FIXED
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet && len(parts) == 2 {
			handleGetState(w, session, gameID)
		} else if r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "move" {
			handleMove(w, r, session, store, gameID, hub)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // FIXED
		}
	})

	log.Println("♟️  MCHESS API listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleGetState(w http.ResponseWriter, s *shell.GameSession, id string) {
	s.Mu.RLock()
	resp := snapshotStateResponse(s, id)
	s.Mu.RUnlock()
	json.NewEncoder(w).Encode(resp)
}

func handleMove(w http.ResponseWriter, r *http.Request, s *shell.GameSession, store GameStore, id string, hub *ws.Hub) {
	var req MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest) // FIXED
		return
	}

	from, to, promo, err := socrates.ParseMove(req.Move)
	if err != nil {
		http.Error(w, "Invalid notation: "+err.Error(), http.StatusBadRequest) // FIXED
		return
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	if !s.Engine.MakeMove(*from, *to, promo) {
		http.Error(w, "Illegal move", http.StatusConflict) // FIXED
		return
	}

	// Persist state after move!
	if err := store.Save(id, s); err != nil {
		log.Printf("Failed to save game: %v", err)
	}

	resp := snapshotStateResponse(s, id)
	if hub != nil {
		if payload, err := json.Marshal(resp); err == nil {
			hub.BroadcastTo(id, payload)
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func snapshotStateResponse(s *shell.GameSession, id string) GameStateResponse {
	status := "Active"
	isOver := false
	if s.Engine.IsCheckmate() {
		status = "Checkmate"
		isOver = true
	} else if s.Engine.IsStalemate() {
		status = "Stalemate"
		isOver = true
	} else if s.Engine.IsInCheck(s.Engine.Turn) {
		status = "Check"
	}
	turn := "White"
	if s.Engine.Turn == 1 {
		turn = "Black"
	}

	fen := s.Engine.Board.ToFEN(s.Engine.State)

	return GameStateResponse{
		ID:       id,
		Turn:     turn,
		IsOver:   isOver,
		Status:   status,
		BoardFEN: fen,
		Board:    materializeBoard(s.Engine.Board),
	}
}

func materializeBoard(b *board.Board) [][]string {
	out := make([][]string, 8)
	for r := 7; r >= 0; r-- {
		row := make([]string, 8)
		for f := 0; f < 8; f++ {
			a := address.MakeAddr(address.Rank(r), address.File(f))
			p := b.PieceAt(a)
			if p == nil {
				row[f] = "--"
				continue
			}
			row[f] = pieceSymbol(p)
		}
		out[7-r] = row
	}
	return out
}

func pieceSymbol(p pieces.Piece) string {
	switch t := p.(type) {
	case *pieces.Pawn:
		if t.Color() == pieces.WHITE {
			return "P"
		}
		return "p"
	case *pieces.Rook:
		if t.Color() == pieces.WHITE {
			return "R"
		}
		return "r"
	case *pieces.Knight:
		if t.Color() == pieces.WHITE {
			return "N"
		}
		return "n"
	case *pieces.Bishop:
		if t.Color() == pieces.WHITE {
			return "B"
		}
		return "b"
	case *pieces.Queen:
		if t.Color() == pieces.WHITE {
			return "Q"
		}
		return "q"
	case *pieces.King:
		if t.Color() == pieces.WHITE {
			return "K"
		}
		return "k"
	default:
		_ = t
		return ""
	}
}
