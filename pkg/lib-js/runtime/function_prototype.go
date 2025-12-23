package runtime

func NewFunctionPrototype(runtime *Runtime) ObjectInterface {
	realm := runtime.GetRunningRealm()
	prototype := CreateBuiltinFunction(
		runtime,
		FunctionPrototypeConstructor,
		0,
		NewStringValue(""),
		realm,
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype),
	)

	return prototype
}

func DefineFunctionPrototypeProperties(runtime *Runtime, functionProto *FunctionObject) {
	DefineBuiltinFunction(runtime, functionProto, "call", FunctionPrototypeCall, 1)

	// TODO: Define other properties.
}

func FunctionPrototypeConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	return NewNormalCompletion(NewUndefinedValue())
}

func FunctionPrototypeCall(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = append(arguments, NewUndefinedValue())
	}

	callThisArg := arguments[0]
	callArguments := arguments[1:]

	if functionObj, ok := thisArg.Value.(*FunctionObject); ok {
		PrepareForTailCall()
		return functionObj.Call(runtime, callThisArg, callArguments)
	}

	return NewThrowCompletion(NewTypeError("'this' is not callable"))
}
