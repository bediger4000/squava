package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	markers := []rune{'X', 'O'}

	board := make([][]rune, 5)
	for i := 0; i < 5; i++ {
		board[i] = []rune{'_', '_', '_', '_', '_'}
	}

	fin, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	defer fin.Close()

	scanner := bufio.NewScanner(fin)

	lineCounter := 0

	for scanner.Scan() {
		lineCounter++
		line := scanner.Text()

		fields := strings.Split(line, ",")
		if len(fields) != 2 {
			fmt.Fprintf(os.Stderr, "Line %d, %q, problem\n", lineCounter, line)
			continue
		}
		if len(fields[0]) == 2 {
			fields[0] = fields[0][0:1]
		}
		if len(fields[1]) == 2 {
			fields[1] = fields[1][0:1]
		}
		n, nerr := strconv.Atoi(fields[0])
		if nerr != nil {
			fmt.Fprintf(os.Stderr, "Line %d, %q, problem with 1st field %v\n", lineCounter, line, nerr)
			continue
		}
		m, merr := strconv.Atoi(fields[1])
		if merr != nil {
			fmt.Fprintf(os.Stderr, "Line %d, %q, problem with 2nd field %v\n", lineCounter, line, merr)
			continue
		}
		fmt.Printf("%c move %d,%d\n", markers[lineCounter%2], n, m)
		board[n][m] = markers[lineCounter%2]

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

	if err := scanner.Err(); err != nil {
		log.Fatalf("problem line %d: %v", lineCounter, err)
	}
}
