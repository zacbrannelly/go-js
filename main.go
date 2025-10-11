package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"zbrannelly.dev/go-js/cmd/lexer"
	"zbrannelly.dev/go-js/cmd/parser"
	"zbrannelly.dev/go-js/cmd/parser/ast"
	"zbrannelly.dev/go-js/cmd/runtime"
)

func lexerREPL() {
	fmt.Println("Welcome to the JavaScript lexer REPL!")
	fmt.Println("Select lexer goal:")
	fmt.Println("1) InputElementDiv")
	fmt.Println("2) InputElementRegExp")
	fmt.Println("3) InputElementRegExpOrTemplateTail")
	fmt.Println("4) InputElementHashbangOrRegExp")
	fmt.Println("5) InputElementTemplateTail")

	scanner := bufio.NewScanner(os.Stdin)

	var goal lexer.LexicalGoal
	for {
		fmt.Print("Enter choice: ")
		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			goal = lexer.InputElementDiv
		case "2":
			goal = lexer.InputElementRegExp
		case "3":
			goal = lexer.InputElementRegExpOrTemplateTail
		case "4":
			goal = lexer.InputElementHashbangOrRegExp
		case "5":
			goal = lexer.InputElementTemplateTail
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

		tokens := lexer.LexAll(input, goal)
		for _, token := range tokens {
			fmt.Printf("%d: %s\n", token.Type, token.Value)
		}
	}
}

func parserREPL() {
	fmt.Println("Welcome to the JavaScript parser REPL!")
	fmt.Println("Select parser goal:")
	fmt.Println("1) Script")

	scanner := bufio.NewScanner(os.Stdin)

	var goal ast.NodeType
	for {
		fmt.Print("Enter choice: ")
		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			goal = ast.Script
		default:
			fmt.Println("Invalid choice")
			continue
		}
		break
	}

	fmt.Println("\nEnter JavaScript code to parse (press Ctrl+D to exit):")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		scriptNode, err := parser.ParseText(input, goal)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		var traverse func(node ast.Node, depth int)
		traverse = func(node ast.Node, depth int) {
			indent := strings.Repeat("  ", depth)
			fmt.Printf("%s%s\n", indent, node.ToString())

			// The node's ToString method will handle it's children.
			if !node.IsComposable() {
				return
			}

			for _, child := range node.GetChildren() {
				traverse(child, depth+1)
			}
		}

		traverse(scriptNode, 0)
		fmt.Println()
	}
}

func runtimeREPL() {
	fmt.Println("Welcome to the JavaScript runtime REPL!")
	fmt.Println("Enter JavaScript code to evaluate (press Ctrl+D to exit):")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		realm := runtime.NewRealm()
		script, err := runtime.ParseScript(input, realm)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		rt := &runtime.Runtime{}
		result := script.Evaluate(rt)

		if result.Type == runtime.Throw {
			fmt.Println(result.Value)
			continue
		}

		if result.Value != nil {
			fmt.Println(result.Value.(*runtime.JavaScriptValue).ToString())
		}
	}
}

func main() {
	fmt.Println("Welcome to the JavaScript REPL!")
	fmt.Println("Select mode:")
	fmt.Println("1) Lexer")
	fmt.Println("2) Parser")
	fmt.Println("3) Runtime")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Enter choice: ")
		if !scanner.Scan() {
			return
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			lexerREPL()
		case "2":
			parserREPL()
		case "3":
			runtimeREPL()
		default:
			fmt.Println("Invalid choice")
		}
	}
}
