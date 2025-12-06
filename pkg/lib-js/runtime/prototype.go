package runtime

func DefineBuiltinFunction(
	runtime *Runtime,
	obj ObjectInterface,
	name string,
	behaviour NativeFunctionBehaviour,
	length int,
) {
	functionName := NewStringValue(name)
	functionObject := CreateBuiltinFunction(runtime, behaviour, length, functionName, nil, nil)
	functionValue := NewJavaScriptValue(TypeObject, functionObject)

	obj.DefineOwnProperty(functionName, &DataPropertyDescriptor{
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
		Value:        functionValue,
	})
}
