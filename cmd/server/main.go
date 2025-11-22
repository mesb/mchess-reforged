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

	"github.com/mesb/mchess/pgn"
	"github.com/mesb/mchess/shell"
	"github.com/mesb/mchess/socrates"
)

// --- Interfaces & DTOs ---

type CreateGameResponse struct {
	ID string `json:"game_id"`
}

type GameStateResponse struct {
	ID     string   `json:"game_id"`
	Turn   string   `json:"turn"`
	IsOver bool     `json:"is_game_over"`
	Status string   `json:"status"`
	Board  []string `json:"board_fen"`
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
	// Ensure table exists
	query := `CREATE TABLE IF NOT EXISTS games (id TEXT PRIMARY KEY, pgn TEXT);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Create() (string, error) {
	id := fmt.Sprintf("game_%d", time.Now().UnixNano())
	// Create empty game
	_, err := s.db.Exec("INSERT INTO games (id, pgn) VALUES ($1, '')", id)
	return id, err
}

func (s *PostgresStore) Get(id string) (*shell.GameSession, error) {
	var pgnData string
	err := s.db.QueryRow("SELECT pgn FROM games WHERE id = $1", id).Scan(&pgnData)
	if err != nil {
		return nil, err
	}

	// Rehydrate: Create fresh session -> Replay PGN
	session := shell.NewSession(nil)
	if pgnData != "" {
		if err := pgn.Import(session.Engine, pgnData); err != nil {
			return nil, err
		}
	}
	return session, nil
}

func (s *PostgresStore) Save(id string, session *shell.GameSession) error {
	// Serialize state to PGN
	data := pgn.Export(session.Log)
	_, err := s.db.Exec("UPDATE games SET pgn = $1 WHERE id = $2", data, id)
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
			handleMove(w, r, session, store, gameID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // FIXED
		}
	})

	log.Println("♟️  MCHESS API listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleGetState(w http.ResponseWriter, s *shell.GameSession, id string) {
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

	json.NewEncoder(w).Encode(GameStateResponse{
		ID: id, Turn: turn, IsOver: isOver, Status: status, Board: []string{fen},
	})
}

func handleMove(w http.ResponseWriter, r *http.Request, s *shell.GameSession, store GameStore, id string) {
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

	if !s.Engine.MakeMove(*from, *to, promo) {
		http.Error(w, "Illegal move", http.StatusConflict) // FIXED
		return
	}

	// Persist state after move!
	if err := store.Save(id, s); err != nil {
		log.Printf("Failed to save game: %v", err)
	}

	handleGetState(w, s, "current")
}
