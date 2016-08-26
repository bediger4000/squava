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

const WIN = 1000
const LOSS = -1000

var max_depth int = 6

// Arrays of losing triplets and winning quads, indexed
// by <x,y> coords of all pairs composing each of the quads
// or triplets. Makes check_winner() a lot more efficient.
var indexed_losing_triplets [5][5][][][]int
var indexed_winning_quads [5][5][][][]int

var move_counter int = 0

func main() {

	human_first_ptr := flag.Bool("H", true, "Human takes first move")
	computer_first_ptr := flag.Bool("C", false, "Computer takes first move")
	max_depth_ptr := flag.Int("d", 6, "maximum lookahead depth")
	flag.Parse()

	// Set up for use by check_winner()
	for _, triplet := range losing_triplets {
		for _, pair := range triplet {
			indexed_losing_triplets[pair[0]][pair[1]] = append(indexed_losing_triplets[pair[0]][pair[1]], triplet)
		}
	}
	for _, quad := range winning_quads {
		for _, pair := range quad {
			indexed_winning_quads [pair[0]][pair[1]] = append(
				indexed_winning_quads [pair[0]][pair[1]], quad)
		}
	}

	var human_first bool = *human_first_ptr
	if *computer_first_ptr {
		human_first = false
	}
	max_depth = *max_depth_ptr

	rand.Seed(time.Now().UTC().UnixNano())

	var bd Board
	var winner int
	var end_of_game bool = false

	move_counter := 0

	for !end_of_game {

		var l, m int
		if human_first {
			l, m = read_move(&bd)
			bd[l][m] = -1
			move_counter++
		}

		human_first = true

		winner, end_of_game = check_winner(&bd, l, m)

		if end_of_game {
			break
		}

		var moves [25][2]int
		var next int = 0

		max := LOSS

		if move_counter < 4 { max_depth = 6 }
		if move_counter > 3 { max_depth = 8 }
		if move_counter > 10 { max_depth = *max_depth_ptr }

		for i, row := range bd {
			for j, mark := range row {
				if mark == 0 {
					bd[i][j] = 1
					val := alphabeta(&bd, 1, -1, LOSS, WIN, i, j)
					bd[i][j] = 0
					if val >= max {
						if val > max {
							max = val
							next = 0
						}
						moves[next][0] = i
						moves[next][1] = j
						next++
					}
				}
			}
		}

		r := rand.Intn(next)
		fmt.Printf("My move: %d %d (%d, %d, %d)\n", moves[r][0], moves[r][1], max, next, r)

		bd[moves[r][0]][moves[r][1]] = 1
		move_counter++

		print_board(&bd)

		winner, end_of_game = check_winner(&bd, moves[r][0], moves[r][1])
	}

	var phrase string
	switch winner {
	case 1:
		phrase = "X wins\n"
	case 0:
		phrase = "Cat wins\n"
	case -1:
		phrase = "O wins\n"
	}
	fmt.Printf(phrase)

	os.Exit(0)
}

func alphabeta(bd *Board, ply int, player int, alpha int, beta int, x int, y int) (value int) {
	winner, end_of_game := check_winner(bd, x, y)
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
					n := alphabeta(bd, ply+1, -player, alpha, beta, i, j)
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
					n := alphabeta(bd, ply+1, -player, alpha, beta, i, j)
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

func print_board(bd *Board) {
	fmt.Printf("   0 1 2 3 4\n")
	for i, row := range bd {
		fmt.Printf("%d  ", i)
		for _, v := range row {
			var marker string
			switch v {
			case 1:
				marker = "X"
			case -1:
				marker = "O"
			case 0:
				marker = "_"
			}
			fmt.Printf("%s ", marker)
		}
		fmt.Printf("\n")
	}
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

func check_winner(bd *Board, x int, y int) (winner int, end_of_game bool) {

	relevant_quads := indexed_winning_quads[x][y]
	for _, quad := range relevant_quads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return sum/4, true
		}
	}

	relevant_triplets := indexed_losing_triplets[x][y]
	for _, triplet := range relevant_triplets {
		sum := bd[triplet[0][0]][triplet[0][1]]
		sum += bd[triplet[1][0]][triplet[1][1]]
		sum += bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			return -sum/3, true
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

var scores [][]int = [][]int{
	[]int{2, 2, 2, 2, 2},
	[]int{2, 1, 1, 1, 2},
	[]int{2, 1, 0, 1, 2},
	[]int{2, 1, 1, 1, 2},
	[]int{2, 2, 2, 2, 2},
}

func static_value(bd *Board) (score int) {
	for i, row := range bd {
		for j, _ := range row {
			score += bd[i][j] * scores[i][j]
		}
	}
	for _, quad := range winning_quads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 3 || sum == -3 {
			score += sum * 10
		}
	}
	return score
}

func read_move(bd *Board) (x, y int) {
	for {
		fmt.Printf("Your move: ")
		_, err := fmt.Scanf("%d %d\n", &x, &y)
		if err == io.EOF {
			os.Exit(0)
		}
		if err != nil {
			fmt.Printf("Failed to read: %v\n", err)
			os.Exit(1)
		}
		if bd[x][y] == 0 {
			break
		}
		fmt.Printf("Cell (%d, %d) already occupied, try again\n", x, y)
	}
	return x, y
}
