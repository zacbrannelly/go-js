package runtime

func NewArrayIteratorPrototype(runtime *Runtime) ObjectInterface {
	return OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicIteratorPrototype))
}

func DefineArrayIteratorPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// ArrayIterator.prototype.next
	DefineBuiltinFunction(runtime, prototype, "next", ArrayIteratorPrototypeNext, 0)

	// %Symbol.toStringTag%
	completion := prototype.DefineOwnProperty(runtime, runtime.SymbolToStringTag, &DataPropertyDescriptor{
		Value:        NewStringValue("Array Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: DefineOwnProperty threw an unexpected error in ArrayIterator.prototype constructor.")
	}

	// TODO: Define properties.
}

func ArrayIteratorPrototypeNext(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	return GeneratorResume(runtime, thisArg.Value.(*Object), nil, "%ArrayIteratorPrototype%")
}
