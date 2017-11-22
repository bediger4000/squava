package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Board [5][5]int

const WIN = 10000
const LOSS = -10000
const MAXIMIZER = 1
const MINIMIZER = -1
const UNSET = 0

// Arrays of losing triplets and winning quads, indexed
// by <x,y> coords of all pairs composing each of the quads
// or triplets. Makes alphaBeta() a lot more efficient
var indexedLosingTriplets [5][5][][][]int
var indexedWinningQuads [5][5][][][]int

var maxDepth int
var leafNodes int

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("./probe depth [m,n [m,n ...]]\n")
		os.Exit(1)
	}

	// Set up for use by value-calculation section of alphaBeta()
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

	maxDepth, _ = strconv.Atoi(os.Args[1])

	var bd Board

	var moveSequence [][2]int
	var nextPly int
	var nextPlayer int = MAXIMIZER
	var cell [2]int

	for _, str := range os.Args[2:] {
		mn := strings.Split(str, ",")
		m, _ := strconv.Atoi(mn[0])
		n, _ := strconv.Atoi(mn[1])
		cell[0], cell[1] = m, n
		moveSequence = append(moveSequence, cell)
		bd[m][n] = nextPlayer
		nextPly++
		nextPlayer = -nextPlayer
	}

	fmt.Printf("next ply: %d\nnext player: %d\n", nextPly, nextPlayer)
	player := MAXIMIZER
	for ply, cell := range moveSequence {
		fmt.Printf("Ply %d, player %d, move <%d,%d>\n",
			ply, player, cell[0], cell[1])
		player = -player
	}
	fmt.Printf("\n\n")

	printBoard(&bd)

	var bestValue int
	var bestMoves [25][2]int
	var bestNext int
	var totalLeafNodes int

	switch nextPlayer {
	case MAXIMIZER:
		bestValue = 2 * LOSS
	case MINIMIZER:
		bestValue = 2 * WIN
	}

	for i, row := range bd {
		for j, mark := range row {
			if mark == UNSET {
				bd[i][j] = nextPlayer
				stopRecursing, val := deltaValue(maxDepth, &bd, 0, i, j, bestValue)
				leafNodes = 0
				before := time.Now()
				if !stopRecursing {
					val, leafNodes = alphaBeta(maxDepth, &bd, 0, -nextPlayer, 2*LOSS, 2*WIN, i, j, val)
				} else {
					leafNodes = 1
				}
				switch nextPlayer {
				case MAXIMIZER:
					if val >= bestValue {
						if val > bestValue {
							bestValue = val
							bestNext = 0
						}
						bestMoves[bestNext][0] = i
						bestMoves[bestNext][1] = j
						bestNext++
					}
				case MINIMIZER:
					if val <= bestValue {
						if val < bestValue {
							bestValue = val
							bestNext = 0
						}
						bestMoves[bestNext][0] = i
						bestMoves[bestNext][1] = j
						bestNext++
					}
				}
				after := time.Now()
				bd[i][j] = UNSET
				totalLeafNodes += leafNodes
				fmt.Printf("<%d,%d>\t%d [%d]\t%v\n", i, j, val, leafNodes, after.Sub(before))
			}
		}
	}

	fmt.Printf("Best move(s), have value %d:\n", bestValue)
	for i := 0; i < bestNext; i++ {
		fmt.Printf("\t<%d,%d>\n", bestMoves[i][0], bestMoves[i][1])
	}
	fmt.Printf("%d total leafnodes\n", totalLeafNodes)

	fmt.Printf("%s  %d  ", os.Args[0], maxDepth)
	for _, cell := range moveSequence {
		fmt.Printf(" %d,%d", cell[0], cell[1])
	}
	fmt.Printf(" %d,%d", bestMoves[0][0], bestMoves[0][1])
	fmt.Printf("\n")

	bd[bestMoves[0][0]][bestMoves[0][1]] = nextPlayer

	printBoard(&bd)
	fmt.Printf("\n")

	os.Exit(0)
}

// Calculates and returns the value of the move (x,y)
// Only considers value gained or lost from the cell (x,y)
func deltaValue(maxPlies int, bd *Board, ply int, x, y int, currentValue int) (stopRecursing bool, value int) {

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

	for _, triplet := range no2 {
		for _, pair := range triplet {
			if x == pair[0] && y == pair[1] {
				sum := bd[triplet[0][0]][triplet[0][1]]
				sum += bd[triplet[1][0]][triplet[1][1]]
				sum += bd[triplet[2][0]][triplet[2][1]]
				if sum == 2 || sum == -2 {
					value += bd[x][y] * -100
				}
				break
			}
		}
	}

	for _, quad := range noMiddle2 {
		player := bd[x][y]
		if (x == quad[1][0] && y == quad[1][1] && player == bd[quad[2][0]][quad[2][1]]) ||
			(x == quad[2][0] && y == quad[2][1] && player == bd[quad[1][0]][quad[1][1]]) {

			sum := bd[quad[0][0]][quad[0][1]]
			sum += bd[quad[1][0]][quad[1][1]]
			sum += bd[quad[2][0]][quad[2][1]]
			sum += bd[quad[3][0]][quad[3][1]]

			if sum == 2 || sum == -2 {
				value += player * -100
			}
		}
	}

	// Give it a slight bias for those early
	// moves when all losing-triplets and winning-quads
	// are beyond the horizon.
	value += bd[x][y] * scores[x][y]

	// If squava has a "cat game", then this is wrong. Cat
	// games could stop recursing here.
	stopRecursing = false
	if ply == maxPlies {
		stopRecursing = true
		value += currentValue
	}

	return stopRecursing, value
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

// player makes move in this invocation.
// previous player (-player) just made the move <x,y>,
// bd has value boardValue.
func alphaBeta(maxDepth int, bd *Board, ply int, player int, alpha int, beta int, x int, y int, boardValue int) (value int, leafNodes int) {

	leafNodes = 0

	switch player {
	case MAXIMIZER:
		value = 2 * LOSS // Possible to score less than LOSS
		for i, row := range bd {
			for j, marker := range row {
				if marker == UNSET {
					bd[i][j] = MAXIMIZER
					stopRecursing, delta := deltaValue(maxDepth, bd, ply, x, y, boardValue)
					if stopRecursing {
						bd[i][j] = UNSET
						return delta, leafNodes + 1
					}
					n, leaves := alphaBeta(maxDepth, bd, ply+1, MINIMIZER, alpha, beta, i, j, boardValue+delta)
					bd[i][j] = UNSET
					leafNodes += leaves
					if n > value {
						value = n
					}
					if value > alpha {
						alpha = value
					}
					if beta <= alpha {
						return value, leafNodes + 1
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
					stopRecursing, delta := deltaValue(maxDepth, bd, ply, x, y, boardValue)
					if stopRecursing {
						bd[i][j] = UNSET
						return delta, leafNodes + 1
					}
					n, leaves := alphaBeta(maxDepth, bd, ply+1, -player, alpha, beta, i, j, boardValue+delta)
					bd[i][j] = UNSET
					leafNodes += leaves
					if n < value {
						value = n
					}
					if value < beta {
						beta = value
					}
					if beta <= alpha {
						return value, leafNodes + 1
					}
				}
			}
		}
	}

	return value, leafNodes
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

// 4-in-a-row where you don't want to have the middle 2
var noMiddle2 = [4][4][2]int{
	{{3, 0}, {2, 1}, {1, 2}, {0, 3}},
	{{1, 0}, {2, 1}, {3, 2}, {4, 3}},
	{{0, 1}, {1, 2}, {2, 3}, {3, 4}},
	{{1, 4}, {2, 3}, {3, 2}, {4, 1}},
}

// 3-in-a-row where you don't want any 2 plus a blank
var no2 = [4][3][2]int{

	{{2, 0}, {1, 1}, {0, 2}},
	{{0, 2}, {1, 3}, {2, 4}},
	{{4, 2}, {3, 3}, {2, 4}},
	{{4, 2}, {3, 1}, {2, 0}},
}

var losingTriplets [][][]int = [][][]int{
	{{0, 0}, {1, 0}, {2, 0}},
	{{0, 0}, {0, 1}, {0, 2}},
	{{0, 0}, {1, 1}, {2, 2}},
	{{1, 0}, {2, 0}, {3, 0}},
	{{1, 0}, {1, 1}, {1, 2}},
	{{1, 0}, {2, 1}, {3, 2}},
	{{2, 0}, {3, 0}, {4, 0}},
	{{2, 0}, {2, 1}, {2, 2}},
	{{2, 0}, {1, 1}, {0, 2}},
	{{2, 0}, {3, 1}, {4, 2}},
	{{3, 0}, {3, 1}, {3, 2}},
	{{3, 0}, {2, 1}, {1, 2}},
	{{4, 0}, {4, 1}, {4, 2}},
	{{4, 0}, {3, 1}, {2, 2}},
	{{0, 1}, {1, 1}, {2, 1}},
	{{0, 1}, {0, 2}, {0, 3}},
	{{0, 1}, {1, 2}, {2, 3}},
	{{1, 1}, {2, 1}, {3, 1}},
	{{1, 1}, {1, 2}, {1, 3}},
	{{1, 1}, {2, 2}, {3, 3}},
	{{2, 1}, {3, 1}, {4, 1}},
	{{2, 1}, {2, 2}, {2, 3}},
	{{2, 1}, {1, 2}, {0, 3}},
	{{2, 1}, {3, 2}, {4, 3}},
	{{3, 1}, {3, 2}, {3, 3}},
	{{3, 1}, {2, 2}, {1, 3}},
	{{4, 1}, {4, 2}, {4, 3}},
	{{4, 1}, {3, 2}, {2, 3}},
	{{0, 2}, {1, 2}, {2, 2}},
	{{0, 2}, {0, 3}, {0, 4}},
	{{0, 2}, {1, 3}, {2, 4}},
	{{1, 2}, {2, 2}, {3, 2}},
	{{1, 2}, {1, 3}, {1, 4}},
	{{1, 2}, {2, 3}, {3, 4}},
	{{2, 2}, {3, 2}, {4, 2}},
	{{2, 2}, {2, 3}, {2, 4}},
	{{2, 2}, {1, 3}, {0, 4}},
	{{2, 2}, {3, 3}, {4, 4}},
	{{3, 2}, {3, 3}, {3, 4}},
	{{3, 2}, {2, 3}, {1, 4}},
	{{4, 2}, {4, 3}, {4, 4}},
	{{4, 2}, {3, 3}, {2, 4}},
	{{0, 3}, {1, 3}, {2, 3}},
	{{1, 3}, {2, 3}, {3, 3}},
	{{2, 3}, {3, 3}, {4, 3}},
	{{0, 4}, {1, 4}, {2, 4}},
	{{1, 4}, {2, 4}, {3, 4}},
	{{2, 4}, {3, 4}, {4, 4}},
}
var winningQuads [][][]int = [][][]int{
	{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
	{{0, 0}, {0, 1}, {0, 2}, {0, 3}},
	{{0, 0}, {1, 1}, {2, 2}, {3, 3}},
	{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
	{{0, 1}, {0, 2}, {0, 3}, {0, 4}},
	{{0, 1}, {1, 2}, {2, 3}, {3, 4}},
	{{0, 2}, {1, 2}, {2, 2}, {3, 2}},
	{{0, 3}, {1, 3}, {2, 3}, {3, 3}},
	{{0, 4}, {1, 4}, {2, 4}, {3, 4}},
	{{1, 0}, {2, 0}, {3, 0}, {4, 0}},
	{{1, 0}, {1, 1}, {1, 2}, {1, 3}},
	{{1, 0}, {2, 1}, {3, 2}, {4, 3}},
	{{1, 1}, {2, 1}, {3, 1}, {4, 1}},
	{{1, 1}, {1, 2}, {1, 3}, {1, 4}},
	{{1, 1}, {2, 2}, {3, 3}, {4, 4}},
	{{1, 2}, {2, 2}, {3, 2}, {4, 2}},
	{{1, 3}, {2, 3}, {3, 3}, {4, 3}},
	{{1, 4}, {2, 4}, {3, 4}, {4, 4}},
	{{2, 0}, {2, 1}, {2, 2}, {2, 3}},
	{{2, 1}, {2, 2}, {2, 3}, {2, 4}},
	{{3, 0}, {3, 1}, {3, 2}, {3, 3}},
	{{3, 0}, {2, 1}, {1, 2}, {0, 3}},
	{{3, 1}, {3, 2}, {3, 3}, {3, 4}},
	{{3, 1}, {2, 2}, {1, 3}, {0, 4}},
	{{4, 0}, {4, 1}, {4, 2}, {4, 3}},
	{{4, 0}, {3, 1}, {2, 2}, {1, 3}},
	{{4, 1}, {4, 2}, {4, 3}, {4, 4}},
	{{4, 1}, {3, 2}, {2, 3}, {1, 4}},
}

var uniqueCells [][]int = [][]int{
	{0, 0},
	{0, 1},
	{0, 2},
	{1, 1},
	{1, 2},
	{2, 2},
}

var orderedCells [25][2]int = [25][2]int{
	{1, 1},
	{1, 3},
	{3, 1},
	{3, 3},

	{2, 2},

	{1, 2},
	{2, 1},
	{2, 3},
	{3, 2},

	{0, 1},
	{0, 2},
	{0, 3},

	{1, 0},
	{2, 0},
	{3, 0},

	{1, 4},
	{2, 4},
	{3, 4},

	{4, 1},
	{4, 2},
	{4, 3},

	{0, 0},
	{0, 4},
	{4, 0},
	{4, 4},
}

var scores [][]int = [][]int{
	{3, 3, 0, 3, 3},
	{3, 4, 1, 4, 3},
	{0, 1, 0, 1, 0},
	{3, 4, 1, 4, 3},
	{3, 3, 0, 3, 3},
}
