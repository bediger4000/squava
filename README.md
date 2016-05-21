# The Game of Squava

## Rules

Like tic-tac-toe on a 5x5 grid of cells. Players alternate marking cells.
Four cells of the same mark in a row (verical, horizontal or diagonal) wins
for the player with that mark. Three cells in a row loses.

There's a little ambiguity in that it isn't clear what to do if a single marker
fills in a row of 3, and a row of 4. Does that player win or lose?

I chose "win", mainly because it's computationally easier to check for 4-in-a-row
as a win separately from 3-in-a-row as a loss. After all, every 4-in-a-row has
3-in-a-row inside it.

## References

Apparently Squava was invented by
[Néstor Romeral Andrés](https://bitcoinmagazine.com/articles/rise-of-the-machines-1383576469)

https://boardgamegeek.com/boardgame/112745/squava

https://en.aeriesguard.com/Squava-puzzles

## This Program

Command line, text interface.

Go programming language.

Alpha-Beta minimax, [algorithm](https://en.wikipedia.org/wiki/Alpha%E2%80%93beta_pruning)
from Wikipedia.

### Static Valuation Function

By default it looks 6 plies into the game tree. That's 3 moves each for 'X' and
'O'. The static value has a slight bias towards moves at the center of the
board, and a slight bias towards winning in as few moves as possible.

After the slight biases, it gives larger magnitude scores for having
a non-losing any 3 out of a winnning 4-in-a-row combination.

It gives a large negative score for the computer losing by 3-in-a-row, or by
human winnning.  It gives a large positive score for for computer winning, or
human losing by 3-in-a-row.

Other than tic-tac-toe, where it's feasible to check the entire game tree
on every move, this is the first static valuation function I've written
that actually produces a worthwhile opponent, and it's also quite simple.

## Building the program

    go build squava.go

## Running the program

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
