package abgeo

/* abgeo - alpha/beta minimax, with some scoring for avoiding 2-in-a-rows
 * that can't possibly amount to a win ever.
 */

import (
	"fmt"
	"math/rand"

	"squava/src/movekeeper"
)

type board [5][5]int

const (
	WIN       = 10000
	LOSS      = -10000
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

type AlphaBetaGeo struct {
	bd            *board
	leafNodeCount int
	maxDepth      int
	deterministic bool
}

// Arrays of losing triplets and winning quads, indexed
// by <x,y> coords of all pairs composing each of the quads
// or triplets. Makes deltaValue() a lot more efficient
var indexedLosingTriplets [5][5][][][]int
var indexedWinningQuads [5][5][][][]int
var indexedCalcs bool = false

func calculateIndexedMatrices() {
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
}

func New(deterministic bool, maxdepth int) *AlphaBetaGeo {
	if !indexedCalcs {
		calculateIndexedMatrices()
		indexedCalcs = true
	}
	return &AlphaBetaGeo{
		bd:            new(board),
		maxDepth:      maxdepth,
		deterministic: deterministic,
	}
}

func (p *AlphaBetaGeo) Name() string {
	return "A/B+Avoid"
}

func (p *AlphaBetaGeo) MakeMove(x, y int, player int) {
	p.bd[x][y] = player
}

func (p *AlphaBetaGeo) SetDepth(moveCounter int) {
	if moveCounter < 4 {
		p.maxDepth = 8
	}
	if moveCounter > 3 {
		p.maxDepth = 10
	}
	if moveCounter > 10 {
		p.maxDepth = 12
	}
}

// Choose computer's next move: return x,y coords of move and its score.
func (p *AlphaBetaGeo) ChooseMove() (xcoord int, ycoord int, value int, leafcount int) {

	moves := movekeeper.New(2*LOSS, p.deterministic)

	p.leafNodeCount = 0

	for i, row := range p.bd {
		for j, mark := range row {
			if mark == UNSET {
				p.bd[i][j] = MAXIMIZER
				stop, value := p.deltaValue(0, i, j, 0)
				if !stop {
					value = p.alphaBeta(1, MINIMIZER, 2*LOSS, 2*WIN, i, j, value)
				}
				p.bd[i][j] = UNSET
				moves.SetMove(i, j, value)
			}
		}
	}

	a, b, v := moves.ChooseMove()

	p.MakeMove(a, b, MAXIMIZER)

	return a, b, v, p.leafNodeCount
}

func (p *AlphaBetaGeo) findWinner() int {
	for _, quad := range winningQuads {
		sum := p.bd[quad[0][0]][quad[0][1]]
		sum += p.bd[quad[1][0]][quad[1][1]]
		sum += p.bd[quad[2][0]][quad[2][1]]
		sum += p.bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return p.bd[quad[0][0]][quad[0][1]]
		}
	}

	for _, triplet := range losingTriplets {
		sum := p.bd[triplet[0][0]][triplet[0][1]]
		sum += p.bd[triplet[1][0]][triplet[1][1]]
		sum += p.bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			return -p.bd[triplet[0][0]][triplet[0][1]]
		}
	}

	return 0
}

// Calculates and returns the value of the move (x,y)
// Only considers value gained or lost from the cell (x,y)
func (p *AlphaBetaGeo) deltaValue(ply int, x, y int, currentValue int) (stopRecursing bool, value int) {

	relevantQuads := indexedWinningQuads[x][y]
	for _, quad := range relevantQuads {
		sum := p.bd[quad[0][0]][quad[0][1]]
		sum += p.bd[quad[1][0]][quad[1][1]]
		sum += p.bd[quad[2][0]][quad[2][1]]
		sum += p.bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return true, p.bd[quad[0][0]][quad[0][1]] * (WIN - ply)
		}
		if sum == 3 || sum == -3 {
			value += sum * 10
		}
	}

	relevantTriplets := indexedLosingTriplets[x][y]
	for _, triplet := range relevantTriplets {
		sum := p.bd[triplet[0][0]][triplet[0][1]]
		sum += p.bd[triplet[1][0]][triplet[1][1]]
		sum += p.bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			return true, sum / 3 * (LOSS + ply)
		}
	}

	for _, triplet := range no2 {
		for _, pair := range triplet {
			if x == pair[0] && y == pair[1] {
				sum := p.bd[triplet[0][0]][triplet[0][1]]
				sum += p.bd[triplet[1][0]][triplet[1][1]]
				sum += p.bd[triplet[2][0]][triplet[2][1]]
				if sum == 2 || sum == -2 {
					value += p.bd[x][y] * -100
				}
				break
			}
		}
	}

	for _, quad := range noMiddle2 {
		player := p.bd[x][y]
		if (x == quad[1][0] && y == quad[1][1] && player == p.bd[quad[2][0]][quad[2][1]]) ||
			(x == quad[2][0] && y == quad[2][1] && player == p.bd[quad[1][0]][quad[1][1]]) {

			sum := p.bd[quad[0][0]][quad[0][1]]
			sum += p.bd[quad[1][0]][quad[1][1]]
			sum += p.bd[quad[2][0]][quad[2][1]]
			sum += p.bd[quad[3][0]][quad[3][1]]

			if sum == 2 || sum == -2 {
				value += player * -100
			}
		}
	}

	// Give it a slight bias for those early
	// moves when all losing-triplets and winning-quads
	// are beyond the horizon.
	value += p.bd[x][y] * scores[x][y]

	// If squava has a "cat game", then this is wrong. Cat
	// games could stop recursing here.
	stopRecursing = false
	if ply >= p.maxDepth {
		stopRecursing = true
		value += currentValue
	}

	return stopRecursing, value
}

func (p *AlphaBetaGeo) alphaBeta(ply int, player int, alpha int, beta int, x int, y int, boardValue int) (value int) {

	switch player {
	case MAXIMIZER:
		value = 2 * LOSS // Possible to score less than LOSS
		for i, row := range p.bd {
			for j, marker := range row {
				if marker == UNSET {
					p.bd[i][j] = MAXIMIZER
					stopRecursing, delta := p.deltaValue(ply, x, y, boardValue)
					if stopRecursing {
						p.bd[i][j] = UNSET
						p.leafNodeCount++
						return delta
					}
					n := p.alphaBeta(ply+1, MINIMIZER, alpha, beta, i, j, delta)
					p.bd[i][j] = UNSET
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
		for i, row := range p.bd {
			for j, marker := range row {
				if marker == UNSET {
					p.bd[i][j] = player
					stopRecursing, delta := p.deltaValue(ply, x, y, boardValue)
					if stopRecursing {
						p.bd[i][j] = UNSET
						p.leafNodeCount++
						return delta
					}
					n := p.alphaBeta(ply+1, -player, alpha, beta, i, j, delta)
					p.bd[i][j] = UNSET
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

func (p *AlphaBetaGeo) PrintBoard() {
	fmt.Printf("   0 1 2 3 4\n")
	for i, row := range p.bd {
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
	fmt.Printf("\n")
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

func (p *AlphaBetaGeo) SetScores(randomize bool) {
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

func (p *AlphaBetaGeo) FindWinner() int {
	for _, quad := range winningQuads {
		sum := p.bd[quad[0][0]][quad[0][1]]
		sum += p.bd[quad[1][0]][quad[1][1]]
		sum += p.bd[quad[2][0]][quad[2][1]]
		sum += p.bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return p.bd[quad[0][0]][quad[0][1]]
		}
	}

	for _, triplet := range losingTriplets {
		sum := p.bd[triplet[0][0]][triplet[0][1]]
		sum += p.bd[triplet[1][0]][triplet[1][1]]
		sum += p.bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			return -p.bd[triplet[0][0]][triplet[0][1]]
		}
	}

	return 0 // Cat got the game
}

var a [12][5][2]int = [12][5][2]int{
	{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}},
	{{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}},
	{{2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}},
	{{3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}},
	{{4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}},

	{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}},
	{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {4, 1}},
	{{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2}},
	{{0, 3}, {1, 3}, {2, 3}, {3, 3}, {4, 3}},
	{{0, 4}, {1, 4}, {2, 4}, {3, 4}, {4, 4}},

	{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}},
	{{0, 4}, {1, 3}, {2, 2}, {3, 1}, {4, 0}},
}

// 4-in-a-row where you don't want to have the middle 2
var noMiddle2 [4][4][2]int = [4][4][2]int{
	{{3, 0}, {2, 1}, {1, 2}, {0, 3}},
	{{1, 0}, {2, 1}, {3, 2}, {4, 3}},
	{{0, 1}, {1, 2}, {2, 3}, {3, 4}},
	{{1, 4}, {2, 3}, {3, 2}, {4, 1}},
}

// 3-in-a-row where you don't want any 2 plus a blank
var no2 [4][3][2]int = [4][3][2]int{

	{{2, 0}, {1, 1}, {0, 2}},
	{{0, 2}, {1, 3}, {2, 4}},
	{{4, 2}, {3, 3}, {2, 4}},
	{{4, 2}, {3, 1}, {2, 0}},
}
