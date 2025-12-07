package main

import (
	"fmt"
	"syscall/js"

	"zbrannelly.dev/go-js/pkg/lib-js/runtime"
)

var (
	rt    *runtime.Runtime
	realm *runtime.Realm
)

func evalJS(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return map[string]any{
			"error": "evalJS requires a code string argument",
		}
	}

	code := args[0].String()

	script, err := runtime.ParseScript(code, realm)
	if err != nil {
		return map[string]any{
			"error": fmt.Sprintf("%v", err),
		}
	}

	result := script.Evaluate(rt)
	if result.Type == runtime.Throw {
		return map[string]any{
			"error": fmt.Sprintf("%v", result.Value),
		}
	}

	if value, ok := result.Value.(*runtime.JavaScriptValue); ok {
		valueString, err := value.ToString()
		if err != nil {
			return map[string]any{
				"error": fmt.Sprintf("%v", err),
			}
		}
		return map[string]any{
			"result": valueString,
		}
	}

	return map[string]any{
		"result": fmt.Sprintf("%v", result.Value),
	}
}

func main() {
	rt = runtime.NewRuntime()
	realm = runtime.NewRealm(rt)

	// Expose JS runtime functions
	js.Global().Set("evalJS", js.FuncOf(evalJS))

	fmt.Println("Go JS Runtime initialized. Available functions: evalJS(code)")

	// Prevent the Go program from exiting so JS callbacks keep working
	select {}
}
