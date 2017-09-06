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

var leafNodeCount int = 0
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

		leafNodeCount = 0
		a, b, score := chooseMove(&bd, *deterministic)

		if a < 0 {
			break // Cat gets the game
		}

		bd[a][b] = MAXIMIZER
		moveCounter++

		if *printBoardPtr {
			fmt.Printf("My move: %d %d (%d) [%d]\n", a, b, score, leafNodeCount)
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
var initialOrderedMoves [25][2]int = [25][2]int{
	{1, 1}, {1, 3}, {3, 3}, {3, 1},
	{0, 1}, {0, 3}, {1, 4}, {3, 4}, {4, 3}, {4, 1}, {3, 0}, {1, 0},
	{0, 0}, {0, 4}, {4, 4}, {4, 0},
	{2, 2},
	{1, 2}, {2, 3}, {3, 2}, {2, 1},
	{0, 2}, {2, 0}, {2, 4}, {4, 2},
}

var orderedMoves [3][][2]int

func reorderMoves(bd *Board) {
	var goodCells [][2]int
	var badCells [][2]int
	var dullCells [][2]int
	unsetCount := 0
	for _, cell := range initialOrderedMoves {
		if bd[cell[0]][cell[1]] == UNSET {
			unsetCount++
			interesting := false
			relevantQuads := indexedWinningQuads[cell[0]][cell[1]]
			for _, quad := range relevantQuads {
				sum := bd[quad[0][0]][quad[0][1]]
				sum += bd[quad[1][0]][quad[1][1]]
				sum += bd[quad[2][0]][quad[2][1]]
				sum += bd[quad[3][0]][quad[3][1]]

				if sum == 3 {
					// It's a hole in a potential 4-in-a-row
					goodCells = append(goodCells, cell)
					interesting = true
					break
				}
				if sum == 2 {
					goodCells = append(goodCells, cell)
					interesting = true
					break
				}
				if sum == -3 {
					// It's a hole in a potential 4-in-a-row
					badCells = append(badCells, cell)
					interesting = true
					break
				}
				if sum == -2 {
					badCells = append(badCells, cell)
					interesting = true
					break
				}
			}
			if !interesting {
				relevantTriplets := indexedLosingTriplets[cell[0]][cell[1]]
				for _, triplet := range relevantTriplets {
					sum := bd[triplet[0][0]][triplet[0][1]]
					sum += bd[triplet[1][0]][triplet[1][1]]
					sum += bd[triplet[2][0]][triplet[2][1]]

					if sum == -2 {
						goodCells = append(goodCells, cell)
						interesting = true
						break
					}
					if sum == 2 {
						badCells = append(badCells, cell)
						interesting = true
						break
					}
				}
			}
			if !interesting {
				dullCells = append(dullCells, cell)
			}
		}
	}

	goodCells = append(goodCells, dullCells...)
	goodCells = append(goodCells, badCells...)
	badCells = append(badCells, dullCells...)
	badCells = append(badCells, goodCells...)

	orderedMoves[0] = badCells
	orderedMoves[2] = goodCells
}

func chooseMove(bd *Board, deterministic bool) (int, int, int) {

	var moves = MoveKeeper{next: 0, max: 3 * LOSS}

	reorderMoves(bd)

	beta := 2 * WIN
	alpha := 2 * LOSS
	score := 3 * LOSS
	n := beta

	for _, cell := range orderedMoves[2] {
		i, j := cell[0], cell[1]
		if bd[i][j] == UNSET {
			bd[i][j] = MAXIMIZER
			cur := -negaScout(bd, 1, MINIMIZER, -n, -alpha)
			if cur > score {
				if n == beta {
					score = cur
				} else {
					score = -negaScout(bd, 1, MINIMIZER, -beta, -cur)
				}
			}
			bd[i][j] = UNSET
			if score > alpha {
				alpha = score
			}
			moves.setMove(i, j, alpha)
			if alpha >= beta {
				break
			}
			n = alpha + 1
		}
	}

	return moves.chooseMove(deterministic)
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

var deadlyQuads [4][4][2]int = [4][4][2]int{
	{{1, 0}, {2, 1}, {3, 2}, {4, 3}},
	{{4, 1}, {3, 2}, {2, 3}, {1, 4}},
	{{0, 1}, {1, 2}, {2, 3}, {3, 4}},
	{{3, 0}, {2, 1}, {1, 2}, {0, 3}},
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

func staticValue(bd *Board, ply int) (stopRecursing bool, value int) {

	leafNodeCount++

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

	for _, quad := range deadlyQuads {
		outer := bd[quad[0][0]][quad[0][1]] + bd[quad[3][0]][quad[3][1]]
		inner := bd[quad[1][0]][quad[1][1]] + bd[quad[2][0]][quad[2][1]]

		if (inner == 2 || inner == -2) && outer == 0 {
			value += -inner * 5
		}
	}

	if value == 0 {
		// Give it a slight bias for those early
		// moves when all losing-triplets and winning-quads
		// are beyond the horizon.
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				value += bd[x][y] * scores[x][y]
			}
		}
	}

	// If squava has a "cat game", then this is wrong. Cat
	// games could stop recursing here.
	stopRecursing = false
	if ply > maxDepth {
		stopRecursing = true
	}

	return stopRecursing, value
}

func negaScout(bd *Board, ply int, player int, alpha int, beta int) (value int) {

	stopRecursing, boardValue := staticValue(bd, ply)
	if stopRecursing {
		return player * boardValue
	}

	score := 3 * LOSS
	n := beta

	for _, cell := range orderedMoves[player+1] {
		i, j := cell[0], cell[1]
		if bd[i][j] == UNSET {
			bd[i][j] = player
			cur := -negaScout(bd, ply+1, -player, -n, -alpha)
			if cur > score {
				if n == beta || ply == maxDepth-2 {
					score = cur
				} else {
					score = -negaScout(bd, ply+1, -player, -beta, -cur)
				}
			}
			bd[i][j] = UNSET
			if score > alpha {
				alpha = score
			}
			if alpha >= beta {
				break
			}
			n = alpha + 1
		}
	}

	return score
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

// Struct and 2 functions to encapsulate tracking of
// best possible move.

type MoveKeeper struct {
	moves [25][2]int
	next  int // index into moves[]
	max   int
}

func (p *MoveKeeper) setMove(a, b int, value int) {
	if value >= p.max {
		if value > p.max {
			p.max = value
			p.next = 0
		}
		p.moves[p.next][0] = a
		p.moves[p.next][1] = b
		p.next++
	}
}

func (p *MoveKeeper) chooseMove(deterministic bool) (x, y int, value int) {

	if p.next == 0 {
		// Loop over all 25 cells couldn't find any
		// empty cells. Cat got the game.
		return -1, -1, 0
	}

	r := 0
	if !deterministic {
		r = rand.Intn(p.next)
	}

	return p.moves[r][0], p.moves[r][1], p.max
}
