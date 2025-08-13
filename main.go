package main

import (
	"bufio"
	"fmt"
	"os"

	"zbrannelly.dev/go-js/cmd/lexer"
)

func main() {
	fmt.Println("Welcome to the JavaScript lexer REPL!")
	fmt.Println("Select lexer goal:")
	fmt.Println("1) InputElementDiv")
	fmt.Println("2) InputElementRegExp")

	scanner := bufio.NewScanner(os.Stdin)

	var goal lexer.LexicalGoal
	for {
		fmt.Print("Enter choice (1 or 2): ")
		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			goal = lexer.InputElementDiv
		case "2":
			goal = lexer.InputElementRegExp
		default:
			fmt.Println("Invalid choice, please enter 1 or 2")
			continue
		}
		break
	}

	fmt.Println("\nEnter JavaScript code to lex (press Ctrl+D to exit):")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		tokens := lexer.Lex(input, goal)
		for _, token := range tokens {
			fmt.Printf("%d: %s\n", token.Type, token.Value)
		}
	}
}
