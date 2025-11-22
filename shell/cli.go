// --- shell/cli.go ---

package shell

import (
	"bufio"
	"fmt"
	"os"
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

func handleInput(input string, session *GameSession) bool {
	// ... (Previous commands like 'b', 'q', 'h' remain)

	if strings.HasPrefix(input, "m ") {
		err := socrates.Dialog(input, session.Engine)
		if err != nil {
			session.Renderer.Message(err.Error())
			return false
		}

		showBoard(session)

		// Improved Endgame Handling
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
// func normalizeInput(input string) string {
// 	coordRe := regexp.MustCompile(`^[a-h][1-8][a-h][1-8]$`)
// 	if coordRe.MatchString(input) {
// 		return "m " + input
// 	}
// 	return input
// }

// colorName returns a human-readable color string.
func colorName(c int) string {
	if c == pieces.WHITE {
		return "White"
	}
	return "Black"
}
