package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"squava/src/abgeo"
	"squava/src/alphabeta"
	"squava/src/mcts"
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

	humanFirstPtr := flag.Bool("H", true, "Human takes first move")
	computerFirstPtr := flag.Bool("C", false, "Computer takes first move")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	maxDepthPtr := flag.Int("d", 10, "maximum lookahead depth (alpha/beta)")
	typ := flag.String("t", "A", "first player type, A: alphabeta, G: A/B+avoid bad positions, M: MCTS")
	u := flag.Float64("u", 0.50, "UCTK coefficient, player 1 (MCTS)")
	i := flag.Int("i", 500000, "MCTS iterations, player 1")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	var winner int

	moveCounter := 0

	computerPlayer := createPlayer(*typ, *maxDepthPtr)

	if *typ == "M" {
		computerPlayer.(*mcts.MCTS).SetUCTK(*u)
		computerPlayer.(*mcts.MCTS).SetIterations(*i)
	}

	computerPlayer.SetScores(*randomizeScores)

	humanFirst := *humanFirstPtr
	if *computerFirstPtr {
		humanFirst = false
	}

	// computerPlayer keeps track of the board internally,
	// but we'll keep track too, so the human can be informed
	// that an input move has already been taken.
	var bd [5][5]int

	for moveCounter < 25 {

		if humanFirst {
			l, m := readMove(&bd)
			bd[l][m] = MINIMIZER
			computerPlayer.MakeMove(l, m, MINIMIZER)
			winner = computerPlayer.FindWinner()
			moveCounter++
			if winner != 0 || moveCounter >= 25 {
				break
			}
		}

		humanFirst = true

		computerPlayer.SetDepth(moveCounter)
		before := time.Now()
		i, j, value, leafCount := computerPlayer.ChooseMove()
		et := time.Since(before)

		bd[i][j] = MAXIMIZER
		moveCounter++
		fmt.Printf("X (%s) <%d,%d> (%d) [%d] %v\n", computerPlayer.Name(), i, j, value, leafCount, et)

		winner = computerPlayer.FindWinner()
		if winner != 0 || moveCounter >= 25 {
			break
		}

		computerPlayer.PrintBoard()
	}

	switch winner {
	case 1:
		fmt.Printf("player 1 X (%s) wins\n", computerPlayer.Name())
	case -1:
		fmt.Printf("player 2 O (human) wins\n")
	default:
		fmt.Printf("Cat wins\n")
	}

	computerPlayer.PrintBoard()
}

func createPlayer(typ string, maxDepth int) Player {

	typ = strings.ToUpper(typ)

	var computerPlayer Player

	switch typ {
	case "A":
		computerPlayer = alphabeta.New(false, maxDepth)
	case "G":
		computerPlayer = abgeo.New(false, maxDepth)
	case "M":
		computerPlayer = mcts.New(false, maxDepth)
	}

	return computerPlayer
}

func readMove(bd *[5][5]int) (x, y int) {
	readMove := false
	for !readMove {
		fmt.Printf("Your move: ")
		_, err := fmt.Scanf("%d %d\n", &x, &y)
		if err == io.EOF {
			os.Exit(0)
		}
		if err != nil {
			fmt.Printf("Failed to read: %v\n", err)
			os.Exit(1)
		}
		switch {
		case x < 0 || x > 4 || y < 0 || y > 4:
			fmt.Printf("Choose two numbers between 0 and 4, try again\n")
		case bd[x][y] == 0:
			readMove = true
		case bd[x][y] != 0:
			fmt.Printf("Cell (%d, %d) already occupied, try again\n", x, y)
		}
	}
	return x, y
}
