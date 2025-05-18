package address

import (
	"testing"
)

func TestShift(t *testing.T) {
	start := Addr{Rank(1), File(4)} // e2
	target, ok := start.Shift(1, 0) // e3
	if !ok || !target.Equals(Addr{Rank(2), File(4)}) {
		t.Errorf("Shift failed: got %v", target)
	}
}

func TestTranslateIndex(t *testing.T) {
	index := 28 // e4
	addr := TranslateIndex(index)
	if addr.String() != "e4::28" {
		t.Errorf("Expected e4::28, got %v", addr.String())
	}
}

func TestDelta(t *testing.T) {
	from := Addr{Rank(6), File(4)} // e7
	to := Addr{Rank(4), File(4)}   // e5
	dr, df := Delta(from, to)
	if dr != -2 || df != 0 {
		t.Errorf("Expected delta (-2, 0), got (%v, %v)", dr, df)
	}
}
