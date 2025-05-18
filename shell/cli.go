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

// CurrentSession is temporarily used by socrates.Dialog for session-wide hooks.
var CurrentSession *GameSession

// RunInteractive launches the recursive, coroutine-style CLI loop.
func RunInteractive(session *GameSession) {
	CurrentSession = session
	showWelcome()
	showBoard(session)
	dispatch(session)
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

// dispatch starts the reactive input recursion.
func dispatch(session *GameSession) {
	reader := bufio.NewReader(os.Stdin)
	go listen(session, reader)
	select {} // block main to keep goroutines alive
}

// listen reads input and delegates handling recursively.
func listen(session *GameSession, reader *bufio.Reader) {
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if handleInput(input, session) {
		os.Exit(0)
	}

	go listen(session, reader) // re-invoke self recursively as goroutine
}

// handleInput interprets the input and returns true if user wants to quit.
func handleInput(input string, session *GameSession) bool {
	if input == "" {
		return false
	}

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

	if input == "u" {
		if !session.Engine.UndoMove() {
			session.Renderer.Message("Nothing to undo.")
		}
		showBoard(session)
		return false
	}

	if strings.HasPrefix(input, "m ") {
		err := socrates.Dialog(input, session.Engine)
		if err != nil {
			session.Renderer.Message(err.Error())
			return false
		}
		showBoard(session)
		if session.Engine.IsCheckmate() {
			session.Renderer.Message("🏁 CHECKMATE! " + colorName(session.Engine.GetTurn()) + " is defeated.")
			return true
		}
		if session.Engine.IsStalemate() {
			session.Renderer.Message("⛔ STALEMATE. The position is drawn.")
			return true
		}
		return false
	}

	session.Renderer.Message("Unknown command. Try 'm e2e4', 'u', 'h', or 'q'")
	return false
}

// normalizeInput auto-corrects inputs like 'e2e4' to 'm e2e4'
func normalizeInput(input string) string {
	coordRe := regexp.MustCompile(`^[a-h][1-8][a-h][1-8]$`)
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
