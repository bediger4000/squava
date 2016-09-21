// Deep evaluation of response to first move of the game.
// Potentially generate a "book" for the first 2 moves.
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
// or triplets. Makes deltaValue() a lot more efficient
var indexedLosingTriplets [5][5][][][]int
var indexedWinningQuads [5][5][][][]int

var maxDepth int
var leafNodes int

func main() {

	maxDepthPtr := flag.Int("d", 9, "maximum lookahead depth")
	flag.Parse()

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

	var bd Board

	var computedMoves [6][5][5]int

	// Set the first move to 1 for all 6 computedMoves[][][]
	// board representations. Other cells in  computedMoves[][][]
	// will have -1 in them (for the human's response) *if* they've
	// previously had an alpha beta value computed.
	for moveIndex, cell := range firstMoves {
		computedMoves[moveIndex][cell[0]][cell[1]] = MAXIMIZER
	}

	var totalResponseMoves int
	var uniqMovesComputed int
	var duplicateMoves int
	var sumLeafNodes int

	for firstMoveIndex, cell := range firstMoves {

		// first move: x, y
		x, y := cell[0], cell[1]

		for p := 0; p < 5; p++ {
			for q := 0; q < 5; q++ {
				if p != x || q != y {
					totalResponseMoves++
					// 2nd, reply move: p,q
					// <a,b> computer's move, <p,q> human's response
					// See if any reflected or rotated boards match
					matchedPrevious := false
					for _, m := range Matrices {
						imageInitialCell := m[x][y]
						imageResponseCell := m[p][q]
						// a,b - reflected or rotated initial move
						// c,d - reflected or rotated response move
						a, b := imageInitialCell[0], imageInitialCell[1]
						c, d := imageResponseCell[0], imageResponseCell[1]
						for _, board := range computedMoves {
							if board[a][b] == MAXIMIZER {
								if board[c][d] == MINIMIZER {
									matchedPrevious = true
									// This move image matches a
									// previously computed move.
									break
								}
							}
						}
						if matchedPrevious {
							break
						}
					}
					if !matchedPrevious {
						uniqMovesComputed++
						// Not previously computed
						computedMoves[firstMoveIndex][p][q] = MINIMIZER
						bd[x][y] = MAXIMIZER // ply 0
						bd[p][q] = MINIMIZER // ply 1
						delta := scores[x][y]
						leafNodes = 0
						before := time.Now()
						val := alphaBeta(&bd, 2, MAXIMIZER, LOSS, WIN, p, q, delta)
						after := time.Now()
						bd[x][y] = UNSET
						bd[p][q] = UNSET
						fmt.Printf("<%d,%d>:<%d,%d>\t%d (%d) [%d]\t%v\n", x, y, p, q, val, delta, leafNodes, after.Sub(before))
						sumLeafNodes += leafNodes
					} else {
						duplicateMoves++
					}
				}
			}
		}
		fmt.Printf("Leaf nodes visited: %d\n", sumLeafNodes)
		fmt.Printf("Unique moves computed: %d\n", uniqMovesComputed)
		fmt.Printf("Duplicated moves: %d\n", duplicateMoves)
		fmt.Printf("\n")
		sumLeafNodes = 0
		uniqMovesComputed = 0
		duplicateMoves = 0
	}

	os.Exit(0)
}

// nextPlayer makes move in this invocation.
// previous player (-nextPlayer) just made the move <x,y>,
// and board has value boardValue not including that move.
func alphaBeta(bd *Board, ply int, nextPlayer int, alpha int, beta int, x int, y int, boardValue int) (value int) {

	delta := 0
	relevantQuads := indexedWinningQuads[x][y]
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

// All the three-in-a-row triplets of coordinates,
// listed systematically.
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

// All the four-in-a-row quadruplets of coordinates,
// listed systematically.
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

// The 6 cells considered for first moves. All other
// first moves are rotations or reflections of these.
var uniqueCells [][]int = [][]int{
	[]int{0, 0},
	[]int{0, 1},
	[]int{0, 2},
	[]int{1, 1},
	[]int{1, 2},
	[]int{2, 2},
}

// List of all 25 cell coordinates. Used in
// alphaBeta() to try to trigger alpha or beta
// cutoffs by cheaply ordering the moves under
// consideration. Seems to help performance some.
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

// Slight bias for first few moves when 3-of-winning-4
// or wins or losses aren't inside the horizon.
// Probably should be based on a deep run of opening2.go
var scores [][]int = [][]int{
	[]int{3, 3, 0, 3, 3},
	[]int{3, 4, 1, 4, 3},
	[]int{0, 1, 0, 1, 2},
	[]int{3, 4, 1, 4, 3},
	[]int{3, 3, 0, 3, 3},
}

// Transformation matrices. Put a mark's <x,y> coords in
// one of the 8 matrices, get a [2]int back that has
// the transformed (rotated or reflected) coords of that
// mark. Used to find out if some pair of moves has already
// been done in a rotated or reflected way.
var Matrices [8][5][5][2]int = [8][5][5][2]int{
	// 0: I - Identity
	[5][5][2]int{
		[5][2]int{[2]int{0, 0}, [2]int{0, 1}, [2]int{0, 2}, [2]int{0, 3}, [2]int{0, 4}},
		[5][2]int{[2]int{1, 0}, [2]int{1, 1}, [2]int{1, 2}, [2]int{1, 3}, [2]int{1, 4}},
		[5][2]int{[2]int{2, 0}, [2]int{2, 1}, [2]int{2, 2}, [2]int{2, 3}, [2]int{2, 4}},
		[5][2]int{[2]int{3, 0}, [2]int{3, 1}, [2]int{3, 2}, [2]int{3, 3}, [2]int{3, 4}},
		[5][2]int{[2]int{4, 0}, [2]int{4, 1}, [2]int{4, 2}, [2]int{4, 3}, [2]int{4, 4}},
	},

	// 1: A - rotate 90 deg CCW about Z axis
	[5][5][2]int{
		[5][2]int{[2]int{0, 4}, [2]int{1, 4}, [2]int{2, 4}, [2]int{3, 4}, [2]int{4, 4}},
		[5][2]int{[2]int{0, 3}, [2]int{1, 3}, [2]int{2, 3}, [2]int{3, 3}, [2]int{4, 3}},
		[5][2]int{[2]int{0, 2}, [2]int{1, 2}, [2]int{2, 2}, [2]int{3, 2}, [2]int{4, 2}},
		[5][2]int{[2]int{0, 1}, [2]int{1, 1}, [2]int{2, 1}, [2]int{3, 1}, [2]int{4, 1}},
		[5][2]int{[2]int{0, 0}, [2]int{1, 0}, [2]int{2, 0}, [2]int{3, 0}, [2]int{4, 0}},
	},

	// 2: B - rotate 180 deg CCW about Z axis
	[5][5][2]int{
		[5][2]int{[2]int{4, 4}, [2]int{4, 3}, [2]int{4, 2}, [2]int{4, 1}, [2]int{4, 0}},
		[5][2]int{[2]int{3, 4}, [2]int{3, 3}, [2]int{3, 2}, [2]int{3, 1}, [2]int{3, 0}},
		[5][2]int{[2]int{2, 4}, [2]int{2, 3}, [2]int{2, 2}, [2]int{2, 1}, [2]int{2, 0}},
		[5][2]int{[2]int{1, 4}, [2]int{1, 3}, [2]int{1, 2}, [2]int{1, 1}, [2]int{1, 0}},
		[5][2]int{[2]int{0, 4}, [2]int{0, 3}, [2]int{0, 2}, [2]int{0, 1}, [2]int{0, 0}},
	},

	// 3: C - rotate 90 deg CW about Z axis
	[5][5][2]int{
		[5][2]int{[2]int{4, 0}, [2]int{3, 0}, [2]int{2, 0}, [2]int{1, 0}, [2]int{0, 0}},
		[5][2]int{[2]int{4, 1}, [2]int{3, 1}, [2]int{2, 1}, [2]int{1, 1}, [2]int{0, 1}},
		[5][2]int{[2]int{4, 2}, [2]int{3, 2}, [2]int{2, 2}, [2]int{1, 2}, [2]int{0, 2}},
		[5][2]int{[2]int{4, 3}, [2]int{3, 3}, [2]int{2, 3}, [2]int{1, 3}, [2]int{0, 3}},
		[5][2]int{[2]int{4, 4}, [2]int{3, 4}, [2]int{2, 4}, [2]int{1, 4}, [2]int{0, 4}},
	},

	// 4: D - reflect across horizontal axis
	[5][5][2]int{
		[5][2]int{[2]int{4, 0}, [2]int{4, 1}, [2]int{4, 2}, [2]int{4, 3}, [2]int{4, 4}},
		[5][2]int{[2]int{3, 0}, [2]int{3, 1}, [2]int{3, 2}, [2]int{3, 3}, [2]int{3, 4}},
		[5][2]int{[2]int{2, 0}, [2]int{2, 1}, [2]int{2, 2}, [2]int{2, 3}, [2]int{2, 4}},
		[5][2]int{[2]int{1, 0}, [2]int{1, 1}, [2]int{1, 2}, [2]int{1, 3}, [2]int{1, 4}},
		[5][2]int{[2]int{0, 0}, [2]int{0, 1}, [2]int{0, 2}, [2]int{0, 3}, [2]int{0, 4}},
	},

	// 5: E - reflect across vertical axis
	[5][5][2]int{
		[5][2]int{[2]int{0, 4}, [2]int{0, 3}, [2]int{0, 2}, [2]int{0, 1}, [2]int{0, 0}},
		[5][2]int{[2]int{1, 4}, [2]int{1, 3}, [2]int{1, 2}, [2]int{1, 1}, [2]int{1, 0}},
		[5][2]int{[2]int{2, 4}, [2]int{2, 3}, [2]int{2, 2}, [2]int{2, 1}, [2]int{2, 0}},
		[5][2]int{[2]int{3, 4}, [2]int{3, 3}, [2]int{3, 2}, [2]int{3, 1}, [2]int{3, 0}},
		[5][2]int{[2]int{4, 4}, [2]int{4, 3}, [2]int{4, 2}, [2]int{4, 1}, [2]int{4, 0}},
	},

	// 6: F - reflect across upper right to lower left diagnoal
	[5][5][2]int{
		[5][2]int{[2]int{4, 4}, [2]int{3, 4}, [2]int{2, 4}, [2]int{1, 4}, [2]int{0, 4}},
		[5][2]int{[2]int{4, 3}, [2]int{3, 3}, [2]int{2, 3}, [2]int{1, 3}, [2]int{0, 3}},
		[5][2]int{[2]int{4, 2}, [2]int{3, 2}, [2]int{2, 2}, [2]int{1, 2}, [2]int{0, 2}},
		[5][2]int{[2]int{4, 1}, [2]int{3, 1}, [2]int{2, 1}, [2]int{1, 1}, [2]int{0, 1}},
		[5][2]int{[2]int{4, 0}, [2]int{3, 0}, [2]int{2, 0}, [2]int{1, 0}, [2]int{0, 0}},
	},

	// 7: G - reflect across upper left to lower right diagnoal
	[5][5][2]int{
		[5][2]int{[2]int{0, 0}, [2]int{1, 0}, [2]int{2, 0}, [2]int{3, 0}, [2]int{4, 0}},
		[5][2]int{[2]int{0, 1}, [2]int{1, 1}, [2]int{2, 1}, [2]int{3, 1}, [2]int{4, 1}},
		[5][2]int{[2]int{0, 2}, [2]int{1, 2}, [2]int{2, 2}, [2]int{3, 2}, [2]int{4, 2}},
		[5][2]int{[2]int{0, 3}, [2]int{1, 3}, [2]int{2, 3}, [2]int{3, 3}, [2]int{4, 3}},
		[5][2]int{[2]int{0, 4}, [2]int{1, 4}, [2]int{2, 4}, [2]int{3, 4}, [2]int{4, 4}},
	},
}

// These are the only 6 first moves to consider:
// all other first moves are a reflection or rotation
// (see Matrices[][][]{] above) of these 6.
var firstMoves [6][2]int = [6][2]int{
	[2]int{0, 0},
	[2]int{0, 1},
	[2]int{0, 2},
	[2]int{1, 1},
	[2]int{1, 2},
	[2]int{2, 2},
}
