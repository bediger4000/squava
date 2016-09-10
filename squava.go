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

const WIN = 10000
const LOSS = -10000

var max_depth int = 10

// Arrays of losing triplets and winning quads, indexed
// by <x,y> coords of all pairs composing each of the quads
// or triplets. Makes check_winner() a lot more efficient.
var indexed_losing_triplets [5][5][][][]int
var indexed_winning_quads [5][5][][][]int

var move_counter int = 0

func main() {

	human_first_ptr := flag.Bool("H", true, "Human takes first move")
	computer_first_ptr := flag.Bool("C", false, "Computer takes first move")
	max_depth_ptr := flag.Int("d", 10, "maximum lookahead depth")
	flag.Parse()

	// Set up for use by staticValue()
	for _, triplet := range losing_triplets {
		for _, pair := range triplet {
			indexed_losing_triplets[pair[0]][pair[1]] = append(indexed_losing_triplets[pair[0]][pair[1]], triplet)
		}
	}
	for _, quad := range winning_quads {
		for _, pair := range quad {
			indexed_winning_quads[pair[0]][pair[1]] = append(
				indexed_winning_quads[pair[0]][pair[1]], quad)
		}
	}

	var human_first bool = *human_first_ptr
	if *computer_first_ptr {
		human_first = false
	}
	max_depth = *max_depth_ptr

	rand.Seed(time.Now().UTC().UnixNano())

	var bd Board
	var end_of_game bool = false

	move_counter := 0

	for !end_of_game {

		var l, m int
		if human_first {
			l, m = readMove(&bd)
			bd[l][m] = -1
			end_of_game, _ = staticValue(&bd, 0, -1, l, m)
			move_counter++
		}

		if end_of_game {
			break
		}

		human_first = true

		var moves [25][2]int
		var next int = 0

		max := LOSS

		if move_counter < 4 {
			max_depth = 6
		}
		if move_counter > 3 {
			max_depth = 8
		}
		if move_counter > 10 {
			max_depth = *max_depth_ptr
		}

		for i, row := range bd {
			for j, mark := range row {
				if mark == 0 {
					bd[i][j] = 1
					val := alphaBeta(&bd, 1, -1, LOSS, WIN, i, j)
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

		printBoard(&bd)

		end_of_game, _ = staticValue(&bd, 0, 1, moves[r][0], moves[r][1])
	}

	var phrase string
	switch findWinnder(&bd) {
	case 1:
		phrase = "\nX wins\n"
	case 0:
		phrase = "\nCat wins\n"
	case -1:
		phrase = "\nO wins\n"
	}
	fmt.Printf(phrase)

	printBoard(&bd)

	os.Exit(0)
}

func findWinnder(bd *Board) int {
	for _, quad := range winning_quads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return bd[quad[0][0]][quad[0][1]]
		}
	}

	for _, triplet := range losing_triplets {
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
var checkable_cells [9][2]int = [9][2]int{
	{0, 2}, {1, 2}, {2, 0},
	{2, 1}, {2, 2}, {2, 3},
	{2, 4}, {3, 2}, {4, 2},
}

func staticValue(bd *Board, ply int, player int, x, y int) (stop_recursing bool, value int) {

	relevant_quads := indexed_winning_quads[x][y]
	for _, quad := range relevant_quads {
		sum := bd[quad[0][0]][quad[0][1]]
		sum += bd[quad[1][0]][quad[1][1]]
		sum += bd[quad[2][0]][quad[2][1]]
		sum += bd[quad[3][0]][quad[3][1]]

		if sum == 4 || sum == -4 {
			return true, bd[quad[0][0]][quad[0][1]] * (WIN - ply)
		}
	}

	relevant_triplets := indexed_losing_triplets[x][y]
	for _, triplet := range relevant_triplets {
		sum := bd[triplet[0][0]][triplet[0][1]]
		sum += bd[triplet[1][0]][triplet[1][1]]
		sum += bd[triplet[2][0]][triplet[2][1]]

		if sum == 3 || sum == -3 {
			return true, -bd[triplet[0][0]][triplet[0][1]] * (WIN - ply)
		}
	}

	if ply == max_depth {

		for _, cell := range checkable_cells {
			relevant_quads := indexed_winning_quads[cell[0]][cell[1]]
			for _, quad := range relevant_quads {
				sum := bd[quad[0][0]][quad[0][1]]
				sum += bd[quad[1][0]][quad[1][1]]
				sum += bd[quad[2][0]][quad[2][1]]
				sum += bd[quad[3][0]][quad[3][1]]
				if sum == 3 || sum == -3 {
					value += sum * 10
				}
			}
		}

		// Try to stay out of 2-of-losing-3 triplets
		for _, cell := range checkable_cells {
			relevant_triplets := indexed_losing_triplets[cell[0]][cell[1]]
			for _, triplet := range relevant_triplets {
				sum := bd[triplet[0][0]][triplet[0][1]]
				sum += bd[triplet[1][0]][triplet[1][1]]
				sum += bd[triplet[2][0]][triplet[2][1]]
				if sum == 3 || sum == -3 {
					value += sum * 5
				}
			}
		}

		if value == 0 {
			// Bive it a slight bias for those early
			// moves when all losing-triplets and winning-quads
			// are beyond the horizon.
			for i, row := range bd {
				for j, _ := range row {
					value += bd[i][j] * scores[i][j]
				}
			}
		}

		return true, value
	}

	// Not at max depth of search, but the whole board
	// might be filled in.
	for _, row := range bd {
		for _, val := range row {
			if val == 0 {
				return false, 0
			}
		}
	}
	// Get here, all 25 spots on board filled, no winning quadruplet
	return true, 0
}

func alphaBeta(bd *Board, ply int, player int, alpha int, beta int, x int, y int) (value int) {

	stop_recursing, score := staticValue(bd, ply, player, x, y)

	if stop_recursing {
		return score
	}

	switch player {
	case 1:
		value = LOSS
		for i, row := range bd {
			for j, marker := range row {
				if marker == 0 {
					bd[i][j] = player
					n := alphaBeta(bd, ply+1, -player, alpha, beta, i, j)
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
					n := alphaBeta(bd, ply+1, -player, alpha, beta, i, j)
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

func printBoard(bd *Board) {
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

var scores [][]int = [][]int{
	[]int{3, 3, 0, 3, 3},
	[]int{3, 4, 1, 4, 3},
	[]int{0, 1, 0, 1, 2},
	[]int{3, 4, 1, 4, 3},
	[]int{3, 3, 0, 3, 3},
}

func readMove(bd *Board) (x, y int) {
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
		switch {
		case x < 0 || x > 4 || y < 0 || y > 4:
			fmt.Printf("Choose two numbers between 0 and 4, try again\n")
		case bd[x][y] == 0:
			break
		case bd[x][y] != 0:
			fmt.Printf("Cell (%d, %d) already occupied, try again\n", x, y)
		}
	}
	return x, y
}
