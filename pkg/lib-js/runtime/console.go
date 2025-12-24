package runtime

import (
	"fmt"
	"strings"
)

func NewConsoleObject(runtime *Runtime) ObjectInterface {
	realm := runtime.GetRunningRealm()
	proto := OrdinaryObjectCreate(realm.GetIntrinsic(IntrinsicObjectPrototype))
	console := OrdinaryObjectCreate(proto)

	// console[Symbol.toStringTag] = "Console"
	console.DefineOwnProperty(runtime, runtime.SymbolToStringTag, &DataPropertyDescriptor{
		Value:        NewStringValue("Console"),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// console.log
	DefineBuiltinFunction(runtime, console, "log", ConsoleLog, 0)

	// TODO: Define properties.

	return console
}

func ConsoleLog(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		fmt.Println()
		return NewNormalCompletion(NewUndefinedValue())
	}

	messageVal := arguments[0]
	completion := ToString(messageVal)
	if completion.Type != Normal {
		return completion
	}

	message := completion.Value.(*JavaScriptValue).Value.(*String).Value

	// Check if % is in the message.
	if strings.Contains(message, "%") {
		// TODO: Implement % formatting.
		panic("TODO: Implement % formatting.")
	}

	messages := []string{message}

	if len(arguments) > 1 {
		for i := 1; i < len(arguments); i++ {
			argument := arguments[i]
			completion := ToString(argument)
			if completion.Type != Normal {
				return completion
			}

			argumentStr := completion.Value.(*JavaScriptValue).Value.(*String).Value
			messages = append(messages, argumentStr)
		}
	}

	fmt.Println(strings.Join(messages, " "))
	return NewNormalCompletion(NewUndefinedValue())
}
