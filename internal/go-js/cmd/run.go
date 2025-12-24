package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"zbrannelly.dev/go-js/pkg/lib-js/runtime"
)

var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "Run a JavaScript file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runFile(args[0])
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
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

	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}

	// Evaluate the script.
	result := script.Evaluate(rt)

	if result.Type == runtime.Throw {
		jsError, ok := result.Value.(*runtime.JavaScriptValue)
		if !ok {
			panic("Assert failed: Expected a JavaScript value for the thrown error.")
		}
		fmt.Println("Uncaught " + runtime.ErrorToString(rt, jsError))
		os.Exit(1)
	} else if result.Value != nil {
		// Converting to a string may throw an error.
		// For example, a reference to a non-existent property.
		valueString, err := result.Value.(*runtime.JavaScriptValue).ToString(rt)
		if err != nil {
			fmt.Println("Uncaught " + runtime.ErrorToString(rt, err))
		} else {
			fmt.Println(valueString)
		}
	}
}
