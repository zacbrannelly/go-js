package main

import (
	"bufio"
	"fmt"
	"os"

	"zbrannelly.dev/go-js/cmd/lexer"
)

func main() {
	fmt.Println("Welcome to the JavaScript lexer REPL!")
	fmt.Println("Enter JavaScript code to lex (press Ctrl+D to exit):")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		tokens := lexer.Lex(input, lexer.InputElementDiv)
		for _, token := range tokens {
			fmt.Printf("%d: %s\n", token.Type, token.Value)
		}
	}
}
