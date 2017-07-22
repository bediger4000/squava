package abbook

import (
	"fmt"
	"math/rand"
	"movekeeper"
	"os"
)

type board [5][5]int

const (
	WIN       = 10000
	LOSS      = -10000
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

type AlphaBetaBook struct {
	bd             *board
	leafNodeCount  int
	maxDepth       int
	deterministic  bool
	moveCount      int
	state          int
	c_x            int
	c_y            int
	bookInProgress bool
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

func New(deterministic bool, maxdepth int) *AlphaBetaBook {
	if !indexedCalcs {
		calculateIndexedMatrices()
		indexedCalcs = true
	}
	var r AlphaBetaBook
	r.bd = new(board)
	r.maxDepth = maxdepth
	r.deterministic = deterministic
	r.state = FIRST
	r.bookInProgress = true
	return &r
}

func (p *AlphaBetaBook) MakeMove(x, y int, player int) {
	p.moveCount++
	p.bd[x][y] = player
}

func (p *AlphaBetaBook) SetDepth(moveCounter int) {
	if moveCounter < 4 {
		p.maxDepth = 6
	}
	if moveCounter > 3 {
		p.maxDepth = 8
	}
	if moveCounter > 10 {
		p.maxDepth = 10
	}
}

// Choose computer's next move: return x,y coords of move and its score.
func (p *AlphaBetaBook) ChooseMove() (xcoord int, ycoord int, value int, leafcount int) {

	if p.bookInProgress {
		p.bookStart()
		return p.c_x, p.c_y, 0, 0
	}

	moves := movekeeper.New(2*LOSS, p.deterministic)

	p.leafNodeCount = 0

	for i, row := range p.bd {
		for j, mark := range row {
			if mark == UNSET {
				p.bd[i][j] = MAXIMIZER
				stop, value := p.deltaValue(0, i, j)
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

func (p *AlphaBetaBook) findWinner() int {
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
func (p *AlphaBetaBook) deltaValue(ply int, x, y int) (stopRecursing bool, value int) {

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
	}

	return stopRecursing, value
}

func (p *AlphaBetaBook) alphaBeta(ply int, player int, alpha int, beta int, x int, y int, boardValue int) (value int) {

	switch player {
	case MAXIMIZER:
		value = 2 * LOSS
		for i, row := range p.bd {
			for j, marker := range row {
				if marker == UNSET {
					p.bd[i][j] = MAXIMIZER
					stopRecursing, delta := p.deltaValue(ply, x, y)
					if stopRecursing {
						p.bd[i][j] = UNSET
						p.leafNodeCount++
						return delta
					}
					n := p.alphaBeta(ply+1, MINIMIZER, alpha, beta,
						i, j, boardValue+delta)
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
		value = 2 * WIN
		for i, row := range p.bd {
			for j, marker := range row {
				if marker == UNSET {
					p.bd[i][j] = player
					stopRecursing, delta := p.deltaValue(ply, x, y)
					if stopRecursing {
						p.bd[i][j] = UNSET
						p.leafNodeCount++
						return delta
					}
					n := p.alphaBeta(ply+1, -player, alpha, beta,
						i, j, boardValue+delta)
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

func (p *AlphaBetaBook) PrintBoard() {
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

var scores [5][5]int

func (p *AlphaBetaBook) SetScores(randomize bool) {
	if randomize {
		var vals [11]int = [11]int{-5, -4, -3 - 2, -1, 0, 1, 2, 3, 4, 5}
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

func (p *AlphaBetaBook) FindWinner() int {
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

// Implement an opening "book". Make the first move in
// one of the 4 upper-right-hand cells. Next try to make
// the diagonal end of 4-in-a-row off that first move.
// If that cell is taken, try to get the "corner", so
// as to be able to make a diagonal off the corner,
// while getting the ends of a 4-in-a-row with the
// original move. Then, complete the "triangle of
// doom" with the third move. A configuration where
// the human takes the end of the initial diagonal
// and the opposite corner causes it to bail out of
// the book early.

const (
	FIRST = iota
	DIAGONAL
	CORNER
	OTHERCORNER
	OTHERDIAGONAL
	LAST
)

var firstMoves [4][2]int = [4][2]int{
	{0, 0},
	{0, 1},
	{1, 0},
	{1, 1},
}

func (p *AlphaBetaBook) bookStart() {

	looping := true

	for looping {
		looping = false
		switch p.state {
		case FIRST:
			c := firstMoves[rand.Intn(4)]
			p.c_x, p.c_y = c[0], c[1]
			p.bd[p.c_x][p.c_y] = MAXIMIZER
			p.state = DIAGONAL
		case DIAGONAL:
			p_x, p_y := p.c_x+3, p.c_y+3
			if p.bd[p_x][p_y] == UNSET {
				p.c_x += 3
				p.c_y += 3
				p.bd[p.c_x][p.c_y] = MAXIMIZER
				p.moveCount++
				p.state = CORNER
			} else {
				p.state = OTHERCORNER
				looping = true
			}
		case CORNER:
			p_x, p_y := p.c_x-3, p.c_y
			if p.bd[p_x][p_y] == UNSET {
				p.c_x -= 3
				p.bd[p.c_x][p.c_y] = MAXIMIZER
				p.moveCount++
			} else {
				p_x, p_y = p.c_x, p.c_y-3
				if p.bd[p_x][p_y] == UNSET {
					p.c_y -= 3
					p.bd[p.c_x][p.c_y] = MAXIMIZER
					p.moveCount++
				}
			}
			p.state = LAST
			p.bookInProgress = false
		case OTHERCORNER:
			// Didn't get desired diagonal
			p_x, p_y := p.c_x, p.c_y+3
			if p.bd[p_x][p_y] == UNSET {
				p.c_y += 3
				p.bd[p.c_x][p.c_y] = MAXIMIZER
				p.moveCount++
				p.state = OTHERDIAGONAL
			} else {
				fmt.Printf("Unreachable state in OTHERCORNER\n")
				p.PrintBoard()
				os.Exit(99)
			}
		case OTHERDIAGONAL:
			p_x, p_y := p.c_x+3, p.c_y-3
			if p.bd[p_x][p_y] == UNSET {
				p.c_x += 3
				p.c_y -= 3
				p.bd[p.c_x][p.c_y] = MAXIMIZER
				p.moveCount++
			}
			p.state = LAST
			p.bookInProgress = false
		}
	}
}
