package negascout


import (
	"fmt"
	"math/rand"
	"movekeeper"
)

type board [5][5]int

const (
	WIN       = 10000
	LOSS      = -10000
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)
type NegaScout struct {
    bd *board
    leafNodeCount int
    maxDepth int
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

func New(deterministic bool, maxdepth int) (*NegaScout) {
    if !indexedCalcs {
        calculateIndexedMatrices()
        indexedCalcs = true
    }
	var r NegaScout
    r.bd = new(board)
    r.maxDepth = maxdepth
    r.deterministic = deterministic
    return &r
}

func (p *NegaScout) MakeMove(x, y int, player int) {
    p.bd[x][y] = player
}

func (p *NegaScout) SetDepth(moveCounter int) {
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

func (p *NegaScout) ChooseMove() (int, int, int, int) {

	moves := movekeeper.New(3*LOSS, p.deterministic)
	p.leafNodeCount = 0

	p.reorderMoves()

	beta := 2*WIN
	alpha := 2*LOSS
	score := 3*LOSS
	n := beta

	for _, cell := range orderedMoves[2] {
		i, j := cell[0], cell[1]
		if p.bd[i][j] == UNSET {
			p.bd[i][j] = MAXIMIZER
			cur := -p.negaScout(1, MINIMIZER, -n, -alpha)
			if cur > score {
				if n == beta {
					score = cur
				} else {
					score = -p.negaScout(1, MINIMIZER, -beta, -cur)
				}
			}
			p.bd[i][j] = UNSET
			if score > alpha { alpha = score }
			moves.SetMove(i, j, alpha)
			if alpha >= beta { break }
			n = alpha + 1
		}
	}


	a, b, v := moves.ChooseMove()

	p.MakeMove(a, b, MAXIMIZER)

	return a, b, v, p.leafNodeCount
}

func (p *NegaScout) FindWinner() int {
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

var deadlyQuads [4][4][2]int = [4][4][2]int{
	{{1,0}, {2,1}, {3,2}, {4,3}},
	{{4,1}, {3,2}, {2,3}, {1,4}},
	{{0,1}, {1,2}, {2,3}, {3,4}},
	{{3,0}, {2,1}, {1,2}, {0,3}},
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

func (p *NegaScout) staticValue(ply int) (stopRecursing bool, value int) {

	p.leafNodeCount++

	for _, cell := range checkableCells {
		relevantQuads := indexedWinningQuads[cell[0]][cell[1]]
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
	}

	for _, cell := range checkableCells {
		relevantTriplets := indexedLosingTriplets[cell[0]][cell[1]]
		for _, triplet := range relevantTriplets {
			sum := p.bd[triplet[0][0]][triplet[0][1]]
			sum += p.bd[triplet[1][0]][triplet[1][1]]
			sum += p.bd[triplet[2][0]][triplet[2][1]]

			if sum == 3 || sum == -3 {
				return true, -sum / 3 * (WIN - ply)
			}
		}
	}

	for _, quad := range deadlyQuads {
		outer := p.bd[quad[0][0]][quad[0][1]] + p.bd[quad[3][0]][quad[3][1]]
		inner := p.bd[quad[1][0]][quad[1][1]] + p.bd[quad[2][0]][quad[2][1]]

		if (inner == 2 || inner == -2) && outer == 0 {
			value -= inner/2 * 5
		}
	}

	if value == 0 {
		// Give it a slight bias for those early
		// moves when all losing-triplets and winning-quads
		// are beyond the horizon.
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				value += p.bd[x][y] * scores[x][y]
			}
		}
	}

	// If squava has a "cat game", then this is wrong. Cat
	// games could stop recursing here.
	stopRecursing = false
	if ply > p.maxDepth {
		stopRecursing = true
	}

	return stopRecursing, value
}

func (p *NegaScout) negaScout(ply int, player int, alpha int, beta int) (value int) {

	stopRecursing, boardValue := p.staticValue(ply)
	if stopRecursing {
		return player*boardValue
	}

	score := 3*LOSS  // Even 2*LOSS greater than this
	n := beta

	for _, cell := range orderedMoves[player+1] {
		i, j := cell[0], cell[1]
		if p.bd[i][j] == UNSET {
			p.bd[i][j] = player
			cur := -p.negaScout(ply+1, -player, -n, -alpha)
			if cur > score {
				if n == beta || ply == p.maxDepth - 2 {
					score = cur
				} else {
					score = -p.negaScout(ply+1, -player, -beta, -cur)
				}
			}
			p.bd[i][j] = UNSET
			if score > alpha { alpha = score }
			if alpha >= beta { break }
			n = alpha + 1
		}
	}

	return score
}

func (p *NegaScout) PrintBoard() {
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

func (p *NegaScout) SetScores(randomize bool) {
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

// Need a list of all possible moves, in an order that
// works for the first 2 or 3 moves by instances of NegaScout.
var initialOrderedMoves [25][2]int = [25][2]int{
    {1,1}, {1,3}, {3,3}, {3,1},
    {0,1}, {0,3}, {1,4}, {3,4}, {4,3}, {4,1}, {3,0}, {1,0},
    {0,0}, {0,4}, {4,4}, {4,0},
    {2,2},
    {1,2}, {2,3}, {3,2}, {2,1},
    {0,2}, {2,0}, {2,4}, {4,2},
}

// Used in func negaScout(), orderedMoves[0] by MINIMIZER,
// orderedMoves[2] by MAXIMIZER, orderedMoves[1] unused.
var orderedMoves [3][][2]int

// Reorder all legal moves into something like a "best order"
// This considers moves as good, bad or dull. Good or bad depends
// on who's playing, but both good and bad moves are holes in winning quads,
// and holes in losing triplets. Other moves are dull.
// MINIMIZER uses an array of bad-dull-good moves, MAXIMIZER uses an
// array of good-dull-bad moves, both in ChooseMove() and negaScout().
func (p *NegaScout) reorderMoves() {
	var goodCells [][2]int
	var badCells [][2]int
	var dullCells [][2]int
	unsetCount := 0
	for _, cell := range initialOrderedMoves {
		if p.bd[cell[0]][cell[1]] == UNSET {
			unsetCount++
			interesting := false
			relevantQuads := indexedWinningQuads[cell[0]][cell[1]]
			for _, quad := range relevantQuads {
				sum := p.bd[quad[0][0]][quad[0][1]]
				sum += p.bd[quad[1][0]][quad[1][1]]
				sum += p.bd[quad[2][0]][quad[2][1]]
				sum += p.bd[quad[3][0]][quad[3][1]]

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
					sum := p.bd[triplet[0][0]][triplet[0][1]]
					sum += p.bd[triplet[1][0]][triplet[1][1]]
					sum += p.bd[triplet[2][0]][triplet[2][1]]

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

	l := len(goodCells)
	for _, cell := range dullCells {
		goodCells = append(goodCells, cell)
	}
	for _, cell := range badCells {
		goodCells = append(goodCells, cell)
	}

	for _, cell := range dullCells {
		badCells = append(badCells, cell)
	}
	for _, cell := range goodCells[0:l] {
		badCells = append(badCells, cell)
	}

	if len(goodCells) != unsetCount {
		fmt.Printf("Problem, found %d maximizer-ordered moves, should find %d\n", len(goodCells), unsetCount)
	}
	if len(badCells) != unsetCount {
		fmt.Printf("Problem, found %d minimizer-ordered moves, should find %d\n", len(badCells), unsetCount)
	}

	orderedMoves[0] = badCells
	orderedMoves[2] = goodCells
}
