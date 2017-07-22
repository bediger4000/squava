package main

import (
	"alphabeta"
	"abbook"
	"flag"
	"fmt"
	"math/rand"
	"negascout"
	"os"
	"time"
)

const (
	MAXIMIZER = 1
	MINIMIZER = -1
)

type Player interface {
	MakeMove(int, int, int) // x,y corrds, type of player (MINIMIZER, MAXIMIZER)
	SetDepth(int)
	ChooseMove() (int, int, int, int) // x,y coords of move, value, leaf node count
	PrintBoard()
	SetScores(bool)
	FindWinner() int
}

func main() {

	maxDepthPtr := flag.Int("d", 10, "maximum lookahead depth")
	deterministic := flag.Bool("D", false, "Play deterministically")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	firstType := flag.String("1", "A", "first player type, A: alphabeta, N: negascout")
	secondType := flag.String("2", "N", "first player type, A: alphabeta, N: negascout")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	var winner int

	moveCounter := 0

	var first, second Player

	var firstName string
	var secondName string
	switch *firstType {
	case "A":
		first = alphabeta.New(*deterministic, *maxDepthPtr)
		firstName = "AlphaBeta"
	case "N":
		first = negascout.New(*deterministic, *maxDepthPtr)
		firstName = "NegaScout"
	case "B":
		first = abbook.New(*deterministic, *maxDepthPtr)
		firstName = "A/B+Book"
	}

	switch *secondType {
	case "A":
		second = alphabeta.New(*deterministic, *maxDepthPtr)
		secondName = "AlphaBeta"
	case "N":
		second = negascout.New(*deterministic, *maxDepthPtr)
		secondName = "NegaScout"
	case "B":
		second = abbook.New(*deterministic, *maxDepthPtr)
		secondName = "A/B+Book"
	}

	first.SetScores(*randomizeScores)
	second.SetScores(*randomizeScores)

	for moveCounter < 25 {

		first.SetDepth(moveCounter)

		i, j, value, leafCount := first.ChooseMove()
		second.MakeMove(i, j, MINIMIZER)

		moveCounter++
		fmt.Printf("X (%s) <%d,%d> (%d) [%d]\n", firstName, i, j, value, leafCount)

		winner = first.FindWinner() // main() thinks first is maximizer
		if winner != 0 || moveCounter >= 25 {
			break
		}

		second.SetDepth(moveCounter)

		i, j, value, leafCount = second.ChooseMove()
		first.MakeMove(i, j, MINIMIZER)

		moveCounter++
		fmt.Printf("O (%s) <%d,%d> (%d) [%d]\n", secondName, i, j, value, leafCount)

		first.PrintBoard()

		winner = -second.FindWinner() // main thinks second is minimizer
		if winner != 0 {
			break
		}

	}

	switch winner {
	case 1:
		fmt.Printf("X (%s) wins\n", firstName)
	case -1:
		fmt.Printf("0 (%s) wins\n", secondName)
	default:
		fmt.Printf("Cat wins\n")
	}

	first.PrintBoard()

	os.Exit(0)
}
