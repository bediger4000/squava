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

var leafNodeCount int
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
	useBook := flag.Bool("B", false, "Use book start")
	flag.Parse()

	*printBoardPtr = !*printBoardPtr

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

	rand.Seed(time.Now().UTC().UnixNano())

	setScores(*randomizeScores)

	var humanFirst bool = *humanFirstPtr
	if *computerFirstPtr {
		humanFirst = false
	}

	moveCounter := 0
	var bd Board

	if *useBook {
		fmt.Printf("Using opening book\n")
		*firstMovePtr = ""
		if humanFirst {
			l, m := readMove(&bd, *printBoardPtr)
			bd[l][m] = MINIMIZER
			moveCounter = 1 + bookDefend(&bd, l, m)
		} else {
			moveCounter += bookStart(&bd)
		}
	}

	if *firstMovePtr != "" {
		var x1, y1 int
		fmt.Sscanf(*firstMovePtr, "%d,%d", &x1, &y1)
		fmt.Printf("My move: %d %d\n", x1, y1)
		humanFirst = true
		bd[x1][y1] = MAXIMIZER
		printBoard(&bd)
	}

	var endOfGame bool = false

	for !endOfGame {
		setDepth(moveCounter, *maxDepthPtr)

		var l, m int
		if humanFirst {
			l, m = readMove(&bd, *printBoardPtr)
			bd[l][m] = MINIMIZER
			endOfGame, _ = deltaValue(&bd, 0, l, m)
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

		endOfGame, _ = deltaValue(&bd, 0, a, b)

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
		fmt.Print(phrase)

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
func chooseMove(bd *Board, deterministic bool) (xcoord int, ycoord int, value int) {

	var moves = MoveKeeper{next: 0, max: 2 * LOSS}

	for i, row := range bd {
		for j, mark := range row {
			if mark == UNSET {
				bd[i][j] = MAXIMIZER
				stopRecursing, value := deltaValue(bd, 0, i, j)
				if !stopRecursing {
					value = alphaBeta(bd, 1, MINIMIZER, 2*LOSS, 2*WIN, i, j, value)
				}
				bd[i][j] = UNSET
				moves.setMove(i, j, value)
			}
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
	value += bd[x][y] * scores[x][y]

	// If squava has a "cat game", then this is wrong. Cat
	// games could stop recursing here.
	stopRecursing = false
	if ply == maxDepth {
		stopRecursing = true
	}

	return stopRecursing, value
}

func alphaBeta(bd *Board, ply int, player int, alpha int, beta int, x int, y int, boardValue int) (value int) {

	switch player {
	case MAXIMIZER:
		value = 2 * LOSS // Possible to score less than LOSS
		for i, row := range bd {
			for j, marker := range row {
				if marker == UNSET {
					bd[i][j] = MAXIMIZER
					stopRecursing, delta := deltaValue(bd, ply, x, y)
					if stopRecursing {
						bd[i][j] = UNSET
						leafNodeCount++
						return delta
					}
					n := alphaBeta(bd, ply+1, MINIMIZER, alpha, beta, i, j, boardValue+delta)
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
					stopRecursing, delta := deltaValue(bd, ply, x, y)
					if stopRecursing {
						bd[i][j] = UNSET
						leafNodeCount++
						return delta
					}
					n := alphaBeta(bd, ply+1, -player, alpha, beta, i, j, boardValue+delta)
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

func bookDefend(bd *Board, firstX int, firstY int) int {
	state := FIRST
	moveCount := 0

	var c_x, c_y int

	for state != LAST {
		switch state {
		case FIRST:
			// Find the diagonal, and block it
		OUTERFIRST:
			for i := -3; i < 4; i += 6 {
				for j := -3; j < 4; j += 6 {
					a := firstX + i
					b := firstY + j
					if a >= 0 && a <= 4 && b >= 0 && b <= 4 {
						// Since <firstX, firstY> have an X, <a,b> must be empty
						c_x = a
						c_y = b
						bd[c_x][c_y] = MAXIMIZER
						moveCount++
						break OUTERFIRST
					}
				}
			}
			state = DIAGONAL

			fmt.Printf("My move: %d %d\n", c_x, c_y)
			printBoard(bd)
			l, m := readMove(bd, true)
			bd[l][m] = MINIMIZER
			moveCount++

		case DIAGONAL:
			state = LAST
			var lastx, lasty int
		OUTERDIAGONAL:
			for i, row := range bd {
				for j, mark := range row {
					if !(i == firstX && j == firstY) && mark == MINIMIZER {
						lastx = i
						lasty = j
						break OUTERDIAGONAL
					}
				}
			}
			// lastx, lasty - coords of move to respond to
		FOUNDMOVE:
			for i := -3; i < 4; i += 6 {
				for j := -3; j < 4; j += 6 {
					a, b := lastx+i, lasty+j
					if a >= 0 && a <= 4 && b >= 0 && b <= 4 {
						if bd[a][b] == UNSET {
							bd[a][b] = MAXIMIZER
							moveCount++
							c_x, c_y = a, b
							break FOUNDMOVE
						}
					}
				}
			}
			fmt.Printf("My move: %d %d\n", c_x, c_y)
			printBoard(bd)
		}
	}

	return moveCount
}

func bookStart(bd *Board) int {

	state := FIRST
	moveCount := 0

	var c_x, c_y int

	for state != LAST {
		switch state {
		case FIRST:
			c := firstMoves[rand.Intn(4)]
			c_x, c_y = c[0], c[1]
			bd[c_x][c_y] = MAXIMIZER
			state = DIAGONAL
			moveCount++
		case DIAGONAL:
			p_x, p_y := c_x+3, c_y+3
			if bd[p_x][p_y] == UNSET {
				c_x += 3
				c_y += 3
				bd[c_x][c_y] = MAXIMIZER
				moveCount++
				state = CORNER
			} else {
				state = OTHERCORNER
			}
		case CORNER:
			p_x, p_y := c_x-3, c_y
			if bd[p_x][p_y] == UNSET {
				c_x -= 3
				bd[c_x][c_y] = MAXIMIZER
				moveCount++
			} else {
				p_x, p_y = c_x, c_y-3
				if bd[p_x][p_y] == UNSET {
					c_y -= 3
					bd[c_x][c_y] = MAXIMIZER
					moveCount++
				}
			}
			state = LAST
		case OTHERCORNER:
			// Didn't get desired diagonal
			p_x, p_y := c_x, c_y+3
			if bd[p_x][p_y] == UNSET {
				c_y += 3
				bd[c_x][c_y] = MAXIMIZER
				moveCount++
				state = OTHERDIAGONAL
			} else {
				fmt.Printf("Unreachable state in OTHERCORNER\n")
				printBoard(bd)
				os.Exit(99)
			}
		case OTHERDIAGONAL:
			p_x, p_y := c_x+3, c_y-3
			if bd[p_x][p_y] == UNSET {
				c_x += 3
				c_y -= 3
				bd[c_x][c_y] = MAXIMIZER
				moveCount++
			}
			state = LAST
		}
		if (moveCount % 2) == 1 {
			fmt.Printf("My move: %d %d\n", c_x, c_y)
			printBoard(bd)
			l, m := readMove(bd, true)
			bd[l][m] = MINIMIZER
			moveCount++
		}
	}

	return moveCount
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
