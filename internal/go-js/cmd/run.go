package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"zbrannelly.dev/go-js/pkg/lib-js/parser"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
	"zbrannelly.dev/go-js/pkg/lib-js/runtime"
)

var (
	runCmd = &cobra.Command{
		Use:   "run [file]",
		Short: "Run a JavaScript file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]

			mode, err := ParseMode(modeStr)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			switch mode {
			case ModeLexer:
				panic("Lexer mode not supported")
			case ModeParser:
				parseFile(path)
			case ModeRuntime:
				runFile(path)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&modeStr, "mode", "m", "runtime", "The mode to run the script in: parser, runtime")
}

func parseFile(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	scriptNode, err := parser.ParseText(string(content), ast.Script)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
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

func runFile(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse the script.
	rt := runtime.NewRuntime()
	realm := runtime.NewRealm(rt)
	script, err := runtime.ParseScript(string(content), realm)

	// TODO: Make ParseScript return SyntaxError objects (NativeError objects).
	if err != nil {
		fmt.Printf("SyntaxError: %v\n", err)
		os.Exit(1)
	}

	// Evaluate the script.
	result := script.Evaluate(rt)

	if result.Type == runtime.Throw {
		jsError, ok := result.Value.(*runtime.JavaScriptValue)
		if !ok {
			panic("Assert failed: Expected a JavaScript value for the thrown error.")
		}
		fmt.Println(runtime.ErrorToString(rt, jsError))
		os.Exit(1)
	} else if result.Value != nil {
		// Converting to a string may throw an error.
		// For example, a reference to a non-existent property.
		valueString, err := result.Value.(*runtime.JavaScriptValue).ToString(rt)
		if err != nil {
			fmt.Println(runtime.ErrorToString(rt, err))
		} else if valueString != "undefined" {
			fmt.Println(valueString)
		}
	}
}
