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

func DefineBuiltinAccessorFunction(
	runtime *Runtime,
	obj ObjectInterface,
	name string,
	getBehaviour NativeFunctionBehaviour,
	setBehaviour NativeFunctionBehaviour,
	descriptor *AccessorPropertyDescriptor,
) {
	functionName := NewStringValue(name)
	getFunctionObject := CreateBuiltinFunction(runtime, getBehaviour, 0, functionName, nil, nil)
	setFunctionObject := CreateBuiltinFunction(runtime, setBehaviour, 1, functionName, nil, nil)

	descriptor.Get = getFunctionObject
	descriptor.Set = setFunctionObject
	obj.DefineOwnProperty(functionName, descriptor)
}
