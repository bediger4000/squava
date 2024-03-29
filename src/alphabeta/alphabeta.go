package alphabeta

import (
	"fmt"
	"math/rand"

	"squava/src/movekeeper"
)

type board [5][5]int

// Semantically meaningful constant names
const (
	WIN       = 10000
	LOSS      = -10000
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

type AlphaBeta struct {
	bd            *board
	name          string
	leafNodeCount int
	maxDepth      int
	deterministic bool
	boardValue    func(*AlphaBeta, int, int, int, int) (bool, int)
}

// Arrays of losing triplets and winning quads, indexed
// by <x,y> coords of all pairs composing each of the quads
// or triplets. Makes deltaValue() a lot more efficient
var indexedLosingTriplets [5][5][][][]int
var indexedWinningQuads [5][5][][][]int
var indexedCalcs = false

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

func New(deterministic bool, maxdepth int) *AlphaBeta {
	if !indexedCalcs {
		calculateIndexedMatrices()
		indexedCalcs = true
	}
	return &AlphaBeta{
		bd:            new(board),
		name:          "AlphaBeta",
		maxDepth:      maxdepth,
		deterministic: deterministic,
		boardValue:    deltaValue,
	}
}

// Name of the player
func (p *AlphaBeta) Name() string {
	return p.name
}

// MakeMove changes internal board representation,
// making opposing player's move
func (p *AlphaBeta) MakeMove(x, y int, player int) {
	p.bd[x][y] = player
}

// SetDepth changes the max recursion depth based
// on how far along the game has gotten.
func (p *AlphaBeta) SetDepth(moveCounter int) {
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

// ChooseMove - choose computer's next move: return x,y coords of move and its score.
func (p *AlphaBeta) ChooseMove() (xcoord int, ycoord int, value int, leafcount int) {

	moves := movekeeper.New(2*LOSS, p.deterministic)

	p.leafNodeCount = 0

	for i, row := range p.bd {
		for j, mark := range row {
			if mark == UNSET {
				p.bd[i][j] = MAXIMIZER
				stop, value := p.boardValue(p, 0, i, j, 0)
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

// deltaValue calculates the value of the board,
// including value change from move (x,y).
func deltaValue(p *AlphaBeta, ply int, x, y int, currentValue int) (stopRecursing bool, value int) {

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

func (p *AlphaBeta) alphaBeta(ply int, player int, alpha int, beta int, x int, y int, boardValue int) (value int) {

	switch player {
	case MAXIMIZER:
		value = 2 * LOSS // Possible to score less than LOSS
		for i, row := range p.bd {
			for j, marker := range row {
				if marker == UNSET {
					p.bd[i][j] = MAXIMIZER
					stopRecursing, delta := p.boardValue(p, ply, x, y, boardValue)
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
					stopRecursing, delta := p.boardValue(p, ply, x, y, boardValue)
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

// PrintBoard prints the board in a human-readable fashion.
// Necessary to encapsulate the internal representation of
// a 5x5 board
func (p *AlphaBeta) PrintBoard() {
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

var losingTriplets = [][][]int{
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
var winningQuads = [][][]int{
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

var scores [5][5]int

// SetScores does any prep on a new board, like
// initializing a small bias on each cell
func (p *AlphaBeta) SetScores(randomize bool) {
	if randomize {
		var vals = [11]int{-5, -4, -3 - 2, -1, 0, 1, 2, 3, 4, 5}
		for i, row := range scores {
			for j := range row {
				scores[i][j] = vals[rand.Intn(11)]
			}
		}
	} else {
		scores = [5][5]int{
			{3, 3, 0, 3, 3},
			{3, 4, 1, 4, 3},
			{0, 1, 0, 1, 0},
			{3, 4, 1, 4, 3},
			{3, 3, 0, 3, 3},
		}
	}
}

// FindWinner returns the winner of the current game,
// if any, based on internal board representation
func (p *AlphaBeta) FindWinner() int {
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

// Calculates and returns the value of the move (x,y)
// Only considers value gained or lost from the cell (x,y)
func deltaValue2(p *AlphaBeta, ply int, x, y int, currentValue int) (stopRecursing bool, value int) {

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

func (p *AlphaBeta) SetAvoid() {
	p.name = "A/B+Avoid"
	p.boardValue = deltaValue2
}
