package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Board [5][5]int

const WIN = 10000
const LOSS = -10000
const MAXIMIZER = 1
const MINIMIZER = -1
const UNSET = 0

var maxDepth int = 9
var deterministic bool

// Arrays of losing triplets and winning quads, indexed
// by <x,y> coords of all pairs composing each of the quads
// or triplets. Makes deltaValue() a lot more efficient
var indexedLosingTriplets [5][5][][][]int
var indexedWinningQuads [5][5][][][]int

func main() {

	gameCountPtr := flag.Int("N", 10, "Number of games to play")
	maxDepthPtr := flag.Int("d", 10, "maximum lookahead depth")
	deterministicPtr := flag.Bool("D", false, "Play deterministically")
	printBoardPtr := flag.Bool("n", false, "Don't print board, just emit moves")
	printAnything := flag.Bool("P", false, "Print things")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	flag.Parse()

	*printBoardPtr = !*printBoardPtr

	deterministic = *deterministicPtr

	// Set up for use by deltaValue()
	for _, triplet := range losingTriplets {
		for _, pair := range triplet {
			indexedLosingTriplets[pair[0]][pair[1]] = append(indexedLosingTriplets[pair[0]][pair[1]], triplet)
		}
	}
	for _, quad := range winningQuads {
		for _, pair := range quad {
			indexedWinningQuads[pair[0]][pair[1]] = append(
				indexedWinningQuads[pair[0]][pair[1]], quad)
		}
	}

	maxDepth = *maxDepthPtr

	rand.Seed(time.Now().UTC().UnixNano())

	// fmt.Printf("Per-game summary (winner, number of moves in game, opening move, answering move)\n")
	for gameCount := 0; gameCount < *gameCountPtr; gameCount++ {
		var bd Board
		var endOfGame bool = false

		var moveRecord [25][2]int
		for _, z := range moveRecord {
			z[0] = -1
			z[1] = -1
		}

		setScores(*randomizeScores)

		moveCount := 0

		for !endOfGame {

			i, j, val := maximizerMove(&bd, maxDepth)
			bd[i][j] = MAXIMIZER
			if *printAnything {
				if *printBoardPtr {
					fmt.Printf("Maximizer move: %d %d (%d)\n", i, j, val)
					printBoard(&bd)
				} else {
					fmt.Printf("X %d %d (%d)\n", i, j, val)
				}
			}

			moveRecord[moveCount][0], moveRecord[moveCount][1] = i, j
			moveCount++

			endOfGame, _ = deltaValue(&bd, 0, i, j)
			if endOfGame {
				break
			}

			i, j, val = minimizerMove(&bd, maxDepth)
			bd[i][j] = MINIMIZER
			moveRecord[moveCount][0], moveRecord[moveCount][1] = i, j
			if *printAnything {
				if *printBoardPtr {
					fmt.Printf("Minimizer move: %d %d (%d)\n", i, j, val)
					printBoard(&bd)
				} else {
					fmt.Printf("O %d %d (%d)\n", i, j, val)
				}
			}
			moveCount++
			endOfGame, _ = deltaValue(&bd, 0, i, j)
		}
		winner := findWinner(&bd)
		var player string
		switch winner {
		case -1:
			player = "O"
		case 0:
			player = "C"
		case 1:
			player = "X"
		}
		fmt.Printf("%s\t%d", player, moveCount)
		for m := 0; m < moveCount; m++ {
			fmt.Printf("\t%d,%d", moveRecord[m][0], moveRecord[m][1])
		}
		fmt.Printf("\n")
	}

	os.Exit(0)
}

func findWinner(bd *Board) int {
	for _, quad := range winningQuads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return bd[quad[0][0]][quad[0][1]]
		}
	}

	for _, triplet := range losingTriplets {
		sum := bd[triplet[0][0]][triplet[0][1]]
		sum += bd[triplet[1][0]][triplet[1][1]]
		sum += bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			return -bd[triplet[0][0]][triplet[0][1]]
		}
	}

	return 0
}

// It turns out that you only have to look at
// the 4-in-a-rows that contain these 9 cells
// to check every 4-in-a-row. Similarly, you
// only need to check these 9 cells to check
// all the losing 3-in-a-row combos. You don't
// have to look at each and every cell.
var checkableCells [9][2]int = [9][2]int{
	{0, 2}, {1, 2}, {2, 0},
	{2, 1}, {2, 2}, {2, 3},
	{2, 4}, {3, 2}, {4, 2},
}

// Calculates and returns the value of the move (x,y)
// Only considers value gained or lost from the cell (x,y)
func deltaValue(bd *Board, ply int, x, y int) (stopRecursing bool, value int) {

	relevantQuads := indexedWinningQuads[x][y]
	for _, quad := range relevantQuads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return true, bd[quad[0][0]][quad[0][1]] * (WIN - ply)
		}
		if sum == 3 || sum == -3 {
			value += sum * 10
		}
	}

	relevantTriplets := indexedLosingTriplets[x][y]
	for _, triplet := range relevantTriplets {
		sum := bd[triplet[0][0]][triplet[0][1]]
		sum += bd[triplet[1][0]][triplet[1][1]]
		sum += bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			return true, -sum / 3 * (WIN - ply)
		}
	}

	// Give it a slight bias for those early
	// moves when all losing-triplets and winning-quads
	// are beyond the horizon.
	if (ply%2) == 0 {
		value += bd[x][y] * maximizerScores[x][y]
	} else {
		value += bd[x][y] * minimizerScores[x][y]
	}

	// If squava has a "cat game", then this is wrong. Cat
	// games could stop recursing here.
	stopRecursing = false
	if ply >= maxDepth {
		stopRecursing = true
	}

	return stopRecursing, value
}

// Calculates and returns the value of the entire board.
// It only looks at the cells in checkableCells[], so it
// doesn't double-count very many combos.
func wholeBoardValue(bd *Board, player int) (value int) {

	for _, cell := range checkableCells {
		relevantQuads := indexedWinningQuads[cell[0]][cell[1]]
		for _, quad := range relevantQuads {
			sum := bd[quad[0][0]][quad[0][1]]
			sum += bd[quad[1][0]][quad[1][1]]
			sum += bd[quad[2][0]][quad[2][1]]
			sum += bd[quad[3][0]][quad[3][1]]

			if sum == 4 || sum == -4 {
				return bd[quad[0][0]][quad[0][1]] * WIN
			}

			// Avoid 2 loops over checkableCells[] in the case of
			// no 4-in-a-row wins
			// Try to get into 3-of-winning-4 situtations
			if sum == 3 || sum == -3 {
				value += sum * 10
			}
		}

		relevantTriplets := indexedLosingTriplets[cell[0]][cell[1]]
		for _, triplet := range relevantTriplets {
			sum := bd[triplet[0][0]][triplet[0][1]]
			sum += bd[triplet[1][0]][triplet[1][1]]
			sum += bd[triplet[2][0]][triplet[2][1]]

			if sum == 3 || sum == -3 {
				return -sum / 3 * WIN
			}
		}
	}

	// Give it a slight bias for those early
	// moves when all losing-triplets and winning-quads
	// are beyond the horizon.
	var scores [][]int
	switch player {
	case MAXIMIZER:
		scores = maximizerScores
	case MINIMIZER:
		scores = minimizerScores
	}
	for i, row := range bd {
		for j, _ := range row {
			value += bd[i][j] * scores[i][j]
		}
	}
	return value
}

func alphaBeta(bd *Board, ply int, player int, alpha int, beta int, x int, y int, boardValue int) (value int) {

	stopRecursing, delta := deltaValue(bd, ply, x, y)

	boardValue += delta

	if stopRecursing {
		return boardValue
	}

	switch player {
	case MAXIMIZER:
		value = 2 * LOSS // Possible to score less than LOSS
		for i, row := range bd {
			for j, marker := range row {
				if marker == UNSET {
					bd[i][j] = MAXIMIZER
					n := alphaBeta(bd, ply+1, MINIMIZER, alpha, beta, i, j, boardValue)
					bd[i][j] = UNSET
					if n > value {
						value = n
					}
					if value > alpha {
						alpha = value
					}
					if beta <= alpha {
						return value
					}
				}
			}
		}
	case MINIMIZER:
		value = 2 * WIN // You can score greater than WIN
		for i, row := range bd {
			for j, marker := range row {
				if marker == UNSET {
					bd[i][j] = player
					n := alphaBeta(bd, ply+1, -player, alpha, beta, i, j, boardValue)
					bd[i][j] = UNSET
					if n < value {
						value = n
					}
					if value < beta {
						beta = value
					}
					if beta <= alpha {
						return value
					}
				}
			}
		}
	}

	return value
}

func printBoard(bd *Board) {
	fmt.Printf("   0 1 2 3 4\n")
	for i, row := range bd {
		fmt.Printf("%d  ", i)
		for _, v := range row {
			var marker string
			switch v {
			case MAXIMIZER:
				marker = "X"
			case MINIMIZER:
				marker = "O"
			case UNSET:
				marker = "_"
			}
			fmt.Printf("%s ", marker)
		}
		fmt.Printf("\n")
	}
}

var losingTriplets [][][]int = [][][]int{
	[][]int{[]int{0, 0}, []int{1, 0}, []int{2, 0}},
	[][]int{[]int{0, 0}, []int{0, 1}, []int{0, 2}},
	[][]int{[]int{0, 0}, []int{1, 1}, []int{2, 2}},
	[][]int{[]int{1, 0}, []int{2, 0}, []int{3, 0}},
	[][]int{[]int{1, 0}, []int{1, 1}, []int{1, 2}},
	[][]int{[]int{1, 0}, []int{2, 1}, []int{3, 2}},
	[][]int{[]int{2, 0}, []int{3, 0}, []int{4, 0}},
	[][]int{[]int{2, 0}, []int{2, 1}, []int{2, 2}},
	[][]int{[]int{2, 0}, []int{1, 1}, []int{0, 2}},
	[][]int{[]int{2, 0}, []int{3, 1}, []int{4, 2}},
	[][]int{[]int{3, 0}, []int{3, 1}, []int{3, 2}},
	[][]int{[]int{3, 0}, []int{2, 1}, []int{1, 2}},
	[][]int{[]int{4, 0}, []int{4, 1}, []int{4, 2}},
	[][]int{[]int{4, 0}, []int{3, 1}, []int{2, 2}},
	[][]int{[]int{0, 1}, []int{1, 1}, []int{2, 1}},
	[][]int{[]int{0, 1}, []int{0, 2}, []int{0, 3}},
	[][]int{[]int{0, 1}, []int{1, 2}, []int{2, 3}},
	[][]int{[]int{1, 1}, []int{2, 1}, []int{3, 1}},
	[][]int{[]int{1, 1}, []int{1, 2}, []int{1, 3}},
	[][]int{[]int{1, 1}, []int{2, 2}, []int{3, 3}},
	[][]int{[]int{2, 1}, []int{3, 1}, []int{4, 1}},
	[][]int{[]int{2, 1}, []int{2, 2}, []int{2, 3}},
	[][]int{[]int{2, 1}, []int{1, 2}, []int{0, 3}},
	[][]int{[]int{2, 1}, []int{3, 2}, []int{4, 3}},
	[][]int{[]int{3, 1}, []int{3, 2}, []int{3, 3}},
	[][]int{[]int{3, 1}, []int{2, 2}, []int{1, 3}},
	[][]int{[]int{4, 1}, []int{4, 2}, []int{4, 3}},
	[][]int{[]int{4, 1}, []int{3, 2}, []int{2, 3}},
	[][]int{[]int{0, 2}, []int{1, 2}, []int{2, 2}},
	[][]int{[]int{0, 2}, []int{0, 3}, []int{0, 4}},
	[][]int{[]int{0, 2}, []int{1, 3}, []int{2, 4}},
	[][]int{[]int{1, 2}, []int{2, 2}, []int{3, 2}},
	[][]int{[]int{1, 2}, []int{1, 3}, []int{1, 4}},
	[][]int{[]int{1, 2}, []int{2, 3}, []int{3, 4}},
	[][]int{[]int{2, 2}, []int{3, 2}, []int{4, 2}},
	[][]int{[]int{2, 2}, []int{2, 3}, []int{2, 4}},
	[][]int{[]int{2, 2}, []int{1, 3}, []int{0, 4}},
	[][]int{[]int{2, 2}, []int{3, 3}, []int{4, 4}},
	[][]int{[]int{3, 2}, []int{3, 3}, []int{3, 4}},
	[][]int{[]int{3, 2}, []int{2, 3}, []int{1, 4}},
	[][]int{[]int{4, 2}, []int{4, 3}, []int{4, 4}},
	[][]int{[]int{4, 2}, []int{3, 3}, []int{2, 4}},
	[][]int{[]int{0, 3}, []int{1, 3}, []int{2, 3}},
	[][]int{[]int{1, 3}, []int{2, 3}, []int{3, 3}},
	[][]int{[]int{2, 3}, []int{3, 3}, []int{4, 3}},
	[][]int{[]int{0, 4}, []int{1, 4}, []int{2, 4}},
	[][]int{[]int{1, 4}, []int{2, 4}, []int{3, 4}},
	[][]int{[]int{2, 4}, []int{3, 4}, []int{4, 4}},
}
var winningQuads [][][]int = [][][]int{
	[][]int{[]int{0, 0}, []int{1, 0}, []int{2, 0}, []int{3, 0}},
	[][]int{[]int{0, 0}, []int{0, 1}, []int{0, 2}, []int{0, 3}},
	[][]int{[]int{0, 0}, []int{1, 1}, []int{2, 2}, []int{3, 3}},
	[][]int{[]int{0, 1}, []int{1, 1}, []int{2, 1}, []int{3, 1}},
	[][]int{[]int{0, 1}, []int{0, 2}, []int{0, 3}, []int{0, 4}},
	[][]int{[]int{0, 1}, []int{1, 2}, []int{2, 3}, []int{3, 4}},
	[][]int{[]int{0, 2}, []int{1, 2}, []int{2, 2}, []int{3, 2}},
	[][]int{[]int{0, 3}, []int{1, 3}, []int{2, 3}, []int{3, 3}},
	[][]int{[]int{0, 4}, []int{1, 4}, []int{2, 4}, []int{3, 4}},
	[][]int{[]int{1, 0}, []int{2, 0}, []int{3, 0}, []int{4, 0}},
	[][]int{[]int{1, 0}, []int{1, 1}, []int{1, 2}, []int{1, 3}},
	[][]int{[]int{1, 0}, []int{2, 1}, []int{3, 2}, []int{4, 3}},
	[][]int{[]int{1, 1}, []int{2, 1}, []int{3, 1}, []int{4, 1}},
	[][]int{[]int{1, 1}, []int{1, 2}, []int{1, 3}, []int{1, 4}},
	[][]int{[]int{1, 1}, []int{2, 2}, []int{3, 3}, []int{4, 4}},
	[][]int{[]int{1, 2}, []int{2, 2}, []int{3, 2}, []int{4, 2}},
	[][]int{[]int{1, 3}, []int{2, 3}, []int{3, 3}, []int{4, 3}},
	[][]int{[]int{1, 4}, []int{2, 4}, []int{3, 4}, []int{4, 4}},
	[][]int{[]int{2, 0}, []int{2, 1}, []int{2, 2}, []int{2, 3}},
	[][]int{[]int{2, 1}, []int{2, 2}, []int{2, 3}, []int{2, 4}},
	[][]int{[]int{3, 0}, []int{3, 1}, []int{3, 2}, []int{3, 3}},
	[][]int{[]int{3, 0}, []int{2, 1}, []int{1, 2}, []int{0, 3}},
	[][]int{[]int{3, 1}, []int{3, 2}, []int{3, 3}, []int{3, 4}},
	[][]int{[]int{3, 1}, []int{2, 2}, []int{1, 3}, []int{0, 4}},
	[][]int{[]int{4, 0}, []int{4, 1}, []int{4, 2}, []int{4, 3}},
	[][]int{[]int{4, 0}, []int{3, 1}, []int{2, 2}, []int{1, 3}},
	[][]int{[]int{4, 1}, []int{4, 2}, []int{4, 3}, []int{4, 4}},
	[][]int{[]int{4, 1}, []int{3, 2}, []int{2, 3}, []int{1, 4}},
}

var maximizerScores [][]int = [][]int{
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
}
var minimizerScores [][]int = [][]int{
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
}

var scores [][]int = [][]int{
	/*
		[]int{3, 3, 0, 3, 3},
		[]int{3, 4, 1, 4, 3},
		[]int{0, 1, 0, 1, 0},
		[]int{3, 4, 1, 4, 3},
		[]int{3, 3, 0, 3, 3},
	*/
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
	[]int{0, 0, 0, 0, 0},
}


func maximizerMove(bd *Board, desiredDepth int) (int, int, int) {

	var moves [25][2]int
	var next int

	maxDepth = desiredDepth

	max := 2 * LOSS // A board can score less than LOSS

	boardValue := wholeBoardValue(bd, MINIMIZER)

	for i, row := range bd {
		for j, mark := range row {
			if mark == UNSET {
				bd[i][j] = MAXIMIZER
				val := alphaBeta(bd, 1, MINIMIZER, LOSS, WIN, i, j, boardValue)
				bd[i][j] = UNSET
				if val >= max {
					if val > max {
						max = val
						next = 0
					}
					moves[next][0] = i
					moves[next][1] = j
					next++
				}
			}
		}
	}

	r := 0
	if !deterministic {
		r = rand.Intn(next)
	}

	return moves[r][0], moves[r][1], max
}

func minimizerMove(bd *Board, desiredDepth int) (int, int, int) {

	var moves [25][2]int
	var next int

	maxDepth = desiredDepth

	min := 2 * WIN

	boardValue := wholeBoardValue(bd, MAXIMIZER)

	for i, row := range bd {
		for j, mark := range row {
			if mark == UNSET {
				bd[i][j] = MINIMIZER
				val := alphaBeta(bd, 1, MAXIMIZER, LOSS, WIN, i, j, boardValue)
				bd[i][j] = UNSET
				if val <= min {
					if val < min {
						min = val
						next = 0
					}
					moves[next][0] = i
					moves[next][1] = j
					next++
				}
			}
		}
	}

	r := 0
	if !deterministic {
		r = rand.Intn(next)
	}

	return moves[r][0], moves[r][1], min
}
func setScores(randomize bool) {
	if randomize {
		var vals [11]int = [11]int{-5, -4, -3 - 2, -1, 0, 1, 2, 3, 4, 5}
		for i, row := range scores {
			for j, _ := range row {
				maximizerScores[i][j] = vals[rand.Intn(11)]
				minimizerScores[i][j] = vals[rand.Intn(11)]
			}
		}
	} else {
		maximizerScores = [][]int{
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
		}
		minimizerScores = [][]int{
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
			[]int{0, 0, 0, 0, 0},
		}
	}
}
