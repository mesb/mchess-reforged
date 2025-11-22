// --- cmd/server/main.go ---

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mesb/mchess/shell"
	"github.com/mesb/mchess/socrates"
)

// --- 1. Data Transfer Objects (JSON Contracts) ---

type CreateGameResponse struct {
	ID string `json:"game_id"`
}

type GameStateResponse struct {
	ID     string   `json:"game_id"`
	Turn   string   `json:"turn"` // "White" or "Black"
	IsOver bool     `json:"is_game_over"`
	Status string   `json:"status"`    // "Active", "Checkmate", etc.
	Board  []string `json:"board_fen"` // Standard FEN string (encapsulated in array)
}

type MoveRequest struct {
	Move string `json:"move"` // e.g. "e2e4"
}

// --- 2. The Storage Interface (Scalability Layer) ---

type GameStore interface {
	Create() string
	Get(id string) (*shell.GameSession, bool)
}

// InMemoryStore: Simple, fast, thread-safe storage.
type InMemoryStore struct {
	sync.RWMutex
	games map[string]*shell.GameSession
}

func NewMemoryStore() *InMemoryStore {
	return &InMemoryStore{games: make(map[string]*shell.GameSession)}
}

func (s *InMemoryStore) Create() string {
	s.Lock()
	defer s.Unlock()

	// Generate a simple ID (use UUID in production)
	id := fmt.Sprintf("game_%d", time.Now().UnixNano())

	// Initialize a headless session (nil renderer)
	session := shell.NewSession(nil)
	s.games[id] = session
	return id
}

func (s *InMemoryStore) Get(id string) (*shell.GameSession, bool) {
	s.RLock()
	defer s.RUnlock()
	g, ok := s.games[id]
	return g, ok
}

// --- 3. The Server Implementation ---

func main() {
	store := NewMemoryStore()

	// POST /games -> Create a new game
	http.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := store.Create()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateGameResponse{ID: id})
	})

	// GET /games/{id} -> Get game state
	// POST /games/{id}/move -> Make a move
	http.HandleFunc("/games/", func(w http.ResponseWriter, r *http.Request) {
		// URL pattern: /games/{id} or /games/{id}/move
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		gameID := parts[1]

		session, exists := store.Get(gameID)
		if !exists {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet && len(parts) == 2 {
			handleGetState(w, session, gameID)
		} else if r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "move" {
			handleMove(w, r, session)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("♟️  MCHESS API listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- Helpers ---

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

	// Generate Standard FEN string using the board method
	fen := s.Engine.Board.ToFEN(s.Engine.State)

	resp := GameStateResponse{
		ID:     id,
		Turn:   turn,
		IsOver: isOver,
		Status: status,
		Board:  []string{fen}, // Returns actual board state
	}
	json.NewEncoder(w).Encode(resp)
}

func handleMove(w http.ResponseWriter, r *http.Request, s *shell.GameSession) {
	var req MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	from, to, promo, err := socrates.ParseMove(req.Move)
	if err != nil {
		http.Error(w, "Invalid notation: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !s.Engine.MakeMove(*from, *to, promo) {
		http.Error(w, "Illegal move", http.StatusConflict)
		return
	}

	// Return updated state
	handleGetState(w, s, "current")
}
