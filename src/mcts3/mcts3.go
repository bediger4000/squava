package mcts3

import (
	"fmt"
	"math/rand"
)

type GameState struct {
	player        int
	board         [25]int
	cachedResults [3]float64
}

type Node struct {
	move       int
	player     int
	parent     *Node
	childNodes []*Node
	wins       float64
	visits     float64
	score      float64
	// score should be 0 for a losing move,
	// 1 for a winning move
	untriedMoves []int
	winner       int
}

// Manifest constants to improve understanding
const (
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

type MCTS3 struct {
	board           [25]int
	playerJustMoved int
	iterations      int
}

func New(_ bool, _ int) *MCTS3 {
	return &MCTS3{iterations: 500000}
}

func (p *MCTS3) Name() string {
	return "MCTS/Plain"
}

func (p *MCTS3) SetIterations(iterations int) {
	p.iterations = iterations
}

func (p *MCTS3) MakeMove(x, y int, player int) {
	p.board[5*x+y] = player
	p.playerJustMoved = player
}

func (p *MCTS3) SetDepth(_ int) {
}

// ChooseMove should choose computer's next move and
// return x,y coords of move and its score.
func (p *MCTS3) ChooseMove() (xcoord int, ycoord int, value int, leafcount int) {

	var best int

	best, value, leafcount = bestMove(p.board, p.iterations)

	p.board[best] = MAXIMIZER

	// Since this player's "board" is a plain array, a move has to translate
	// to <x,y> coords
	xcoord = best / 5
	ycoord = best % 5
	fmt.Printf("best move %d <%d,%d>\n", best, xcoord, ycoord)

	return
}

func bestMove(board [25]int, iterations int) (move int, value int, leafCount int) {

	fmt.Printf("enter bestMove, %d iterations\n", iterations)

	root := &Node{
		player: MINIMIZER,
	}
	root.untriedMoves = make([]int, 0, 25)
	for i := range board {
		if board[i] == UNSET {
			root.untriedMoves = append(root.untriedMoves, i)
		}
	}

	var state GameState

	for iters := 0; iters < iterations; iters++ {
		fmt.Printf("iteration %d\n", iters)
		for j := 0; j < 25; j++ {
			state.board[j] = board[j]
		}
		state.player = MINIMIZER

		node := root

		fmt.Printf("Selection\n")
		// Selection
		for len(node.untriedMoves) == 0 && len(node.childNodes) > 0 {
			fmt.Printf("Node move %d/player %d has best child ", node.move, node.player)
			node = node.selectBestChild()
			fmt.Printf("with move %d, player %d,  %.0f/%.0f/%.3f ", node.move, node.player, node.wins, node.visits, node.score)
			state.player = 0 - state.player
			state.board[node.move] = state.player
			fmt.Printf("board now:\n%s\n", boardString(state.board))
		}

		// state should represent the board resulting from following
		// the "best child" nodes, and node points to a Node struct
		// that has no child nodes.

		var win bool

		fmt.Printf("Expansion\n")
		// Expansion
		if len(node.untriedMoves) > 0 {
			mv := node.untriedMoves[rand.Intn(len(node.untriedMoves))]

			state.player = 0 - state.player
			state.board[mv] = state.player
			fmt.Printf("expansion move %d, board:\n%s\n", mv, boardString(state.board))

			node = node.AddChild(mv, state.player) // AddChild take mv out of untriedMoves slice
			node.winner = findWinner(&(state.board))
			if node.winner == MAXIMIZER {
				node.score = 1.0
				win = true
			}
			// node represents mv, the previously untried move
		}

		fmt.Printf("Simulation\n")
		// Simulation
		if node.winner == UNSET {
			moves := (&state).RemainingMoves()

			for len(moves) > 0 {
				m := moves[rand.Intn(len(moves))]
				state.player = 0 - state.player
				state.board[m] = state.player
				winner := findWinner(&(state.board))
				if winner != UNSET {
					if winner == MAXIMIZER {
						win = true
					}
					break
				}
				cutElement(&moves, m)
			}
		}

		fmt.Printf("board after playout:\n%s\n", boardString(state.board))

		leafCount++

		winIncr := 0.0
		if win {
			winIncr = 1.0
		}

		fmt.Printf("Back propagation: win %v\n", win)
		// Back propagation
		for node != nil {
			node.visits += 1.0
			if node.winner == UNSET {
				node.wins += winIncr
				node.score = node.wins / node.visits
			}
			fmt.Printf("node move %d, player %d: %.0f/%.0f/%.3f\n",
				node.move, node.player, node.wins, node.visits, node.score,
			)
			node = node.parent
		}
	}

	moveNode := root.selectBestChild()
	move = moveNode.move

	return
}

// cutElement removes element from slice ary
// that has value v. Disorders ary.
func cutElement(ary *[]int, v int) {
	for i, m := range *ary {
		if m == v {
			(*ary)[i] = (*ary)[len(*ary)-1]
			*ary = (*ary)[:len((*ary))-1]
			break
		}
	}
}

func (node *Node) AddChild(mv int, player int) *Node {
	fmt.Printf("node.AddChild(%d, %d)\n", mv, player)
	ch := &Node{
		move:   mv,
		parent: node,
		player: player,
	}
	node.childNodes = append(node.childNodes, ch)
	// weed out mv as an untried move
	cutElement(&(node.untriedMoves), mv)

	fmt.Printf("Child nodes %d:\n", len(node.childNodes))
	for _, n := range node.childNodes {
		fmt.Printf("\tmove %d player %d, %.0f/%.0f/%.3f\n", n.move, n.player, n.wins, n.visits, n.score)
	}
	fmt.Printf("untried moves: %v\n", node.untriedMoves)

	return ch
}

func (node *Node) selectBestChild() *Node {
	best := node.childNodes[0]
	bestScore := node.childNodes[0].score

	for _, c := range node.childNodes {
		if c.score > bestScore {
			best = c
			bestScore = c.score
		}
	}

	return best
}

// RemainingMoves returns an array of all moves left
// unmade on state.board
func (state *GameState) RemainingMoves() []int {
	mvs := make([]int, 25)
	j := 0
	for i := 0; i < 25; i++ {
		if state.board[i] == UNSET {
			mvs = append(mvs, i)
			j++
		}
	}
	return mvs
}

func (p *MCTS3) PrintBoard() {
	fmt.Printf("%s\n", p)
}

func (p *MCTS3) SetScores(_ bool) {
}

// FindWinner will return MAXIMIZER or MINIMIZER if somebody won,
// UNSET if nobody wins based on current board.
func (p *MCTS3) FindWinner() int {
	return findWinner(&(p.board))
}

// findWinner will return MAXIMIZER or MINIMIZER if somebody won,
// UNSET if nobody wins based on argument board.
func findWinner(board *[25]int) int {
	for _, i := range importantCells {
		if (*board)[i] != UNSET {
			for _, quad := range winningQuads[i] {
				sum := (*board)[quad[0]] + (*board)[quad[1]] + (*board)[quad[2]] + (*board)[quad[3]]
				switch sum {
				case 4:
					return MAXIMIZER
				case -4:
					return MINIMIZER
				}
			}
		}
	}
	for _, i := range importantCells {
		if (*board)[i] != UNSET {
			for _, triplet := range losingTriplets[i] {
				sum := (*board)[triplet[0]] + (*board)[triplet[1]] + (*board)[triplet[2]]
				switch sum {
				case 3:
					return MINIMIZER
				case -3:
					return MAXIMIZER
				}
			}
		}
	}
	return UNSET
}

func (p *MCTS3) String() string {
	return boardString(p.board)
}

func boardString(board [25]int) string {
	s := "   0 1 2 3 4\n"
	for i := 0; i < 25; i++ {
		if (i % 5) == 0 {
			s += string((i/5)+'0') + "  "
		}
		s += string("O_X"[board[i]+1]) + " "
		if (i % 5) == 4 {
			s += "\n"
		}
	}
	return s
}

var importantCells = [9]int{2, 7, 10, 11, 12, 13, 14, 17, 22}

// 25 rows only to make looping easier. The filled-in
// rows are the only quads you actually have to check
// to find out if there's a win
var winningQuads = [25][][]int{
	{}, {},
	{{0, 1, 2, 3}, {1, 2, 3, 4}, {2, 7, 12, 17}},
	{}, {}, {}, {},
	{{5, 6, 7, 8}, {6, 7, 8, 9}, {7, 12, 17, 22}},
	{}, {},
	{{0, 5, 10, 15}, {5, 10, 15, 20}},
	{{1, 6, 11, 16}, {6, 11, 16, 21}, {3, 7, 11, 15}, {5, 11, 17, 23}},
	{{10, 11, 12, 13}, {11, 12, 13, 14}, {4, 8, 12, 16}, {8, 12, 16, 20}, {0, 6, 12, 18}, {6, 12, 18, 24}},
	{{3, 8, 13, 18}, {8, 13, 18, 23}, {1, 7, 13, 19}, {9, 13, 17, 21}},
	{{4, 9, 14, 19}, {9, 14, 19, 24}},
	{}, {},
	{{15, 16, 17, 18}, {16, 17, 18, 19}},
	{}, {}, {}, {},
	{{20, 21, 22, 23}, {21, 22, 23, 24}},
	{}, {},
}

// 25 rows only to make looping easier. The filled-in
// rows are the only triplets you actually have to check
// to find out if there's a loss.
var losingTriplets = [][][]int{
	{}, {},
	{{0, 1, 2}, {1, 2, 3}, {2, 3, 4}, {2, 7, 12}, {2, 6, 10}, {14, 8, 2}},
	{}, {}, {}, {},
	{{5, 6, 7}, {6, 7, 8}, {7, 8, 9}, {2, 7, 12}, {7, 12, 17}, {3, 7, 11}, {7, 11, 15}, {1, 7, 13}, {7, 13, 19}},
	{}, {},
	{{10, 11, 12}, {0, 5, 10}, {5, 10, 15}, {10, 15, 20}, {2, 6, 10}, {10, 16, 22}},
	{{10, 11, 12}, {11, 12, 13}, {1, 6, 11}, {6, 11, 16}, {11, 16, 21}, {3, 7, 11}, {7, 11, 15}, {5, 11, 17}, {11, 17, 23}},
	{{10, 11, 12}, {11, 12, 13}, {12, 13, 14}, {2, 7, 12}, {7, 12, 17}, {12, 17, 22}, {0, 6, 12}, {6, 12, 18}, {12, 18, 24}, {4, 8, 12}, {8, 12, 16}, {12, 16, 20}},
	{{11, 12, 13}, {12, 13, 14}, {3, 8, 13}, {8, 13, 18}, {13, 18, 23}, {1, 7, 13}, {7, 13, 19}, {21, 17, 13}, {17, 13, 9}},
	{{12, 13, 14}, {4, 9, 14}, {9, 14, 19}, {14, 19, 24}, {22, 18, 14}, {14, 8, 2}},
	{}, {},
	{{15, 16, 17}, {16, 17, 18}, {17, 18, 19}, {7, 12, 17}, {12, 17, 22}, {5, 11, 17}, {11, 17, 23}, {21, 17, 13}, {17, 13, 9}},
	{}, {}, {}, {},
	{{20, 21, 22}, {21, 22, 23}, {22, 23, 24}, {12, 17, 22}, {10, 16, 22}, {22, 18, 14}},
	{}, {},
}
