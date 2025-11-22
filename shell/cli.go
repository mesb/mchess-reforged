// --- shell/cli.go ---

package shell

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mesb/mchess/pieces"
	"github.com/mesb/mchess/socrates"
)

// RunInteractive launches the main game loop.
// OPTIMIZATION: Uses a standard loop instead of recursion to prevent stack overflow.
func RunInteractive(session *GameSession) {
	showWelcome()
	showBoard(session)

	reader := bufio.NewReader(os.Stdin)

	// Main Event Loop
	for {
		// Prompt
		session.Renderer.Prompt(session.Engine.Turn)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		shouldQuit := handleInput(input, session)
		if shouldQuit {
			break
		}
	}
}

// showWelcome prints the initial CLI banner and instructions.
func showWelcome() {
	fmt.Println("       Welcome to MCHESS CLI - Dragon Edition")
	fmt.Println("----------------------------------------------")
	fmt.Println("Enter 'b' to see board")
	fmt.Println("Enter 'q' to quit")
	fmt.Println("Enter 'u' to undo last move")
	fmt.Println("Enter 'h' to view move history")
	fmt.Println("Enter moves like: m e2e4 or simply e2e4")
	fmt.Println()
}

// showBoard prints the current board and captured pieces.
func showBoard(session *GameSession) {
	session.Renderer.Render(session.Engine.Board)
	session.Renderer.ShowCaptured(session.Captured)
}

// handleInput interprets the input and returns true if user wants to quit.
func handleInput(input string, session *GameSession) bool {
	if input == "" {
		return false
	}

	// Restore shorthand support (e.g., allow "e2e4" instead of "m e2e4")
	input = normalizeInput(input)

	if input == "b" {
		showBoard(session)
		return false
	}

	if input == "q" {
		session.Renderer.Message("👋 Goodbye!")
		return true
	}

	if input == "h" {
		session.Log.PrintLog()
		return false
	}

	if input == "analyze" {
		session.Renderer.Message("Thinking...")
		result := session.Engine.Search(4) // depth 4 for quick response
		msg := fmt.Sprintf("Best Move: %s -> %s (Score: %d, Nodes: %d)", result.From, result.To, result.Score, result.Nodes)
		session.Renderer.Message(msg)
		return false
	}

	if input == "u" {
		if !session.Engine.UndoMove() {
			session.Renderer.Message("Nothing to undo.")
		} else {
			session.UpdateCaptured()
			showBoard(session)
		}
		return false
	}

	if strings.HasPrefix(input, "m ") {
		err := socrates.Dialog(input, session.Engine)
		if err != nil {
			session.Renderer.Message(err.Error())
			return false
		}

		session.UpdateCaptured()
		showBoard(session)

		// Check Game End States
		if session.Engine.IsCheckmate() {
			session.Renderer.Message("🏁 CHECKMATE! " + colorName(session.Engine.GetTurn()) + " loses.")
			return true
		}
		if session.Engine.IsStalemate() {
			session.Renderer.Message("⛔ STALEMATE. Draw.")
			return true
		}
		if session.Engine.IsFiftyMoveRule() {
			session.Renderer.Message("⏳ DRAW by 50-move rule.")
			return true
		}
		if session.Engine.IsInsufficientMaterial() {
			session.Renderer.Message("⚖️ DRAW by insufficient material.")
			return true
		}

		if session.Engine.IsInCheck(session.Engine.Turn) {
			session.Renderer.Message("⚠️  CHECK!")
		}

		return false
	}

	session.Renderer.Message("Unknown command. Try 'm e2e4', 'u', 'h', or 'q'")
	return false
}

// normalizeInput auto-corrects inputs like 'e2e4' to 'm e2e4'
// Allows 4 char (e2e4) and 5 char (a7a8q) inputs.
func normalizeInput(input string) string {
	// Regex for standard move or promotion move
	coordRe := regexp.MustCompile(`^[a-h][1-8][a-h][1-8][qrbn]?$`)
	if coordRe.MatchString(input) {
		return "m " + input
	}
	return input
}

// colorName returns a human-readable color string.
func colorName(c int) string {
	if c == pieces.WHITE {
		return "White"
	}
	return "Black"
}
