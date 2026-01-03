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

	// Define well-known symbols.
	DefineWellKnownSymbols(runtime, constructor, "toStringTag", runtime.SymbolToStringTag)
	DefineWellKnownSymbols(runtime, constructor, "iterator", runtime.SymbolIterator)
	DefineWellKnownSymbols(runtime, constructor, "species", runtime.SymbolSpecies)
	DefineWellKnownSymbols(runtime, constructor, "unscopables", runtime.SymbolUnscopables)
	DefineWellKnownSymbols(runtime, constructor, "hasInstance", runtime.SymbolHasInstance)
	DefineWellKnownSymbols(runtime, constructor, "toPrimitive", runtime.SymbolToPrimitive)
	DefineWellKnownSymbols(runtime, constructor, "isConcatSpreadable", runtime.SymbolConcatSpreadable)

	// TODO: Define other properties.

	return constructor
}

func DefineWellKnownSymbols(runtime *Runtime, constructor *FunctionObject, name string, symbol *JavaScriptValue) {
	constructor.DefineOwnProperty(runtime, NewStringValue(name), &DataPropertyDescriptor{
		Value:        symbol,
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
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
