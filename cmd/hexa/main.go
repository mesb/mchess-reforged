package main

import (
	"fmt"
	"time"
)

// A Frame is a snapshot of the window's 1D memory buffer (e.g., terminal)
type Frame []string

// A Window is a recursive structure with a display channel
// It holds a pixel array (1D), and updates via frames
// Inspired by OCaml's functional purity and Emacs' buffers

const (
	width  = 8
	height = 8
	total  = width * height
	fps    = 60
)

// Initializes a board as 1D array of glyphs
func initBoard() []string {
	return []string{
		"\u265C", "\u265E", "\u265D", "\u265B", "\u265A", "\u265D", "\u265E", "\u265C",
		"\u265F", "\u265F", "\u265F", "\u265F", "\u265F", "\u265F", "\u265F", "\u265F",
		"--", "--", "--", "--", "--", "--", "--", "--",
		"--", "--", "--", "--", "--", "--", "--", "--",
		"--", "--", "--", "--", "--", "--", "--", "--",
		"--", "--", "--", "--", "--", "--", "--", "--",
		"\u2659", "\u2659", "\u2659", "\u2659", "\u2659", "\u2659", "\u2659", "\u2659",
		"\u2656", "\u2658", "\u2657", "\u2655", "\u2654", "\u2657", "\u2658", "\u2656",
	}
}

// renderFrame draws the current buffer as a board
func renderFrame(frame Frame) {
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			fmt.Printf("%s ", frame[i*width+j])
		}
		fmt.Println()
	}
	fmt.Println()
}

func main() {
	buffer := initBoard()
	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	frame := make(Frame, total)
	copy(frame, buffer)

	for frameCount := 0; ; frameCount++ {
		select {
		case <-ticker.C:
			fmt.Printf("Frame: %d\n", frameCount)
			renderFrame(frame)
		}
	}
}
