# The Game of Squava

## Rules

Like tic-tac-toe on a 5x5 grid of cells. Players alternate marking cells.
Four cells of the same mark in a row (verical, horizontal or diagonal) wins
for the player with that mark. Three cells in a row loses.

There's a little ambiguity in that it isn't clear what to do if a single marker
fills in a row of 3, and a row of 4.

## References

https://boardgamegeek.com/boardgame/112745/squava

https://en.aeriesguard.com/Squava-puzzles

## This Program

Command line, text interface.

Go programming language.

Alpha-Beta minimax, [algorithm](https://en.wikipedia.org/wiki/Alpha%E2%80%93beta_pruning)
from Wikipedia.

### Static Valuation Function

## Building the program

    go build squava.go

## Running the program

    ./squava
