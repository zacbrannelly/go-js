package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
	"zbrannelly.dev/go-js/pkg/lib-js/runtime"
)

type Mode int

const (
	ModeLexer Mode = iota
	ModeParser
	ModeRuntime
)

func (m Mode) String() string {
	switch m {
	case ModeLexer:
		return "lexer"
	case ModeParser:
		return "parser"
	case ModeRuntime:
		return "runtime"
	default:
		return "unknown"
	}
}

func ParseMode(s string) (Mode, error) {
	switch strings.ToLower(s) {
	case "lexer":
		return ModeLexer, nil
	case "parser":
		return ModeParser, nil
	case "runtime":
		return ModeRuntime, nil
	default:
		return ModeLexer, fmt.Errorf("invalid mode: %s", s)
	}
}

var (
	modeStr string

	replCmd = &cobra.Command{
		Use:   "repl",
		Short: "Start a REPL for the JavaScript engine",
		Run: func(cmd *cobra.Command, args []string) {
			mode, err := ParseMode(modeStr)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			switch mode {
			case ModeLexer:
				lexerREPL()
			case ModeParser:
				parserREPL()
			case ModeRuntime:
				runtimeREPL()
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(replCmd)

	replCmd.Flags().StringVarP(&modeStr, "mode", "m", "runtime", "The mode to run the REPL in: lexer, parser, runtime")
}

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
			// Converting to a string may throw an error.
			// For example, a reference to a non-existent property.
			valueString, err := result.Value.(*runtime.JavaScriptValue).ToString()
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println(valueString)
		}
	}
}
