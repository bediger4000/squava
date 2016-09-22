# The Game of Squava

## Rules

Squava is a tic-tac-toe variant. Moves are made like tic-tac-toe, except on a
5x5 grid of cells. Players alternate marking cells, conventionally with `X` or
`O`.  Four cells of the same mark in a row (verical, horizontal or diagonal)
wins for the player with that mark. Three cells in a row loses. That is, a
player can win outright, or lose.

You can [play the JavaScript version](https://rawgit.com/bediger4000/squava/master/squava.html) right now!

The rules have an ambiguity, in that it isn't clear what to do if a single marker
fills in a row of 3, say, and a diagonal of 4. Does that player win or lose?

I chose "win", mainly because it's computationally easier to check for 4-in-a-row
as a win separately from 3-in-a-row as a loss. After all, every 4-in-a-row has
3-in-a-row inside it.

## Golang Program

Command line, text interface.  `squava` (the program) command line options:

      -C    Computer takes first move
      -D    Play deterministically
      -H    Human takes first move (default true)
      -d int
            maximum lookahead depth (default 9)


Alpha-Beta minimax, [algorithm](https://en.wikipedia.org/wiki/Alpha%E2%80%93beta_pruning)
from Wikipedia. Since more than one move can result in the maximum numerical score,
`squava` keeps a list of all moves that have that maximum numerical score, and chooses
one at random to actually play. The `-D` flag will skip the choice, and always play
the same move in identical situations.

The lookahead (number of moves ahead the code considers) varies during the game.
If less than 8 marks appear on the board, it looks ahead 4 moves (2 moves each for both players).
If less than 13 marks appear on the board, it looks ahead 3 moves for each player.
If more than 12 marks appear, it looks ahead 4 moves for each player.
This is something I set by trial and error. If the human plays right in the first few moves
when the program isn't looking too far ahead, the human can win.

### Book

Every good alpha-beta minimax program has a "book" for the initial moves.
I've included some programs to help build a "book" for `squava`:

`opening2` can performa very deep valuation of the 6 cells that are unique first moves.
There's 25 empty cells at the start of a game. All but 6 first moves can be generated by rotations
or reflections of the board.

`opening3` can perform very deep evaluations of all 6 first moves, and all unique (under rotaion
and reflection) response moves. This lets me check which response is best for each opening move.

Both `opening2` and `opening3` have some ugly code changes relative to `squava` to speed things
up. First, they calculate a board's numerical score progressively. Each move made during alpha-beta
minimaxing has its individual contribution to the board's score added when the program considers it.
This avoids repeating for every leaf node of a game tree most of the calculations made by `squava` .
I folded the static valuation of the board into the function `alphaBeta()` to avoid any function
call overhead. Both changes sped up the programs.

## JavaScript Program

Point-n-click, runs in your browser. Single HTML file.

JavaScript transliteration of an earlier version of the Golang version.

### Static Valuation Function

After reaching its lookahead depth (which varies throughout the game) the
code does a static valuation of the board - it assigns a numerical value
to the layout of X's and O's.
The static value has a slight bias towards moves at the corners and edges of the
board, and a slight bias towards winning (or forcing a loss) in as few moves as possible.

After the slight biases, it gives larger magnitude scores for having
a non-losing any 3 out of a winnning 4-in-a-row combination.

It gives a large negative score for the computer losing by 3-in-a-row, or by
human winnning.  It gives a large positive score for for computer winning, or
human losing by 3-in-a-row.

Experimentally, considering 2-of-loosing-3 does not make a difference in
the program's play.  I think this is because every winning 4-in-a-row
stems from a losing 3-in-a-row created two moves before.

Other than tic-tac-toe, where it's feasible to check the entire game tree
on every move, this is the first static valuation function I've written
that actually produces a worthwhile opponent, and it's also quite simple.

## Building the Golang program

    go build squava.go

## Running the Golang program

    ./squava
    Your move:

You enter 2 digits in the range 0 to 4, with a space or spaces between them.
The computer ponders, announces its move, and displays the board. Human plays
'O', computer plays 'X'.

You can have the computer go first:

    ./squava -C
	My move: 2 2
       0 1 2 3 4
    0  _ _ _ _ _
    1  _ _ _ _ _
    2  _ _ X _ _
    3  _ _ _ _ _
    4  _ _ _ _ _
    Your move:

I find that a typical game has two phases: opening, where there's up to 5
pieces on the board.  A midgame, where you try to win by getting 4-in-a-row,
while keeping the computer from getting 4-in-a-row. Rarely, you can get to a
third phase, an end game, where no 4-in-a-row is possible, and the goal becomes
to avoid losing by being forced into 3-in-a-row.

You can see the end game taking place by running two instances of `squava` in
two `xterms`. Start one as `./squava -C`. It will chose a move first. Type
that move into the second instance, which expects the "human" to move first.

## References

Apparently Squava was invented by
[Néstor Romeral Andrés](https://bitcoinmagazine.com/articles/rise-of-the-machines-1383576469)

https://boardgamegeek.com/boardgame/112745/squava

https://en.aeriesguard.com/Squava-puzzles
