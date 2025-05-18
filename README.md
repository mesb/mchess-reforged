# ♟ MCHESS: Dragon Edition

> *“The board is a scroll, the pieces are glyphs, the game is dialectic.” — Hexa*

Welcome to **MCHESS**, a command-line chess engine forged in Go, designed not only for play, but for understanding.

This is not just a chess engine.  
It is a Platonic system of **move logic**, **rendering**, and **undoable histories**, built for humans and dragons alike.

---

## 🌱 Features

- 📦 Fully playable CLI chess game with legal move validation
- ♻️ Undo move functionality
- 📜 Move history log
- 🎭 Captured piece tracking
- 🧠 Structured engine built around modular principles: `engine`, `renderer`, `shell`, `dialogue`, `pieces`
- 🔁 Interactive coroutine-like CLI loop with recursion via `go listen(...)`
- 🧩 Easy to extend for multiplayer, AI, or GUI

---

## 🚀 Getting Started

### 🛠 Requirements

- Go ≥ 1.20

### 🧬 Build and Run

```bash
git clone https://github.com/mesb/mchess-reforged.git
cd mchess-reforged
go run ./cmd/main.go