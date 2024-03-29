package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"squava/src/abbook"
	"squava/src/alphabeta"
	"squava/src/mcts"
	"squava/src/negascout"
)

const (
	MAXIMIZER = 1
	MINIMIZER = -1
)

type Player interface {
	Name() string
	MakeMove(int, int, int) // x,y coords, type of player (MINIMIZER, MAXIMIZER)
	SetDepth(int)
	ChooseMove() (int, int, int, int) // x,y coords of move, value, leaf node count
	PrintBoard()
	SetScores(bool)
	FindWinner() int
}

func main() {

	maxDepthPtr := flag.Int("d", 10, "maximum lookahead depth (alpha/beta)")
	deterministic := flag.Bool("D", false, "Play deterministically")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	firstType := flag.String("1", "A", "first player type, A: alphabeta, N: negascout, B: A/B+book opening, G: A/B+avoid bad positions, M: MCTS")
	secondType := flag.String("2", "M", "second player type, A: alphabeta, N: negascout, B: A/B+book opening, G: A/B+avoid bad positions, M: MCTS")
	nonInteractive := flag.Int("n", 1, "play <number> games non-interactively")
	u1 := flag.Float64("u1", 0.50, "UCTK coefficient, player 1 (MCTS)")
	u2 := flag.Float64("u2", 0.50, "UCTK coefficient, player 2 (MCTS)")
	i1 := flag.Int("i1", 500000, "MCTS iterations, player 1")
	i2 := flag.Int("i2", 500000, "MCTS iterations, player 2")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	if *nonInteractive > 1 {
		nonInteractiveGames(*nonInteractive, *firstType, *secondType, *randomizeScores, *maxDepthPtr)
		return
	}

	var winner int

	moveCounter := 0

	first, second := createPlayers(*firstType,
		*secondType, *maxDepthPtr, *deterministic)

	if *firstType == "M" {
		first.(*mcts.MCTS).SetUCTK(*u1)
		first.(*mcts.MCTS).SetIterations(*i1)
	}

	if *secondType == "M" {
		second.(*mcts.MCTS).SetUCTK(*u2)
		second.(*mcts.MCTS).SetIterations(*i2)
	}

	first.SetScores(*randomizeScores)
	second.SetScores(*randomizeScores)

	gameStart := time.Now()
	for moveCounter < 25 {

		first.SetDepth(moveCounter)

		before := time.Now()
		i, j, value, leafCount := first.ChooseMove()
		et := time.Since(before)
		second.MakeMove(i, j, MINIMIZER)

		moveCounter++
		fmt.Printf("X (%s) <%d,%d> (%d) [%d] %v\n", first.Name(), i, j, value, leafCount, et)

		winner = first.FindWinner() // main() thinks first is maximizer
		if winner != 0 || moveCounter >= 25 {
			break
		}

		second.SetDepth(moveCounter)

		before = time.Now()
		i, j, value, leafCount = second.ChooseMove()
		et = time.Since(before)
		first.MakeMove(i, j, MINIMIZER)

		moveCounter++
		fmt.Printf("O (%s) <%d,%d> (%d) [%d] %v\n", second.Name(), i, j, value, leafCount, et)

		first.PrintBoard()

		winner1 := first.FindWinner()
		winner2 := -second.FindWinner() // main thinks second is minimizer
		if winner1 != winner2 {
			fmt.Printf("Winner disagreement. First %d, second %d\n", winner1, winner2)
		}
		if winner2 != 0 {
			winner = winner2
			break
		}

	}
	gameET := time.Since(gameStart)

	switch winner {
	case 1:
		fmt.Printf("player 1 X (%s) wins, %v\n", first.Name(), gameET)
	case -1:
		fmt.Printf("player 2 O (%s) wins, %v\n", second.Name(), gameET)
	default:
		fmt.Printf("Cat wins\n")
	}

	first.PrintBoard()

}

func nonInteractiveGames(gameCount int, firstType, secondType string, randomize bool, maxDepth int) {

	for i := 0; i < gameCount; i++ {
		moveCounter := 0

		first, second := createPlayers(firstType, secondType, maxDepth, randomize)

		fmt.Printf("%d %s %s %d %v ", i, first.Name(), second.Name(), maxDepth, randomize)

		var moves [25][2]int
		var values [25][2]int
		var winner int

		for moveCounter < 25 {

			first.SetDepth(moveCounter)
			i, j, value, _ := first.ChooseMove()
			moves[moveCounter][0], moves[moveCounter][1] = i, j
			values[moveCounter][0] = value
			second.MakeMove(i, j, MINIMIZER)
			moveCounter++
			winner = first.FindWinner()
			if winner != 0 || moveCounter >= 25 {
				break
			}

			second.SetDepth(moveCounter)
			i, j, value, _ = second.ChooseMove()
			moves[moveCounter][0], moves[moveCounter][1] = i, j
			values[moveCounter][1] = value
			first.MakeMove(i, j, MINIMIZER)
			moveCounter++
			winner = -second.FindWinner() // main thinks second is minimizer
			if winner != 0 {
				break
			}
		}

		fmt.Printf("%d %d", moveCounter, winner)

		for i := 0; i < moveCounter; i++ {
			marker := [2]string{"", ""}
			for j := 0; j < 2; j++ {
				if values[i][j] > 9000 {
					marker[j] = "+"
				}
				if values[i][j] < -9000 {
					marker[j] = "-"
				}
			}
			fmt.Printf(" %d%s,%d%s", moves[i][0], marker[0], moves[i][1], marker[1])
		}

		fmt.Printf("\n")
	}
}

func createPlayers(firstType, secondType string, maxDepth int, deterministic bool) (Player, Player) {

	firstType = strings.ToUpper(firstType)
	secondType = strings.ToUpper(secondType)

	var first, second Player

	switch firstType {
	case "A":
		first = alphabeta.New(deterministic, maxDepth)
	case "N":
		first = negascout.New(deterministic, maxDepth)
	case "B":
		first = abbook.New(deterministic, maxDepth)
	case "G":
		first = alphabeta.New(deterministic, maxDepth)
		first.(*alphabeta.AlphaBeta).SetAvoid()
	case "M":
		first = mcts.New(deterministic, maxDepth)
	}

	switch secondType {
	case "A":
		second = alphabeta.New(deterministic, maxDepth)
	case "N":
		second = negascout.New(deterministic, maxDepth)
	case "B":
		second = abbook.New(deterministic, maxDepth)
	case "G":
		second = alphabeta.New(deterministic, maxDepth)
		second.(*alphabeta.AlphaBeta).SetAvoid()
	case "M":
		second = mcts.New(deterministic, maxDepth)
	}

	return first, second
}
