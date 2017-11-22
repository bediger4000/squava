package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

// Board is internal representation of a 5x5 tictactoe style
// game board.
type Board [5][5]int

type gameState struct {
	bd        Board
	maxDepth  int
	x, y      int
	value     int
	leafNodes int
	next      *gameState
}

var toDo chan *gameState
var finished chan *gameState

// Manifest constants used repeatedly.
const (
	WIN       = 10000
	LOSS      = -10000
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

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
	threadCountPtr := flag.Int("N", 4, "Use this many threads")
	randomizeScores := flag.Bool("r", false, "Randomize bias scores")
	useBook := flag.Bool("B", false, "Use book start or defense")
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

	toDo = make(chan *gameState, 25)
	finished = make(chan *gameState, 25)

	for i := 0; i < *threadCountPtr; i++ {
		go worker(i, toDo, finished)
	}

	setScores(*randomizeScores)

	humanFirst := *humanFirstPtr
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
			if moveCounter%2 == 1 {
				humanFirst = false
			}
		} else {
			moveCounter += bookStart(&bd)
		}
	}

	if *firstMovePtr != "" {
		var x1, y1 int
		n, e := fmt.Sscanf(*firstMovePtr, "%d,%d", &x1, &y1)
		if n != 2 || e != nil {
			fmt.Printf("Wrong move input. Format N,M, where N and M numbers 0 - 4 \n")
		}
		fmt.Printf("My move: %d %d\n", x1, y1)
		humanFirst = true
		bd[x1][y1] = MAXIMIZER
		printBoard(&bd)
	}

	endOfGame := false

	for !endOfGame {

		maxDepth := setDepth(moveCounter, *maxDepthPtr)

		var l, m int
		if humanFirst {
			l, m = readMove(&bd, *printBoardPtr)
			bd[l][m] = MINIMIZER
			endOfGame, _ = deltaValue(100, &bd, 0, l, m, 0)
			moveCounter++
		}

		if endOfGame {
			break
		}

		humanFirst = true

		start := time.Now()
		a, b, score, leaves := chooseMove(&bd, *deterministic, maxDepth)
		end := time.Now()
		elapsed := end.Sub(start)

		if a < 0 {
			break // Cat gets the game
		}

		bd[a][b] = MAXIMIZER
		moveCounter++

		if *printBoardPtr {
			fmt.Printf("My move: %d %d (%d) [%d] %v\n", a, b, score, leaves, elapsed)
			printBoard(&bd)
		} else {
			fmt.Printf("%d %d\n", a, b)
		}

		endOfGame, _ = deltaValue(100, &bd, 0, a, b, 0)

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

	if toDo != nil {
		close(toDo)
	}
}

func setDepth(moveCounter int, endGameDepth int) int {

	var maxDepth int

	if moveCounter < 4 {
		maxDepth = 6
	}
	if moveCounter > 3 {
		maxDepth = 8
	}
	if moveCounter > 10 {
		maxDepth = 10
	}

	return maxDepth
}

var stateStack *gameState

// oldState puts *gameState points on a stack, for re-use
func oldState(gs *gameState) {
	gs.next = stateStack
	stateStack = gs
}

// newState either allocates a new *gameState, or pulls one off
// the stack of unused *gameStates.
func newState(bd *Board, maxDepth int, value int, x int, y int) *gameState {
	var s *gameState
	if stateStack == nil {
		s = new(gameState)
	} else {
		s = stateStack
		stateStack = stateStack.next
	}

	s.maxDepth = maxDepth
	s.value = value

	for i, row := range bd {
		for j, mark := range row {
			s.bd[i][j] = mark
		}
	}

	s.x = x
	s.y = y
	s.bd[x][y] = MAXIMIZER

	return s
}

// Choose computer's next move: return x,y coords of move and its score.
func chooseMove(bd *Board, deterministic bool, maxDepth int) (xcoord int, ycoord int, value int, leafNodes int) {

	var moves = moveKeeper{max: 2 * LOSS}
	maxDepth--

	for i, row := range bd {
		for j, mark := range row {
			if mark == UNSET {
				stopRecursing, value := deltaValue(maxDepth, bd, 0, i, j, 0)
				if stopRecursing {
					moves.setMove(i, j, value)
					leafNodes++
				} else {
					gs := newState(bd, maxDepth, value, i, j)
					toDo <- gs
				}
			}
		}
	}

	for _, row := range bd {
		for _, mark := range row {
			if mark == UNSET {
				gs := <-finished
				// fmt.Printf("	<%d,%d> %d  [%d]\n", gs.x, gs.y, gs.value, gs.leafNodes)
				moves.setMove(gs.x, gs.y, gs.value)
				leafNodes += gs.leafNodes
				oldState(gs)
			}
		}
	}

	xcoord, ycoord, value = moves.chooseMove(deterministic)
	return xcoord, ycoord, value, leafNodes
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

// Calculates and returns the value of the move (x,y)
// Only considers value gained or lost from the cell (x,y)
func deltaValue(maxDepth int, bd *Board, ply int, x, y int, currentValue int) (stopRecursing bool, value int) {

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
	if ply == maxDepth {
		stopRecursing = true
		value += currentValue
	}

	return stopRecursing, value
}

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

func worker(sn int, from chan *gameState, to chan *gameState) {
	for {
		if curr, ok := <-from; ok {
			curr.value, curr.leafNodes = alphaBeta(
				curr.maxDepth,
				&(curr.bd),
				1,
				MINIMIZER,
				2*LOSS,
				2*WIN,
				curr.x,
				curr.y,
				curr.value)
			to <- curr
		} else {
			return
		}
	}
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

func readMove(bd *Board, print bool) (x, y int) {
	readMove := false
	for !readMove {
		if print {
			fmt.Printf("Your move: ")
		}
		x, y = scanMove()
		readMove = checkMove(bd, x, y, print)
	}
	return x, y
}

func checkMove(bd *Board, x, y int, print bool) bool {
	r := false
	switch {
	case x < 0 || x > 4 || y < 0 || y > 4:
		if print {
			fmt.Printf("Choose two numbers between 0 and 4, try again\n")
		}
	case bd[x][y] == 0:
		r = true
	case bd[x][y] != 0:
		if print {
			fmt.Printf("Cell (%d, %d) already occupied, try again\n", x, y)
		}
	}
	return r
}

func scanMove() (int, int) {
	var x, y int

	for looping := true; looping; {

		n, err := fmt.Scanf("%d %d\n", &x, &y)
		if err == io.EOF {
			os.Exit(0)
		}
		if n != 2 {
			fmt.Printf("Give two numbers\n")
			continue
		}
		if err != nil {
			fmt.Printf("Failed to read: %v\n", err)
			continue
		}
		looping = false
	}
	return x, y
}

func setScores(randomize bool) {
	if randomize {
		vals := [11]int{-5, -4, -3 - 2, -1, 0, 1, 2, 3, 4, 5}
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

// States for book offense/defense openings.
// Openings implemented as state machines.
const (
	FIRST = iota
	DIAGONAL
	CORNER
	OTHERCORNER
	OTHERDIAGONAL
	LAST
)

var firstMoves = [4][2]int{
	{0, 0},
	{0, 1},
	{1, 0},
	{1, 1},
}

func bookDefend(bd *Board, firstX int, firstY int) int {
	state := FIRST
	moveCount := 0

	var cX, cY int

	for state != LAST {
		switch state {
		case FIRST:
		OUTERFIRST:
			for i := -3; i < 4; i += 6 {
				a := firstX + i
				for j := -3; j < 4; j += 6 {
					b := firstY + j
					if a >= 0 && a <= 4 && b >= 0 && b <= 4 {
						// Since <firstX, firstY> have an X, <a,b> must be empty
						cX, cY = a, b
						bd[cX][cY] = MAXIMIZER
						moveCount++
						break OUTERFIRST
					}
				}
			}
			state = DIAGONAL

			fmt.Printf("My move: %d %d\n", cX, cY)
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
			cX, cY = -1, -1
		FOUNDMOVE:
			for i := -3; i < 4; i += 6 {
				a := lastx + i
				for j := -3; j < 4; j += 6 {
					b := lasty + j
					if a >= 0 && a <= 4 && b >= 0 && b <= 4 {
						if bd[a][b] == UNSET {
							bd[a][b] = MAXIMIZER
							moveCount++
							cX, cY = a, b
							break FOUNDMOVE
						}
					}
				}
			}
			if cX == -1 || cY == -1 {
				// human not following "triangle" book opening,
				// this function only defends that.
				break // out of for-loop over state
			}
			fmt.Printf("My move: %d %d\n", cX, cY)
			printBoard(bd)
		}
	}

	return moveCount
}

func bookStart(bd *Board) int {

	state := FIRST
	moveCount := 0

	var cX, cY int

	for state != LAST {
		switch state {
		case FIRST:
			c := firstMoves[rand.Intn(4)]
			cX, cY = c[0], c[1]
			bd[cX][cY] = MAXIMIZER
			state = DIAGONAL
			moveCount++
		case DIAGONAL:
			pX, pY := cX+3, cY+3
			if bd[pX][pY] == UNSET {
				cX += 3
				cY += 3
				bd[cX][cY] = MAXIMIZER
				moveCount++
				state = CORNER
			} else {
				state = OTHERCORNER
			}
		case CORNER:
			pX, pY := cX-3, cY
			if bd[pX][pY] == UNSET {
				cX -= 3
				bd[cX][cY] = MAXIMIZER
				moveCount++
			} else {
				pX, pY = cX, cY-3
				if bd[pX][pY] == UNSET {
					cY -= 3
					bd[cX][cY] = MAXIMIZER
					moveCount++
				}
			}
			state = LAST
		case OTHERCORNER:
			// Didn't get desired diagonal
			pX, pY := cX, cY+3
			if bd[pX][pY] == UNSET {
				cY += 3
				bd[cX][cY] = MAXIMIZER
				moveCount++
				state = OTHERDIAGONAL
			} else {
				fmt.Printf("Unreachable state in OTHERCORNER\n")
				printBoard(bd)
				os.Exit(99)
			}
		case OTHERDIAGONAL:
			pX, pY := cX+3, cY-3
			if bd[pX][pY] == UNSET {
				cX += 3
				cY -= 3
				bd[cX][cY] = MAXIMIZER
				moveCount++
			}
			state = LAST
		}
		if (moveCount % 2) == 1 {
			fmt.Printf("My move: %d %d\n", cX, cY)
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

type moveKeeper struct {
	moves [25][2]int
	next  int // index into moves[]
	max   int
}

func (p *moveKeeper) setMove(a, b int, value int) {
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

func (p *moveKeeper) chooseMove(deterministic bool) (x, y int, value int) {

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
