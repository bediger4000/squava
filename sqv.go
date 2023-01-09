package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"squava/src/alphabeta"
	"squava/src/mcts"
	"squava/src/mcts2"
	"squava/src/mcts3"
)

const (
	HUMAN    = -1
	COMPUTER = 1
)

type Player interface {
	Name() string
	MakeMove(int, int, int) // x,y coords, type of player (HUMAN, COMPUTER)
	SetDepth(int)
	ChooseMove() (int, int, int, int) // x,y coords of move, value, leaf node count
	PrintBoard()
	SetScores(bool)
	FindWinner() int
}

func main() {

	computerFirstPtr := flag.Bool("C", false, "Computer takes first move (default false)")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	maxDepthPtr := flag.Int("d", 10, "maximum lookahead depth (alpha/beta)")
	typ := flag.String("t", "A", "first player type, A: alphabeta, G: A/B+avoid bad positions, M: MCTS")
	u := flag.Float64("u", 0.50, "UCTK coefficient, player 1 (MCTS)")
	i := flag.Int("i", 500000, "MCTS iterations, player 1")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	var winner int

	moveCounter := 0

	computerPlayer := createPlayer(*typ, *maxDepthPtr, *u, *i)

	computerPlayer.SetScores(*randomizeScores)

	next := HUMAN
	if *computerFirstPtr {
		next = COMPUTER
	}

	// computerPlayer keeps track of the board internally,
	// but we'll keep track too, so the human can be informed
	// that an input move has already been taken.
	bd := new(Board)

	for moveCounter < 25 {

		switch next {

		case HUMAN:
			l, m := bd.readMove()
			computerPlayer.MakeMove(l, m, HUMAN)
			next = COMPUTER

		case COMPUTER:
			computerPlayer.SetDepth(moveCounter)

			before := time.Now()
			i, j, value, leafCount := computerPlayer.ChooseMove()
			et := time.Since(before)

			fmt.Printf("X (%s) <%d,%d> (%d) [%d] %v\n", computerPlayer.Name(), i, j, value, leafCount, et)

			bd.makeMove(i, j, COMPUTER)
			next = HUMAN
		}

		moveCounter++
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

func createPlayer(typ string, maxDepth int, factor float64, iterations int) Player {

	typ = strings.ToUpper(typ)

	var computerPlayer Player

	switch typ {
	case "A":
		computerPlayer = alphabeta.New(false, maxDepth)
	case "G":
		computerPlayer = alphabeta.New(false, maxDepth)
		computerPlayer.(*alphabeta.AlphaBeta).SetAvoid()
	case "M":
		computerPlayer = mcts.New(false, maxDepth)
		computerPlayer.(*mcts.MCTS).SetUCTK(factor)
		computerPlayer.(*mcts.MCTS).SetIterations(iterations)
	case "L":
		computerPlayer = mcts2.New(false, maxDepth)
		computerPlayer.(*mcts2.MCTS2).SetUCTK(factor)
		computerPlayer.(*mcts2.MCTS2).SetIterations(iterations)
	case "P":
		computerPlayer = mcts3.New(false, maxDepth)
		computerPlayer.(*mcts3.MCTS3).SetIterations(iterations)
	}

	return computerPlayer
}

type Board [5][5]int

func (bd *Board) makeMove(x, y, player int) {
	bd[x][y] = player
}

func (bd *Board) readMove() (x, y int) {
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
	bd.makeMove(x, y, HUMAN)
	return x, y
}
