package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

type Board [5][5]int

const (
	WIN       = 10000
	LOSS      = -10000
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

var maxDepth int = 10 // initializing to 10 not a mistake

// Arrays of losing triplets and winning quads, indexed
// by <x,y> coords of all pairs composing each of the quads
// or triplets. Makes deltaValue() a lot more efficient
var indexedLosingTriplets [5][5][][][]int
var indexedWinningQuads [5][5][][][]int

func main() {

	humanFirstPtr := flag.Bool("H", true, "Human takes first move")
	computerFirstPtr := flag.Bool("C", false, "Computer takes first move")
	maxDepthPtr := flag.Int("d", 10, "maximum lookahead depth")
	deterministic := flag.Bool("D", false, "Play deterministically")
	printBoardPtr := flag.Bool("n", false, "Don't print board, just emit moves")
	firstMovePtr := flag.String("M", "", "Tell computer to make this first move (x,y)")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	flag.Parse()

	*printBoardPtr = !*printBoardPtr

	// Set up for use by staticValue()
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

	rand.Seed(time.Now().UTC().UnixNano())

	setScores(*randomizeScores)

	var humanFirst bool = *humanFirstPtr
	if *computerFirstPtr {
		humanFirst = false
	}

	var bd Board

	if *firstMovePtr != "" {
		var x1, y1 int
		fmt.Sscanf(*firstMovePtr, "%d,%d", &x1, &y1)
		fmt.Printf("My move: %d %d\n", x1, y1)
		humanFirst = true
		bd[x1][y1] = MAXIMIZER
		printBoard(&bd)
	}

	var endOfGame bool = false

	moveCounter := 0

	for !endOfGame {
		setDepth(moveCounter, *maxDepthPtr)

		var l, m int
		if humanFirst {
			l, m = readMove(&bd, *printBoardPtr)
			bd[l][m] = MINIMIZER
			endOfGame, _ = staticValue(&bd, 0)
			moveCounter++
		}

		if endOfGame {
			break
		}

		humanFirst = true

		a, b, score := chooseMove(&bd, *deterministic)

		if a < 0 {
			break // Cat gets the game
		}

		bd[a][b] = MAXIMIZER
		moveCounter++

		if *printBoardPtr {
			fmt.Printf("My move: %d %d (%d)\n", a, b, score)
			printBoard(&bd)
		} else {
			fmt.Printf("%d %d\n", a, b)
		}

		endOfGame, _ = staticValue(&bd, 0)

	}

	if *printBoardPtr {
		var phrase string
		switch findWinner(&bd) {
		case MAXIMIZER:
			phrase = "\nX wins\n"
		case UNSET:
			phrase = "\nCat wins\n"
		case MINIMIZER:
			phrase = "\nO wins\n"
		}
		fmt.Printf(phrase)

		printBoard(&bd)
	}

	os.Exit(0)
}

func setDepth(moveCounter int, endGameDepth int) {
	if moveCounter < 4 {
		maxDepth = 6
	}
	if moveCounter > 3 {
		maxDepth = 8
	}
	if moveCounter > 10 {
		maxDepth = endGameDepth
	}
}

// Choose computer's next move: return x,y coords of move and its score.
func chooseMove(bd *Board, deterministic bool) (int, int, int) {

	var moves [25][2]int
	var next int

	max := 2 * LOSS // A board can score less than LOSS

	for i, row := range bd {
		for j, mark := range row {
			if mark == UNSET {
				bd[i][j] = MAXIMIZER
				val := -negaScout(bd, 1, MINIMIZER, 2*LOSS, 2*WIN)
				bd[i][j] = UNSET
				fmt.Printf("     <%d,%d>  %d\n", i, j, val)
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

	if next == 0 {
		// Loop over all 25 cells couldn't find any
		// empty cells. Cat got the game.
		return -1, -1, 0
	}

	r := 0
	if !deterministic {
		r = rand.Intn(next)
	}

	fmt.Printf("  <%d,%d>  %d\n", moves[r][0], moves[r][1], max)
	return moves[r][0], moves[r][1], max
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
func staticValue(bd *Board, ply int) (stopRecursing bool, value int) {

	for _, cell := range checkableCells {
		relevantQuads := indexedWinningQuads[cell[0]][cell[1]]
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
	}

	for _, cell := range checkableCells {
		relevantTriplets := indexedLosingTriplets[cell[0]][cell[1]]
		for _, triplet := range relevantTriplets {
			sum := bd[triplet[0][0]][triplet[0][1]]
			sum += bd[triplet[1][0]][triplet[1][1]]
			sum += bd[triplet[2][0]][triplet[2][1]]

			if sum == 3 || sum == -3 {
				return true, -sum / 3 * (WIN - ply)
			}
		}
	}

	// Give it a slight bias for those early
	// moves when all losing-triplets and winning-quads
	// are beyond the horizon.
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			value += bd[x][y] * scores[x][y]
		}
	}

	// If squava has a "cat game", then this is wrong. Cat
	// games could stop recursing here.
	stopRecursing = false
	if ply >= maxDepth {
		stopRecursing = true
	}

	return stopRecursing, value
}

func negaScout(bd *Board, ply int, player int, alpha int, beta int) (value int) {

	stopRecursing, boardValue := staticValue(bd, ply)
	if stopRecursing {
		return player*boardValue
	}

	b := beta
	later := false

	for i, row := range bd {
		for j, marker := range row {
			if marker == UNSET {
				bd[i][j] = player
				t := -negaScout(bd, ply+1, -player, -b, -alpha)
				if t > alpha && t < beta && later {
					t = -negaScout(bd, ply+1, -player, -beta, -alpha)
				}
				bd[i][j] = UNSET
				if t > alpha {
					alpha = t
				}
				if alpha >= beta {
					return alpha
				}
				b = alpha + 1
			}
			later = true
		}
	}
	return alpha
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

var scores [5][5]int

func readMove(bd *Board, print bool) (x, y int) {
	readMove := false
	for !readMove {
		if print {
			fmt.Printf("Your move: ")
		}
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
			if print {
				fmt.Printf("Choose two numbers between 0 and 4, try again\n")
			}
		case bd[x][y] == 0:
			readMove = true
		case bd[x][y] != 0:
			if print {
				fmt.Printf("Cell (%d, %d) already occupied, try again\n", x, y)
			}
		}
	}
	return x, y
}

func setScores(randomize bool) {
	if randomize {
		var vals [11]int = [11]int{-5, -4, -3 - 2, -1, 0, 1, 2, 3, 4, 5}
		for i, row := range scores {
			for j, _ := range row {
				scores[i][j] = vals[rand.Intn(11)]
			}
		}
	} else {
		scores = [5][5]int{
			[5]int{3, 3, 0, 3, 3},
			[5]int{3, 4, 1, 4, 3},
			[5]int{0, 1, 0, 1, 0},
			[5]int{3, 4, 1, 4, 3},
			[5]int{3, 3, 0, 3, 3},
		}
	}
}
