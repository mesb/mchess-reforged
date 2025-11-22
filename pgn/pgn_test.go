package pgn

import (
	"strings"
	"testing"

	"github.com/mesb/mchess/address"
	"github.com/mesb/mchess/board"
	"github.com/mesb/mchess/socrates"
)

func TestExportAndImportRoundTrip(t *testing.T) {
	b := board.InitStandard()
	engine := socrates.New(b)

	play := []string{"e2e4", "e7e5", "g1f3"}
	for _, mv := range play {
		from, to := parseCoords(mv)
		if !engine.MakeMove(from, to, 0) {
			t.Fatalf("move %s failed", mv)
		}
	}

	pgnData := Export(engine.Log)
	if !strings.Contains(pgnData, "1. e2e4 e7e5 2. g1f3") {
		t.Fatalf("unexpected PGN: %s", pgnData)
	}

	other := socrates.New(board.InitStandard())
	if err := Import(other, pgnData); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	fenA := engine.Board.ToFEN(engine.State)
	fenB := other.Board.ToFEN(other.State)
	if fenA != fenB {
		t.Fatalf("FEN mismatch after import:\nA: %s\nB: %s", fenA, fenB)
	}
}

func parseCoords(m string) (address.Addr, address.Addr) {
	from := address.MakeAddr(address.Rank(m[1]-'1'), address.File(m[0]-'a'))
	to := address.MakeAddr(address.Rank(m[3]-'1'), address.File(m[2]-'a'))
	return from, to
}
