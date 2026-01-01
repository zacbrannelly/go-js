package runtime

func NewSymbolConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		SymbolConstructor,
		1,
		NewStringValue("Symbol"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// TODO: Define other properties.

	return constructor
}

func SymbolConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if newTarget != nil && newTarget.Type != TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, "Symbol is not a constructor"))
	}

	if len(arguments) == 0 {
		arguments = append(arguments, NewUndefinedValue())
	}

	description := arguments[0]

	if description.Type == TypeUndefined {
		return NewNormalCompletion(NewSymbolValue(""))
	}

	completion := ToString(runtime, description)
	if completion.Type != Normal {
		return completion
	}

	description = completion.Value.(*JavaScriptValue)
	descriptionString := description.Value.(*String).Value
	return NewNormalCompletion(NewSymbolValue(descriptionString))
}
