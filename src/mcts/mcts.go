package mcts

import (
	"fmt"
	"math"
	"math/rand"
)

type GameState struct {
	playerJustMoved int
	board           [25]int
	cachedResults   [3]float64
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

// Manifest constants to improve understanding
const (
	MAXIMIZER = 1
	MINIMIZER = -1
	UNSET     = 0
)

type MCTS struct {
	game       *GameState
	iterations int
	movesNode  *Node
	UCTK       float64
}

func New(deterministic bool, maxdepth int) *MCTS {
	return &MCTS{game: NewGameState(), iterations: 1000000, UCTK: 1.0}
}

func (p *MCTS) Name() string {
	return "MCTS"
}

func (p *MCTS) SetUCTK(UCTK float64) {
	p.UCTK = UCTK
}

func (p *MCTS) MakeMove(x, y int, player int) {
	p.game.board[5*x+y] = player
	p.game.playerJustMoved = player
	p.updateMoves(5*x + y)
}

func (p *MCTS) SetDepth(moveCounter int) {
}

// ChooseMove should choose computer's next move and
// return x,y coords of move and its score.
func (p *MCTS) ChooseMove() (xcoord int, ycoord int, value int, leafcount int) {

	bestnode, leaves, value := UCT(p.game, p.iterations, p.UCTK, p.movesNode)

	p.movesNode = bestnode
	p.movesNode.parentNode = nil

	move := p.movesNode.move

	p.game.DoMove(move)

	a := move / 5
	b := move % 5

	return a, b, value, leaves
}

func (p *MCTS) PrintBoard() {
	fmt.Printf("%v", p.game)
}

func (p *MCTS) SetScores(randomize bool) {
}

// FindWinner will return MAXIMIZER or MINIMIZER if somebody won,
// UNSET if nobody wins based on current board.
func (p *MCTS) FindWinner() int {
	board := p.game.board
	for _, i := range importantCells {
		if board[i] != UNSET {
			for _, quad := range winningQuads[i] {
				sum := board[quad[0]] + board[quad[1]] + board[quad[2]] + board[quad[3]]
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
		if board[i] != UNSET {
			for _, triplet := range losingTriplets[i] {
				sum := board[triplet[0]] + board[triplet[1]] + board[triplet[2]]
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

func UCT(rootstate *GameState, itermax int, UCTK float64, rootnode *Node) (*Node, int, int) {

	leafNodeCount := 0

	if rootnode == nil {
		rootnode = NewNode(-1, nil, rootstate)
		if len(rootnode.untriedMoves) == 25 {
			// Computer has the first move on the board.
			// Include only 1/4 of the cells in rootnode.untriedMoves,
			// since games can be rotated and mirrored to always start there.
			var moves []int
			switch rand.Intn(4) {
			case 0: // Upper left
				moves = []int{0, 1, 2, 5, 6, 7, 10, 11, 12}
			case 1: // Upper right
				moves = []int{2, 3, 4, 7, 8, 9, 12, 13, 14}
			case 2: // Lower left
				moves = []int{10, 11, 12, 15, 16, 17, 20, 21, 22}
			case 3: // Lower right
				moves = []int{12, 13, 14, 17, 18, 19, 22, 23, 24}
			}
			rootnode.untriedMoves = moves
			itermax *= 2 // only 9 first moves, can do more iterations
		}
	} else {
		rootnode.playerJustMoved = rootstate.playerJustMoved
	}

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

		// starting with current state, pick a random
		// branch of the game tree, all the way to a win/loss.
		for !terminalNode {
			m := moves[rand.Intn(len(moves))]
			state.DoMove(m)
			moves, terminalNode = state.GetMoves()
		}

		leafNodeCount++

		// node now points to a board where a player won
		// and the other lost. Trace back up the tree, updating
		// each node's wins and visit count.

		state.resetCachedResults()
		for ; node != nil; node = node.parentNode {
			node.Update(state.GetResult(node.playerJustMoved))
		}
	}

	// The "value" of this move is somewhat fictitious, and
	// not related to Negascout or any minimax value function.
	moveChoice := rootnode.bestMove(UCTK)
	return moveChoice, leafNodeCount, int(1000. * moveChoice.UCB1(UCTK))
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
	if bestmove == nil {
		fmt.Printf("Node has %d children\n", len(p.childNodes))
		fmt.Printf("player just moved %d\n", p.playerJustMoved)
		fmt.Printf("Move %d, wins %v visits %v\n", p.move, p.wins, p.visits)
		for i, c := range p.childNodes {
			ucb1 := c.UCB1(UCTK)
			fmt.Printf("Child %d score %v\n", i, ucb1)
		}
	}
	return bestmove
}

func (p *Node) String() string {
	return fmt.Sprintf("Move %d, parent %p, childNodes %v, wins %f, visits %f, %d untried, %d moved", p.move, p.parentNode, p.childNodes, p.wins, p.visits, len(p.untriedMoves), p.playerJustMoved)
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
	for i, m := range p.untriedMoves {
		if m == move {
			p.untriedMoves = append(p.untriedMoves[:i], p.untriedMoves[i+1:]...)
			break
		}
	}
	n := NewNode(move, p, st)
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

func (p *GameState) resetCachedResults() {
	p.cachedResults[0] = -1
	p.cachedResults[1] = -1
	p.cachedResults[2] = -1
}

func (p *GameState) DoMove(move int) {
	p.playerJustMoved = -p.playerJustMoved
	p.board[move] = p.playerJustMoved
}

func (p *MCTS) updateMoves(m int) {
	if p.movesNode != nil {
		for _, childNode := range p.movesNode.childNodes {
			if childNode.move == m {
				p.movesNode = childNode
				p.movesNode.parentNode = nil
				break
			}
		}

	}
}

func (p GameState) GetMoves() ([]int, bool) {

	// Only have to check the 9 cells in importantCells[]
	// for 4 or 3 in a row configs.
	for _, m := range importantCells {
		if p.board[m] != UNSET {
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
	}

	// Get here, p.board does not represent a win or a loss.
	// Pick out empty cells in p.board for a list of valid moves.
	// I don't believe "cat" games exist in Squava, but this code
	// handles that case.

	endOfGame := true
	var moves []int
	for i := 0; i < 25; i++ {
		if p.board[i] == UNSET {
			endOfGame = false
			moves = append(moves, i)
		}
	}

	return moves, endOfGame
}

func (p *GameState) GetResult(playerjm int) float64 {
	cached := p.cachedResults[playerjm+1]
	if cached >= 0.0 {
		return cached
	}
	// Need to check all 4-in-a-row wins before checking
	// any 3-in-a-row losses, otherwise the result ends
	// up wrong.
	for _, i := range importantCells {
		if p.board[i] != UNSET {
			for _, quad := range winningQuads[i] {
				sum := p.board[quad[0]] + p.board[quad[1]] + p.board[quad[2]] + p.board[quad[3]]
				if sum == 4 || sum == -4 {
					if sum == 4*playerjm {
						p.cachedResults[playerjm+1] = 1.0
						return 1.0
					}
					p.cachedResults[playerjm+1] = 0.0
					return 0.0
				}
			}
		}
	}
	for _, i := range importantCells {
		if p.board[i] != UNSET {
			for _, triplet := range losingTriplets[i] {
				sum := p.board[triplet[0]] + p.board[triplet[1]] + p.board[triplet[2]]
				if sum == 3 || sum == -3 {
					if sum == 3*playerjm {
						p.cachedResults[playerjm+1] = 0.0
						return 0.0
					}
					p.cachedResults[playerjm+1] = 1.0
					return 1.0
				}
			}
		}
	}
	p.cachedResults[playerjm+1] = 0.0
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
