package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
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

type ParserGoal int

const (
	ParserGoalScript ParserGoal = iota
)

func (g ParserGoal) String() string {
	switch g {
	case ParserGoalScript:
		return "Script"
	default:
		return "unknown"
	}
}

func ParseParserGoal(s string) (ParserGoal, error) {
	switch strings.ToLower(s) {
	case "script":
		return ParserGoalScript, nil
	default:
		return ParserGoalScript, fmt.Errorf("invalid parser goal: %s", s)
	}
}

var (
	modeStr       string
	lexerGoalStr  string
	parserGoalStr string
	isolated      bool

	replCmd = &cobra.Command{
		Use:   "repl",
		Short: "REPL for the go-js JavaScript Engine",
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
	replCmd.Flags().StringVarP(&lexerGoalStr, "lexer-goal", "g", "InputElementDiv", "The lexer goal to run the REPL in")
	replCmd.Flags().StringVarP(&parserGoalStr, "parser-goal", "p", "Script", "The parser goal to run the REPL in: Script")
	replCmd.Flags().BoolVarP(&isolated, "isolated", "i", false, "If enabled, each expression will be evaluated in an isolated realm.")
}

func lexerREPL() {
	fmt.Println("go-js lexer REPL (Ctrl+D to exit)")

	goal, err := lexer.ParseLexicalGoal(lexerGoalStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil {
			if err == io.EOF || err == readline.ErrInterrupt {
				break
			}
			fmt.Printf("Error: %v\n", err)
			continue
		}

		tokens := lexer.LexAll(input, goal)
		for _, token := range tokens {
			fmt.Printf("%s: %s\n", token.Type.String(), token.Value)
		}
	}
}

func parserREPL() {
	fmt.Println("go-js parser REPL (Ctrl+D to exit)")

	selectedGoal, err := ParseParserGoal(parserGoalStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var goal ast.NodeType
	switch selectedGoal {
	case ParserGoalScript:
		goal = ast.Script
	default:
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil {
			if err == io.EOF || err == readline.ErrInterrupt {
				break
			}
			fmt.Printf("Error: %v\n", err)
			continue
		}

		scriptNode, err := parser.ParseText(input, goal)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Recursively traverse the AST and print the nodes.
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
	if isolated {
		fmt.Println("go-js runtime REPL (isolated mode) (Ctrl+D to exit)")
	} else {
		fmt.Println("go-js runtime REPL (Ctrl+D to exit)")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	rt := runtime.NewRuntime()
	realm := runtime.NewRealm(rt)

	for {
		input, err := rl.Readline()
		if err != nil {
			if err == io.EOF || err == readline.ErrInterrupt {
				break
			}
			fmt.Printf("Error: %v\n", err)
			continue
		}

		script, err := runtime.ParseScript(input, realm)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		result := script.Evaluate(rt)

		if result.Type == runtime.Throw {
			// Handle errors thrown by the script (might be a JavaScript value).
			// TODO: This might always need to be a JavaScript value.
			if jsError, ok := result.Value.(*runtime.JavaScriptValue); ok {
				valueString, err := jsError.ToString(rt)
				if err != nil {
					fmt.Println("Uncaught " + err.Error())
				} else {
					fmt.Println("Uncaught " + valueString)
				}
			} else {
				// Otherwise, print the error as a string.
				fmt.Println("Uncaught " + result.Value.(error).Error())
			}
		} else if result.Value != nil {
			// Converting to a string may throw an error.
			// For example, a reference to a non-existent property.
			valueString, err := result.Value.(*runtime.JavaScriptValue).ToString(rt)
			if err != nil {
				fmt.Println("Uncaught " + err.Error())
			} else {
				fmt.Println(valueString)
			}
		}

		// Reset the realm and runtime if the isolated flag is enabled.
		if isolated {
			rt = runtime.NewRuntime()
			realm = runtime.NewRealm(rt)
		}
	}
}
