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
		os.Exit(1)
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
}
