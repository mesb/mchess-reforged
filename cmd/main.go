// --- cmd/mchess/main.go ---

package main

import (
	"github.com/mesb/mchess/shell"
)

// main boots the mchess terminal engine using the GameSession + TerminalRenderer.
func main() {
	session := shell.NewSession(shell.TerminalRenderer{})
	shell.RunInteractive(session)
}
