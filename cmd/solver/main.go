package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"

	"overberne.com/worlde/assets"
	"overberne.com/worlde/pkg/wordle"
)

const usage = `usage:

help           prints this message
solve          run the solver
play   [word]  starts a game, optionally pass a word`

func main() {
	printUsage()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		line := strings.ToLower(scanner.Text())
		segments := strings.Split(line, " ")
		command := segments[0]

		switch command {
		case "help":
			printUsage()
		case "solve":
			solve(scanner)
		case "play":
			if len(segments) == 1 {
				play(scanner, "")
			} else {
				play(scanner, segments[1])
			}
		}
	}
}

func printUsage() {
	fmt.Println(usage)
	fmt.Println()
}

func solve(scanner *bufio.Scanner) {
	fmt.Println("Initializing new solver...")
	fmt.Println("Guess 'q', 'quit', or 'exit' to quit the solver")
	solver := wordle.NewSolver()
	fmt.Println()
	fmt.Println(solver)
	fmt.Println()

	for {
		fmt.Print("Guess       : ")
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		word := strings.ToLower(scanner.Text())

		if word == "quit" || word == "exit" || word == "q" {
			return
		}

		fmt.Print("Hint (g/y/ ): ")
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		hintsString := scanner.Text()

		hints := [5]wordle.Hint{}
		for i := range word {
			if i < len(hintsString) && hintsString[i] == 'g' {
				hints[i] = wordle.LetterCorrect
			} else if i < len(hintsString) && hintsString[i] == 'y' {
				hints[i] = wordle.LetterTransposed
			} else {
				hints[i] = wordle.LetterWrong
			}
		}

		if done, err := solver.AddGuess(word, hints); err != nil {
			fmt.Println(err)
		} else if done {
			fmt.Println()
			fmt.Println(solver)
			fmt.Println()
			break
		} else {
			fmt.Println()
			fmt.Println(solver)
			fmt.Println()
		}
	}

	fmt.Println("Solver finished!")
}

func play(scanner *bufio.Scanner, word string) {
	if word == "" {
		word = assets.Wordles[rand.Int()%len(assets.Wordles)]
	}
	game := wordle.NewGame(word)
	fmt.Println(game)

	for numGuesses := 0; numGuesses < wordle.MaxGuesses; {
		fmt.Print("Guess: ")
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		guess := strings.ToLower(scanner.Text())

		if guess == "quit" || guess == "exit" || guess == "q" {
			break
		}

		if done, err := game.Guess(guess); err != nil {
			fmt.Println(err)
		} else if done {
			fmt.Println(game)
			fmt.Println("You won!")
			return
		} else {
			numGuesses++
			fmt.Println(game)
		}
	}

	fmt.Printf("The word was: %s\n\n", word)
}
