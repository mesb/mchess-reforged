import { Component, signal, Signal } from '@angular/core';
import { ConnectionStatus } from '../../../core/chess-kernel';
import { GameService } from '../../services/game.service';

type Square = { id: number; x: number; y: number; colorClass: string };

@Component({
  selector: 'app-board',
  standalone: true,
  templateUrl: './board.component.html',
  styleUrls: ['./board.component.scss'],
})
export class BoardComponent {
  board!: Signal<string[]>;
  turn!: Signal<'w' | 'b'>;
  state!: Signal<string>;
  connection!: Signal<ConnectionStatus>;
  gameId!: Signal<string | null>;
  squares: Square[] = [];

  dragging = signal(false);
  draggingPieceChar = signal('');
  dragTransform = signal('');

  selected = signal<number | null>(null);

  constructor(private game: GameService) {
    this.board = this.game.board();
    this.turn = this.game.turn();
    this.state = this.game.state();
    this.connection = this.game.status();
    this.gameId = this.game.game();
    this.squares = this.buildSquares();
    this.game.startNewGame();
  }

  tap(event: MouseEvent | TouchEvent): void {
    // Tap-to-move: select then move on second tap.
    const idx = this.pickIndex(event);
    if (idx == null) return;
    const current = this.selected();
    const b = this.board();
    if (current == null) {
      if (!b[idx]) return;
      this.selected.set(idx);
    } else {
      if (current !== idx) {
        this.game.move(current, idx);
      }
      this.selected.set(null);
    }
  }

  startDrag(event: MouseEvent | TouchEvent): void {
    const idx = this.pickIndex(event);
    if (idx == null) return;
    const b = this.board();
    if (!b[idx]) return;
    this.selected.set(idx);
    this.dragging.set(true);
    this.draggingPieceChar.set(this.getUnicode(b[idx]));
    this.updateDragTransform(event);
  }

  drag(event: MouseEvent | TouchEvent): void {
    if (!this.dragging()) return;
    this.updateDragTransform(event);
  }

  endDrag(event: MouseEvent | TouchEvent): void {
    const from = this.selected();
    if (from == null) {
      this.dragging.set(false);
      return;
    }
    const to = this.pickIndex(event);
    this.dragging.set(false);
    this.draggingPieceChar.set('');
    this.dragTransform.set('');
    this.selected.set(null);
    if (to != null && to !== from) {
      this.game.move(from, to);
    }
  }

  isWhite(char: string): boolean {
    return char === char.toUpperCase();
  }

  getUnicode(char: string): string {
    const map: Record<string, string> = {
      K: '♔',
      Q: '♕',
      R: '♖',
      B: '♗',
      N: '♘',
      P: '♙',
      k: '♚',
      q: '♛',
      r: '♜',
      b: '♝',
      n: '♞',
      p: '♟',
    };
    return map[char] ?? '';
  }

  rankFromIndex(idx: number): number {
    return 7 - Math.floor(idx / 8);
  }

  private buildSquares(): Square[] {
    const out: Square[] = [];
    for (let r = 0; r < 8; r++) {
      for (let f = 0; f < 8; f++) {
        const id = r * 8 + f;
        const light = (r + f) % 2 === 0;
        out.push({ id, x: f, y: 7 - r, colorClass: light ? 'light' : 'dark' });
      }
    }
    return out;
  }

  private pickIndex(ev: MouseEvent | TouchEvent): number | null {
    const svg = (ev.target as Element).closest('svg');
    if (!svg) return null;
    const rect = svg.getBoundingClientRect();
    const point = 'touches' in ev ? ev.touches[0] ?? ev.changedTouches[0] : ev;
    if (!point) return null;
    const xNorm = ((point.clientX - rect.left) / rect.width) * 8;
    const yNorm = ((point.clientY - rect.top) / rect.height) * 8;
    const file = Math.floor(xNorm);
    const rank = 7 - Math.floor(yNorm);
    if (file < 0 || file > 7 || rank < 0 || rank > 7) return null;
    return rank * 8 + file;
  }

  private updateDragTransform(ev: MouseEvent | TouchEvent): void {
    const svg = (ev.target as Element).closest('svg');
    if (!svg) return;
    const rect = svg.getBoundingClientRect();
    const point = 'touches' in ev ? ev.touches[0] ?? ev.changedTouches[0] : ev;
    if (!point) return;
    const xNorm = ((point.clientX - rect.left) / rect.width) * 8;
    const yNorm = ((point.clientY - rect.top) / rect.height) * 8;
    this.dragTransform.set(`translate(${xNorm} ${yNorm})`);
  }
}
