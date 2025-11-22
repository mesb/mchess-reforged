package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleGetStateAndMove(t *testing.T) {
	store := NewMemoryStore()
	gameID, err := store.Create()
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	session, err := store.Get(gameID)
	if err != nil {
		t.Fatalf("get game: %v", err)
	}

	// Initial state
	w := httptest.NewRecorder()
	handleGetState(w, session, gameID)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	var stateResp GameStateResponse
	if err := json.NewDecoder(w.Body).Decode(&stateResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stateResp.BoardFEN == "" || len(stateResp.Board) != 8 {
		t.Fatalf("bad initial board response: %+v", stateResp)
	}

	// Make a legal move
	moveBody := bytes.NewBufferString(`{"move":"e2e4"}`)
	w2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/games/"+gameID+"/move", moveBody)
	handleMove(w2, req, session, store, gameID)
	if w2.Code != http.StatusOK {
		t.Fatalf("unexpected move status: %d", w2.Code)
	}
	var moveResp GameStateResponse
	if err := json.NewDecoder(w2.Body).Decode(&moveResp); err != nil {
		t.Fatalf("decode move: %v", err)
	}
	if moveResp.ID != gameID {
		t.Fatalf("expected game id %s, got %s", gameID, moveResp.ID)
	}
	if moveResp.Turn != "Black" {
		t.Fatalf("expected Black to move after e2e4, got %s", moveResp.Turn)
	}
	if moveResp.BoardFEN == stateResp.BoardFEN {
		t.Fatalf("board FEN did not change after move")
	}
}
