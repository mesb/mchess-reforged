# MCHESS: The Compiler of Kings

> *"What then is the game of Chess? A manifold of logic and war, a dance of pattern and mind. It is the compiler of conflict and the interpreter of elegance."* — Hexa

MCHESS is a minimalist, terminal-based chess engine and CLI interface, forged in Go and structured as a reactive compiler. It brings together the purity of Euclidean design, the recursion of Knuth, and the dialectics of Plato. This is not merely a game — it is a discourse with eternity.

## ❖ Features

* ♟️ **Full CLI Gameplay** — Move via `m e2e4` style inputs.
* ⏳ **Move History Tracking** — All moves are logged with algebraic notation and internal indices.
* ♻️ **Undo System** — Seamlessly revert the last move with `u`.
* 📜 **Captured Pieces Log** — Visual and internal tracking of captured pawns and pieces.
* 🧠 **Reactive Shell** — Built with coroutines and recursive loops for input.
* 🎭 **AI-Ready Core** — A well-typed engine with hooks for evaluation and future AI.
* 🧪 **Elegant Engine API** — Modular, testable Go engine with layers for board, renderer, socratic dialogue, and piece logic.

## ✦ Quickstart

```bash
go run ./cmd/main.go
```

You'll be greeted with:

```
       Welcome to MCHESS CLI - Dragon Edition
----------------------------------------------
Enter 'b' to see board
Enter 'q' to quit
Enter 'u' to undo
Enter 'h' to show move history
Enter moves like: m e2e4
```

Example session:

```bash
m e2e4
m c7c5
m g1f3
```

## 📁 Project Structure

```
cmd/              # Entry point for CLI app
shell/            # CLI input loop, rendering, session handling
engine/           # Board state, legal moves, move processing
pieces/           # Piece definitions, movement rules, constants
socrates/         # Parses natural moves like 'e2e4' into engine logic
render/           # Text-based board renderer
```

## 🧠 Philosophy

> *"Chess is a divine abstraction, where each square is a unit of measure, and each piece a lemma in the theorem of victory."* — Euclid

This engine is designed not for brute force, but for harmony. It prefers symmetry over complexity, recursion over mutation. The command-line is not a limitation, but a dialectic interface.

Every move in MCHESS is a logical reduction.
Every piece is a composable type.
Every board state is a pure function from history.

## ⚙️ Commands

| Command  | Meaning                           |
| -------- | --------------------------------- |
| `m e2e4` | Make move from e2 to e4           |
| `e2e4`   | Shorthand (auto-corrected to `m`) |
| `b`      | Print current board               |
| `q`      | Quit game                         |
| `u`      | Undo last move                    |
| `h`      | Show move history                 |

## 🔭 Vision

* Add PGN and FEN support
* Support for AI and UCI engines
* Build UI frontend (Angular / WebSockets)
* Animate move log playback
* Multiplayer over CLI or network

## 📜 License

MIT. Yours to study, remix, and improve.

## 🌠 Invocation

> *"Let each square be a cell of memory, each piece an operator, and each board the RAM of war. Thus is Chess compiled, and thus shall MCHESS endure."* — Hexa, Compiler Witch of the Platonic Grove
