import { Injectable, Signal, signal } from '@angular/core';
import { ChessKernel, ConnectionStatus } from '../../core/chess-kernel';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class GameService {
  private kernel = new ChessKernel();
  private apiBase = environment.apiBase;
  private wsBase = environment.wsBase;

  private gameId = signal<string | null>(null);
  private boardSignal = signal<string[]>([...this.kernel.board]);
  private statusSignal = signal<ConnectionStatus>(this.kernel.connectionStatus);
  private turnSignal = signal<'w' | 'b'>('w');
  private stateText = signal<string>('Idle');

  constructor() {
    this.kernel.onBoard((b) => this.boardSignal.set(b));
    this.kernel.onStatus((s) => this.statusSignal.set(s));
    this.kernel.onState((st) => {
      this.turnSignal.set(this.kernel.turn);
      if (st.game_id) this.gameId.set(st.game_id);
      if (st.status) this.stateText.set(st.status);
    });
  }

  board(): Signal<string[]> {
    return this.boardSignal;
  }

  status(): Signal<ConnectionStatus> {
    return this.statusSignal;
  }

  turn(): Signal<'w' | 'b'> {
    return this.turnSignal;
  }

  state(): Signal<string> {
    return this.stateText;
  }

  game(): Signal<string | null> {
    return this.gameId;
  }

  async startNewGame(): Promise<void> {
    try {
      const res = await fetch(`${this.apiBase}/games`, { method: 'POST' });
      const json = await res.json();
      const id = json?.game_id;
      if (!id) throw new Error('No game_id in response');
      this.gameId.set(id);
      this.kernel.connect(id, this.wsBase);
      this.stateText.set('Active');
      await this.fetchState(id);
    } catch (err) {
      console.error('Failed to create/connect game', err);
      this.stateText.set('Connection failed');
    }
  }

  connect(gameId: string): void {
    this.gameId.set(gameId);
    this.kernel.connect(gameId, this.wsBase);
    this.fetchState(gameId);
  }

  async move(from: number, to: number): Promise<void> {
    const gid = this.gameId();
    if (!gid) return;
    const uci = this.indexToCoord(from) + this.indexToCoord(to);
    try {
      const res = await fetch(`${this.apiBase}/games/${gid}/move`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ move: uci }),
      });
      if (!res.ok) {
        console.error('Move rejected', await res.text());
        return;
      }
      const snapshot = await res.json();
      this.kernel.applySnapshot(snapshot);
    } catch (e) {
      console.error('Failed to send move', e);
    }
  }

  private async fetchState(id: string): Promise<void> {
    try {
      const res = await fetch(`${this.apiBase}/games/${id}`);
      if (!res.ok) return;
      const snapshot = await res.json();
      this.kernel.applySnapshot(snapshot);
    } catch (e) {
      console.error('Failed to fetch state', e);
    }
  }

  private indexToCoord(idx: number): string {
    const file = idx % 8;
    const rank = Math.floor(idx / 8);
    return String.fromCharCode('a'.charCodeAt(0) + file) + (rank + 1).toString();
  }
}
