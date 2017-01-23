package main

import (
	"alphabeta"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	MAXIMIZER = 1
	MINIMIZER = -1
)

func main() {

	maxDepthPtr := flag.Int("d", 10, "maximum lookahead depth")
	deterministic := flag.Bool("D", false, "Play deterministically")
	printBoardPtr := flag.Bool("n", false, "Don't print board, just emit moves")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	flag.Parse()

	*printBoardPtr = !*printBoardPtr

	rand.Seed(time.Now().UTC().UnixNano())

	var winner int

	moveCounter := 0

	first := alphabeta.New(*deterministic, *maxDepthPtr)
	second := alphabeta.New(*deterministic, *maxDepthPtr)
	first.SetScores(*randomizeScores)
	second.SetScores(*randomizeScores)

	for moveCounter < 25 {

		first.SetDepth(moveCounter)
		i, j, value, leafCount := first.ChooseMove()
		moveCounter++
		fmt.Printf("First <%d,%d> (%d) [%d]\n", i, j, value, leafCount)
		//first.PrintBoard()

		winner = first.FindWinner()

		if winner != 0 || moveCounter >= 25 {
			break
		}

		second.MakeMove(i, j, MINIMIZER)

		second.SetDepth(moveCounter)
		i, j, value, leafCount = second.ChooseMove()
		moveCounter++
		fmt.Printf("Secnd <%d,%d> (%d) [%d]\n", i, j, value, leafCount)
		second.PrintBoard()

		winner = -second.FindWinner()

		if winner != 0 {
			break
		}

		first.MakeMove(i, j, MINIMIZER)
	}

	os.Exit(0)
}
