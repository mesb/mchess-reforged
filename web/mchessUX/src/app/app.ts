import { Component } from '@angular/core';
import { BoardComponent } from './components/board/board.component';
import { GameService } from './services/game.service';
import { MaterialModule } from './material.module';

@Component({
  selector: 'app-root',
  imports: [BoardComponent, MaterialModule],
  templateUrl: './app.html',
})
export class App {
  connection;
  turn;
  state;
  gameId;

  constructor(private game: GameService) {
    this.connection = this.game.status();
    this.turn = this.game.turn();
    this.state = this.game.state();
    this.gameId = this.game.game();
  }

  turnLabel(): string {
    return this.turn() === 'w' ? 'White' : 'Black';
  }

  newGame(): void {
    this.game.startNewGame();
  }

  join(id: string): void {
    if (!id) return;
    this.game.connect(id);
  }
}
