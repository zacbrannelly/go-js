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
