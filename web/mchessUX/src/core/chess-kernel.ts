export type Turn = 'w' | 'b';
export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected';

type Listener<T> = (value: T) => void;
interface GameStatePayload {
  board_fen?: string;
  turn?: string;
  status?: string;
  game_id?: string;
}

/**
 * ChessKernel is a lightweight, framework-agnostic core that tracks a flat 64-slot board
 * and speaks a minimal move protocol over WebSockets.
 */
export class ChessKernel {
  readonly board: string[] = new Array(64).fill('');
  fen = '';
  turn: Turn = 'w';
  connectionStatus: ConnectionStatus = 'disconnected';
  statusText = 'Idle';
  gameId: string | null = null;

  private socket?: WebSocket;
  private boardListeners: Listener<string[]>[] = [];
  private statusListeners: Listener<ConnectionStatus>[] = [];
  private stateListeners: Listener<GameStatePayload>[] = [];

  connect(gameId: string, endpoint = 'ws://localhost:8080/ws'): void {
    this.disconnect();
    this.setStatus('connecting');
    this.gameId = gameId;
    const url = `${endpoint}?game_id=${encodeURIComponent(gameId)}`;
    this.socket = new WebSocket(url);

    this.socket.onopen = () => this.setStatus('connected');
    this.socket.onclose = () => this.setStatus('disconnected');
    this.socket.onerror = () => this.setStatus('disconnected');
    this.socket.onmessage = (ev) => {
      try {
        const payload = JSON.parse(ev.data);
        if (payload.board_fen) {
          this.applyFen(payload.board_fen);
        }
        if (payload.turn) {
          this.turn = payload.turn.toLowerCase().startsWith('b') ? 'b' : 'w';
        }
        if (payload.status) {
          this.statusText = payload.status;
        }
        this.emitState(payload);
      } catch {
        /* ignore malformed */
      }
    };
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = undefined;
    }
    this.setStatus('disconnected');
  }

  onBoard(listener: Listener<string[]>): () => void {
    this.boardListeners.push(listener);
    return () => {
      this.boardListeners = this.boardListeners.filter((l) => l !== listener);
    };
  }

  onStatus(listener: Listener<ConnectionStatus>): () => void {
    this.statusListeners.push(listener);
    return () => {
      this.statusListeners = this.statusListeners.filter((l) => l !== listener);
    };
  }

  onState(listener: Listener<GameStatePayload>): () => void {
    this.stateListeners.push(listener);
    return () => {
      this.stateListeners = this.stateListeners.filter((l) => l !== listener);
    };
  }

  move(from: number, to: number): void {
    if (!this.socket || this.connectionStatus !== 'connected') return;
    const uci = this.indexToCoord(from) + this.indexToCoord(to);
    this.socket.send(JSON.stringify({ move: uci }));
  }

  applySnapshot(state: GameStatePayload): void {
    if (state.board_fen) {
      this.applyFen(state.board_fen);
    }
    if (state.turn) {
      this.turn = state.turn.toLowerCase().startsWith('b') ? 'b' : 'w';
    }
    if (state.status) {
      this.statusText = state.status;
    }
    if (state.game_id) {
      this.gameId = state.game_id;
    }
    this.emitState(state);
  }

  private applyFen(fen: string): void {
    this.fen = fen;
    const parts = fen.split(' ');
    const placement = parts[0] ?? '';
    let idx = 0;
    for (let i = 0; i < placement.length && idx < 64; i++) {
      const c = placement[i];
      if (c === '/') continue;
      if (c >= '1' && c <= '8') {
        idx += parseInt(c, 10);
      } else {
        this.board[idx++] = c;
      }
    }
    this.turn = (parts[1] ?? 'w') === 'b' ? 'b' : 'w';
    this.emitBoard();
  }

  private emitBoard(): void {
    const snapshot = [...this.board];
    this.boardListeners.forEach((l) => l(snapshot));
  }

  private setStatus(status: ConnectionStatus): void {
    this.connectionStatus = status;
    this.statusListeners.forEach((l) => l(status));
  }

  private emitState(state: GameStatePayload): void {
    this.stateListeners.forEach((l) => l(state));
  }

  private indexToCoord(idx: number): string {
    const file = idx % 8;
    const rank = Math.floor(idx / 8);
    return String.fromCharCode('a'.charCodeAt(0) + file) + (rank + 1).toString();
  }
}
