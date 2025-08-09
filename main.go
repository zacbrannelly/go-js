package main

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/lexer"
)

func main() {
	tokens := lexer.Lex(" ", lexer.InputElementDiv)
	for _, token := range tokens {
		fmt.Printf("%d: %s\n", token.Type, token.Value)
	}
}
