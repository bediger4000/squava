package mcts

import (
	"fmt"
	"math"
	"math/rand"
)

type GameState struct {
	playerJustMoved int
	board           [25]int
}

type Node struct {
	move            int
	parentNode      *Node
	childNodes      []*Node
	wins            float64
	visits          float64
	untriedMoves    []int
	playerJustMoved int
}

const (
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

type MCTS struct {
	game       *GameState
	iterations int
}

func New(deterministic bool, maxdepth int) *MCTS {
	var r MCTS
	r.game = new(GameState)
	r.iterations = 150000
	return &r
}

func (p *MCTS) Name() string {
	return "MCTS"
}

func (p *MCTS) MakeMove(x, y int, player int) {
	// fmt.Printf("MakeMove(%d,%d) <%d>, player %d\n", x, y, 5*x+y, player)
	p.game.board[5*x + y] = player
	p.game.playerJustMoved = player
}

func (p *MCTS) SetDepth(moveCounter int) {
}

// Choose computer's next move: return x,y coords of
// move and its score.
func (p *MCTS) ChooseMove() (xcoord int, ycoord int, value int, leafcount int) {

	// fmt.Printf("ChooseMove(%p) gamestate %p\n", p, p.game)

	move, leaves, value := UCT(p.game, p.iterations, 1.00)

	a := move / 5
	b := move % 5

	return a, b, value, leaves
}

func (p *MCTS) PrintBoard() {
	fmt.Printf("%v", p)
}

func (p *MCTS) SetScores(randomize bool) {
}

func (p *MCTS) FindWinner() int {
	m := p.game.playerJustMoved
	w := p.game.GetResult(m)
	if w > 0.0 {
		return m
	}
	return 0
}

func UCT(rootstate *GameState, itermax int, UCTK float64) (int, int, int) {

	leafNodeCount := 0
	rootnode := NewNode(-1, nil, rootstate)

	for i := 0; i < itermax; i++ {

		node := rootnode           // reset node to root of tree of nodes
		state := rootstate.Clone() // start at rootstate, rootnode's GameState

		for len(node.untriedMoves) == 0 && len(node.childNodes) > 0 {
			node = node.UCTSelectChild(UCTK) // updates node
			state.DoMove(node.move)
		}

		// This condition creates a child node from an untried move
		// (if any exist), makes the move in state, and makes node
		// the child node.
		if len(node.untriedMoves) > 0 {
			m := node.untriedMoves[rand.Intn(len(node.untriedMoves))]
			state.DoMove(m)
			node = node.AddChild(m, state)
			// node now represents m, the previously-untried move.
		}

		moves, terminalNode := state.GetMoves()
		if !terminalNode {
			leafNodeCount++
		}

		// starting with current state, pick a random
		// branch of the game tree, all the way to a win/loss.
		for !terminalNode {
			m := moves[rand.Intn(len(moves))]
			state.DoMove(m)
			moves, terminalNode = state.GetMoves()
		}

		// node now points to a board where a player won
		// and the other lost. Trace back up the tree, updating
		// each node's wins and visit count.

		for ; node != nil; node = node.parentNode {
			node.Update(state.GetResult(node.playerJustMoved))
		}
	}

	bestMove := rootnode.bestMove(UCTK)
	return bestMove.move, leafNodeCount, int(1000.*bestMove.UCB1(1.00))
}

func NewNode(move int, parent *Node, state *GameState) *Node {
	var n Node
	n.move = move
	n.parentNode = parent
	n.untriedMoves, _ = state.GetMoves()
	n.playerJustMoved = state.playerJustMoved
	return &n
}

// Since there's at most 25 moves to consider,
// just look through them rather than incur
// sorting overhead. It seems like maybe caching
// the best child node might help performance. Have
// to figure out how to track changes to childNodes[],
// because a change to that array invalidates the
// choice of "best" move. Also, does UCB1() score for
// a given child node stay the same? I don't think it does.
func (p *Node) bestMove(UCTK float64) *Node {
	bestscore := math.SmallestNonzeroFloat64
	var bestmove *Node
	for _, c := range p.childNodes {
		ucb1 := c.UCB1(UCTK)
		if ucb1 > bestscore {
			bestscore = ucb1
			bestmove = c
		}
	}
	return bestmove
}

func (p *Node) UCTSelectChild(UCTK float64) *Node {
	return p.bestMove(UCTK)
}

func (p *Node) UCB1(UCTK float64) float64 {
	return p.wins/(p.visits+math.SmallestNonzeroFloat64) + UCTK*math.Sqrt(2.*math.Log(p.parentNode.visits)/(p.visits+math.SmallestNonzeroFloat64))
}

// AddChild creates a new *Node with the state of st
// argument, takes move out of p.untriedMoves, adds
// the new *Node to the array of child nodes, returns
// the new *Node, which is then a child of p.
func (p *Node) AddChild(move int, st *GameState) *Node {
	n := NewNode(move, p, st)
	for i, m := range p.untriedMoves {
		if m == move {
			p.untriedMoves = append(p.untriedMoves[:i], p.untriedMoves[i+1:]...)
			break
		}
	}
	p.childNodes = append(p.childNodes, n)
	return n
}

func (p *Node) Update(result float64) {
	p.visits++
	p.wins += result
}

func NewGameState() *GameState {
	var st GameState
	st.playerJustMoved = -1
	return &st
}

func (p *GameState) Clone() *GameState {
	var st GameState
	st.playerJustMoved = p.playerJustMoved
	st.board = p.board // copy since board has type [25]int
	return &st
}

func (p *GameState) DoMove(move int) {
	p.playerJustMoved = -p.playerJustMoved
	p.board[move] = p.playerJustMoved
}

func (p *GameState) GetMoves() ([]int, bool) {

	// Only have to check the 9 cells in important_cells[]
	// for 4 or 3 in a row configs.
	for _, m := range important_cells {
		for _, quad := range winningQuads[m] {
			sum := p.board[quad[0]] + p.board[quad[1]] + p.board[quad[2]] + p.board[quad[3]]
			if sum == 4 || sum == -4 {
				return []int{}, true
			}
		}
		for _, triplet := range losingTriplets[m] {
			sum := p.board[triplet[0]] + p.board[triplet[1]] + p.board[triplet[2]]
			if sum == 3 || sum == -3 {
				return []int{}, true
			}
		}
	}

	// Get here, p.board does not represent a win or a loss.
	// Pick out empty cells in p.board for a list of valid moves.
	// I don't believe "cat" games exist in Squava, but this code
	// handles that case.

	endOfGame := true
	var moves []int
	for i := 0; i < 25; i++ {
		if p.board[i] == 0 {
			endOfGame = false
			moves = append(moves, i)
		}
	}

	return moves, endOfGame
}

func (p *GameState) GetResult(playerjm int) float64 {
	// Need to check all 4-in-a-row wins before checking
	// any 3-in-a-row losses, otherwise the result ends
	// up wrong.
	for _, i := range important_cells {
		for _, quad := range winningQuads[i] {
			sum := p.board[quad[0]] + p.board[quad[1]] + p.board[quad[2]] + p.board[quad[3]]
			if sum == 4 || sum == -4 {
				if sum == 4*playerjm {
					return 1.0
				} else {
					return 0.0
				}
			}
		}
	}
	for _, i := range important_cells {
		for _, triplet := range losingTriplets[i] {
			sum := p.board[triplet[0]] + p.board[triplet[1]] + p.board[triplet[2]]
			if sum == 3 || sum == -3 {
				if sum == 3*playerjm {
					return 0.0
				} else {
					return 1.0
				}
			}
		}
	}
	return 0.0 // Should probably never get here.
}

func (p *GameState) String() string {
	s := "   0 1 2 3 4\n"
	for i := 0; i < 25; i++ {
		if (i % 5) == 0 {
			s += string((i/5)+'0') + "  "
		}
		s += string("O_X"[p.board[i]+1]) + " "
		if (i % 5) == 4 {
			s += "\n"
		}
	}
	return s
}

var important_cells [9]int = [9]int{2, 7, 10, 11, 12, 13, 14, 17, 22}

// 25 rows only to make looping easier. The filled-in
// rows are the only quads you actually have to check
// to find out if there's a win
var winningQuads [25][][]int = [25][][]int{
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
var losingTriplets [][][]int = [][][]int{
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
