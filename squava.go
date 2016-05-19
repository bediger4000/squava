package main

import (
	"fmt"
	"os"
)

type Board [5][5]int

const WIN = 1000
const LOSS = -1000

var max_depth int = 4

func main() {
	var bd Board
	var max int = LOSS
	var move [2]int

	for i, row := range bd {
		for j, marker := range row {
			if marker == 0 {
				bd[i][j] = 1
				val := alphabeta(&bd, 1, -1, LOSS, WIN)
				bd[i][j] = 0
				fmt.Printf("Move (%d, %d) value %d\n", i, j, val)
				if val > max {
					move[0] = i
					move[1] = j
					max = val
				}
			}
		}
	}

	bd[move[0]][move[1]] = 1

	os.Exit(0)
}

func alphabeta(bd *Board, ply int, player int, alpha int, beta int) (value int) {
	winner, end_of_game := check_winner(bd)
	if end_of_game {
		switch winner {
		case 1:
			return WIN - ply
		case -1:
			return LOSS + ply
		case 0:
			return 0
		}
	}

	if ply == max_depth {
		return static_value(bd)
	}

	switch player {
	case 1:
		value = LOSS
		for i, row := range bd {
			for j, marker := range row {
				if marker == 0 {
					bd[i][j] = player
					n := alphabeta(bd, ply+1, -player, alpha, beta)
					bd[i][j] = 0
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
	case -1:
		value = WIN
		for i, row := range bd {
			for j, marker := range row {
				if marker == 0 {
					bd[i][j] = player
					n := alphabeta(bd, ply+1, -player, alpha, beta)
					bd[i][j] = 0
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

var scores [][]int = [][]int{
	[]int{1, 0, 0, 0, 1},
	[]int{0, 2, 2, 2, 0},
	[]int{0, 2, 3, 2, 0},
	[]int{0, 2, 2, 2, 0},
	[]int{1, 0, 0, 0, 1},
}

func static_value(bd *Board) (score int) {
	for i, row := range bd {
		for j, _ := range row {
			score += bd[i][j] * scores[i][j]
		}
	}
	return score
}

var losing_triplets [][][]int = [][][]int{
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
var winning_quads [][][]int = [][][]int{
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

func check_winner(bd *Board) (winner int, end_of_game bool) {
	for _, triplet := range losing_triplets {
		sum := bd[triplet[0][0]][triplet[0][1]]
		sum += bd[triplet[1][0]][triplet[1][1]]
		sum += bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			loser := sum / 3
			return -loser, true
		}
	}
	for _, quad := range winning_quads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			winner := sum / 4
			return winner, true
		}
	}
	for _, row := range bd {
		for _, val := range row {
			if val == 0 {
				return 0, false
			}
		}
	}
	// Get here, all 25 spots on board filled, no winning quadruplet
	return 0, true
}
