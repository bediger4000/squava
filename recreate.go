package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func main() {

	markers := []rune{'X', 'O'}

	board := make([][]rune, 5)
	for i := 0; i < 5; i++ {
		board[i] = []rune{'_', '_', '_', '_', '_'}
	}

	buf, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var moves [][]byte
	lines := bytes.Count(buf, []byte{'\n'})
	if lines > 1 {
		// assume 1 move per line
		moves = bytes.Split(buf, []byte{'\n'})
	} else {
		// assume all moves on one line
		moves = bytes.Split(buf, []byte{' '})
	}

	moveCounter := 0

	for _, move := range moves {
		moveCounter++

		if moveCounter > 25 {
			break
		}

		n, m, useIt := finagleMove(move, moveCounter)
		if !useIt {
			continue
		}

		fmt.Printf("%c move %d,%d\n", markers[moveCounter%2], n, m)
		board[n][m] = markers[moveCounter%2]

		for i := 0; i < 5; i++ {
			for _, marker := range board[i] {
				fmt.Printf("%c ", marker)
			}
			fmt.Println()
		}
		fmt.Println()
		_, err := fmt.Scanf("\n")
		if err != nil {
			log.Print(err)
		}
	}
}

func finagleMove(move []byte, moveCounter int) (int, int, bool) {

	fields := bytes.Split(move, []byte{','})

	if len(fields) != 2 {
		fmt.Fprintf(os.Stderr, "Move %d, %q, problem\n", moveCounter, string(move))
		return 0, 0, false
	}

	if len(fields[0]) == 2 {
		fields[0] = fields[0][0:1]
	}
	if len(fields[1]) == 2 {
		fields[1] = fields[1][0:1]
	}

	if fields[0][0] < '0' || fields[0][0] > '4' {
		fmt.Fprintf(os.Stderr, "Move %d, %q, problem with 1st field\n", moveCounter, string(move))
		return 0, 0, false
	}

	n := int(fields[0][0] - '0')

	if fields[1][0] < '0' || fields[1][0] > '4' {
		fmt.Fprintf(os.Stderr, "Move %d, %q, problem with 2nd field\n", moveCounter, string(move))
		return 0, 0, false
	}

	m := int(fields[1][0] - '0')

	return n, m, true
}
