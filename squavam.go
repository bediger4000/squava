package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	// "sort"
	"time"
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

func main() {

	iterMax := flag.Int("i", 10000, "maximum iterations")
	computerFirstPtr := flag.Bool("C", false, "Computer takes first move")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())

	state := NewGameState()

	if *computerFirstPtr {
		state.playerJustMoved = 1
	} else {
		state.playerJustMoved = -1
	}

	// Why do we ignore the list of valid moves? Should be usable.
	for _, endOfGame := state.GetMoves(); !endOfGame; _, endOfGame = state.GetMoves(){
		var m int
		fmt.Printf("%v\n", state)
		if state.playerJustMoved == 1 {
			start := time.Now()
			m = UCT(state, *iterMax)
			end := time.Now()
			fmt.Printf("My move: %d %d %v\n", m/5, m%5, end.Sub(start))
		} else {
			m = readMove(state.board)
		}
		state.DoMove(m)
	}

	fmt.Printf("%v\n", state)

	lastPlayer := string("O_X"[state.playerJustMoved+1])
	switch state.GetResult(state.playerJustMoved) {
	case 1.0:
		fmt.Printf("Player %s wins!\n", lastPlayer)
	case 0.0:
		fmt.Printf("Player %s loses!\n", lastPlayer)
	default:
		fmt.Printf("Nobody wins!\n")
	}
}

func UCT(rootstate *GameState, itermax int) int {
	rootnode := NewNode(-1, nil, rootstate)

	for i := 0; i < itermax; i++ {

		node := rootnode
		state := rootstate.Clone()

		for len(node.untriedMoves) == 0 && len(node.childNodes) > 0 {
			node = node.UCTSelectChild()
			state.DoMove(node.move)
		}

		if len(node.untriedMoves) > 0 {
			m := node.untriedMoves[rand.Intn(len(node.untriedMoves))]
			state.DoMove(m)
			node = node.AddChild(m, state)
		}

		moves, terminalNode := state.GetMoves()

if len(moves) == 0 && !terminalNode {
	fmt.Printf("Big problem, 0 moves, but not terminal?\n")
	fmt.Printf("Moves: %v\n", moves)
	fmt.Printf("State:\n%v\n", state)
	fmt.Printf("State: %v\n", state.String2())
}

		for  !terminalNode  {
			m := moves[rand.Intn(len(moves))]
			state.DoMove(m)
			moves, terminalNode = state.GetMoves()
		}

		for ; node != nil; node = node.parentNode {
			node.Update(state.GetResult(node.playerJustMoved))
		}
	}

/*
	for idx, cn := range rootnode.childNodes {
		fmt.Printf("child %d: %f  %v\n", idx, cn.UCB1(1.0), cn)
	}
*/

	return rootnode.bestMove().move
}

func NewNode(move int, parent *Node, state *GameState) *Node {
	var n Node
	n.move = move
	n.parentNode = parent
	n.untriedMoves, _ = state.GetMoves()
	n.playerJustMoved = state.playerJustMoved
	return &n
}

func (p *Node) bestMove() *Node {
	bestscore := math.SmallestNonzeroFloat64
	var bestmove *Node
	for _, c := range p.childNodes {
		ucb1 := c.UCB1(1.0)
		if ucb1 > bestscore {
			bestscore = ucb1
			bestmove = c
		}
	}
	return bestmove
}

func (p *Node) String() string {
	return fmt.Sprintf("%c, %d: %f/%f, %d %d, U:%v\n", "O_X"[p.playerJustMoved+1], p.move, p.wins, p.visits, len(p.untriedMoves), len(p.childNodes), p.untriedMoves)
}

func (p *Node) UCTSelectChild() *Node {
	/*
	   """ Use the UCB1 formula to select a child node. Often a constant
	   UCTK is applied so we have lambda c: c.wins/c.visits + UCTK *
	   sqrt(2*log(self.visits)/c.visits to vary the amount of exploration
	   versus exploitation.
	   """
	   s = sorted(self.childNodes, key = lambda c: c.wins/c.visits + sqrt(2*log(self.visits)/c.visits))[-1]
	   return s
	*/
/*
	sort.Sort(p)
	return p.childNodes[0]
*/
	return p.bestMove()
}

func (p *Node) UCB1(UCTK float64) float64 {
	return p.wins/(p.visits+math.SmallestNonzeroFloat64) + UCTK*math.Sqrt(2.*math.Log(p.parentNode.visits)/(p.visits+math.SmallestNonzeroFloat64))
}

func (p *Node) Len() int {
	return len(p.childNodes)
}

func (p *Node) Swap(i, j int) {
	p.childNodes[i], p.childNodes[j] = p.childNodes[j], p.childNodes[i]
}

func (p *Node) Less(i, j int) bool {
	// Seems like a waste of time to calculate UCB1 value every Less() call
	key1 := p.childNodes[i].UCB1(2.0)
	key2 := p.childNodes[j].UCB1(2.0)
	return key1 > key2
}

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
	st.board = p.board  // copy since board has type [25]int
	return &st
}

func (p *GameState) DoMove(move int) {
	p.playerJustMoved = -p.playerJustMoved
	p.board[move] = p.playerJustMoved
}

func (p *GameState) GetMoves() ([]int, bool) {
	var moves []int
	for i := 0; i < 25; i++ {
		if p.board[i] == 0 {
			moves = append(moves, i)
		} else {
			for _, quad := range winningQuads[i] {
				sum := p.board[quad[0]] + p.board[quad[1]] + p.board[quad[2]] + p.board[quad[3]]
				if sum == 4 || sum == -4 {
					return []int{}, true
				}
			}
			for _, triplet := range losingTriplets[i] {
				sum := p.board[triplet[0]] + p.board[triplet[1]] + p.board[triplet[2]]
				if sum == 3 || sum == -3 {
					return []int{}, true
				}
			}
		}
	}
	endOfGame := false
	if len(moves) == 0 {
		endOfGame = true
	}
	return moves, endOfGame
}

func (p *GameState) GetResult(playerjm int) float64 {
	for i := 0; i < 25; i++ {
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
	for i := 0; i < 25; i++ {
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
	return 0.0
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

func (p *GameState) String2() string {
    return fmt.Sprintf("%d, %v", p.playerJustMoved, p.board)
}

func readMove(bd [25]int) int {
	readMove := false
	var x, y, m int
	for !readMove {
		fmt.Printf("Enter move (N M): ")
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
			fmt.Printf("Choose two numbers between 0 and 4, try again\n")
		default:
			m = 5*x + y
			if	bd[m] == 0 {
				readMove = true
			} else {
				fmt.Printf("Cell (%d, %d) already occupied, try again\n", x, y)
			}
		}
	}
	return m
}

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
