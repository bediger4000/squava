// Calculate the value for first moves
// Only does 6 actual first moves:
//     0 1 2 3 4
//    0  . . . _ _ 
//    1  _ . . _ _ 
//    2  _ _ . _ _ 
//    3  _ _ _ _ _ 
//    4  _ _ _ _ _ 
//
// <0,0>, <0,1>, <0,2>, <1,1>, <1,2> and <2,2>
// Every other first move is a reflection or rotation of
// the board away from those 6.

package main

import (
	"flag"
	"fmt"
	"os"
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

	maxDepthPtr := flag.Int("d", 9, "maximum lookahead depth")
	flag.Parse()

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

	maxDepth = *maxDepthPtr

	var bd Board

	for _, cell := range uniqueCells {
		i, j := cell[0], cell[1]
		bd[i][j] = MAXIMIZER
		leafNodes = 0
		before := time.Now()
		val := alphaBeta(&bd, 1, MINIMIZER, LOSS, WIN, i, j, 0)
		after := time.Now()
		bd[i][j] = UNSET
		fmt.Printf("<%d,%d>\t%d (%d) [%d]\t%v\n", i, j, val, scores[i][j], leafNodes, after.Sub(before))
	}

	os.Exit(0)
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

// nextPlayer makes move in this invocation.
// previous player (-nextPlayer) just made the move <x,y>,
// bd has value boardValue not including that move.
func alphaBeta(bd *Board, ply int, nextPlayer int, alpha int, beta int, x int, y int, boardValue int) (value int) {

	relevantQuads := indexedWinningQuads[x][y]
	delta := 0
	for _, quad := range relevantQuads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			leafNodes++
			return bd[quad[0][0]][quad[0][1]] * (WIN - ply)
		}
		if sum == 3 || sum == -3 {
			delta += sum * 10
		}
	}

	relevantTriplets := indexedLosingTriplets[x][y]
	for _, triplet := range relevantTriplets {
		sum := bd[triplet[0][0]][triplet[0][1]]
		sum += bd[triplet[1][0]][triplet[1][1]]
		sum += bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			leafNodes++
			return -sum / 3 * (WIN - ply)
		}
	}

	delta += bd[x][y] * scores[x][y]

	boardValue += delta

	if ply == maxDepth {
		leafNodes++
		return boardValue
	}

	switch nextPlayer {
	case MAXIMIZER:
		value = LOSS
		for _, cell := range orderedCells {
			i, j := cell[0], cell[1]
			marker := bd[i][j]
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
					break
				}
			}
		}
		leafNodes++
		return value
	case MINIMIZER:
		value = WIN
		for _, cell := range orderedCells {
			i, j := cell[0], cell[1]
			marker := bd[i][j]
			if marker == UNSET {
				bd[i][j] = MINIMIZER
				n := alphaBeta(bd, ply+1, MAXIMIZER, alpha, beta, i, j, boardValue)
				bd[i][j] = UNSET
				if n < value {
					value = n
				}
				if value < beta {
					beta = value
				}
				if beta <= alpha {
					break
				}
			}
		}
		leafNodes++
		return value
	}

	leafNodes++
	return value
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

var uniqueCells [][]int = [][]int{
	[]int{0, 0},
	[]int{0, 1},
	[]int{0, 2},
	[]int{1, 1},
	[]int{1, 2},
	[]int{2, 2},
}

var orderedCells [25][2]int = [25][2]int{
	[2]int{1, 1},
	[2]int{1, 3},
	[2]int{3, 1},
	[2]int{3, 3},

	[2]int{2, 2},

	[2]int{1, 2},
	[2]int{2, 1},
	[2]int{2, 3},
	[2]int{3, 2},

	[2]int{0, 1},
	[2]int{0, 2},
	[2]int{0, 3},

	[2]int{1, 0},
	[2]int{2, 0},
	[2]int{3, 0},

	[2]int{1, 4},
	[2]int{2, 4},
	[2]int{3, 4},

	[2]int{4, 1},
	[2]int{4, 2},
	[2]int{4, 3},

	[2]int{0, 0},
	[2]int{0, 4},
	[2]int{4, 0},
	[2]int{4, 4},
}

var scores [][]int = [][]int{
	[]int{3, 3, 0, 3, 3},
	[]int{3, 4, 1, 4, 3},
	[]int{0, 1, 0, 1, 2},
	[]int{3, 4, 1, 4, 3},
	[]int{3, 3, 0, 3, 3},
}
