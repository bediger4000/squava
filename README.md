# The Game of Squava

## Rules

Like tic-tac-toe on a 5x5 grid of cells. Players alternate marking cells.
Four cells of the same mark in a row (verical, horizontal or diagonal) wins
for the player with that mark. Three cells in a row loses.

There's a little ambiguity in that it isn't clear what to do if a single marker
fills in a row of 3, and a row of 4. Does that player win or lose?

## References

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
