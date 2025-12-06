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
		fmt.Printf("Error parsing script: %v\n", err)
		os.Exit(1)
	}

	// Evaluate the script.
	result := script.Evaluate(rt)

	// Handle unhandled errors.
	if result.Type == runtime.Throw {
		fmt.Println(result.Value)
		os.Exit(1)
	}

	// Print string representation of the result.
	if value, ok := result.Value.(*runtime.JavaScriptValue); ok {
		// Converting to a string may throw an error.
		// e.g. a reference to a non-existent property.
		valueString, err := value.ToString()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(valueString)
	}
}
