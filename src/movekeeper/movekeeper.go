package movekeeper

import (
	"math/rand"
)

// Struct and 2 functions to encapsulate tracking of
// best possible move.

type MoveKeeper struct {
	moves [25][2]int
	next  int        // index into moves[]
	max   int
	deterministic bool
}

func New(max int, deterministic bool) (*MoveKeeper) {
	var r MoveKeeper

	r.next = 0
	r.max = max
	r.deterministic = deterministic
	return &r
}

func (p *MoveKeeper) SetMove(a, b int, value int) {
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

func (p *MoveKeeper) ChooseMove() (x, y int, value int) {

	if p.next == 0 {
		// Loop over all 25 cells couldn't find any
		// empty cells. Cat got the game.
		return -1, -1, 0
	}

	r := 0
	if !p.deterministic {
		r = rand.Intn(p.next)
	}

	return p.moves[r][0], p.moves[r][1], p.max
}
