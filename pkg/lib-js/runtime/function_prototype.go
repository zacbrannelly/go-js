package runtime

func NewFunctionPrototype(runtime *Runtime) ObjectInterface {
	realm := runtime.GetRunningRealm()
	prototype := CreateBuiltinFunction(
		runtime,
		FunctionPrototypeCall,
		0,
		NewStringValue(""),
		realm,
		realm.Intrinsics[IntrinsicObjectPrototype],
	)

	return prototype
}

func FunctionPrototypeCall(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	return NewNormalCompletion(NewUndefinedValue())
}
