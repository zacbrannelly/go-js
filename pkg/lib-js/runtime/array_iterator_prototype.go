package runtime

func NewArrayIteratorPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().Intrinsics[IntrinsicIteratorPrototype])

	// TODO: Define properties.

	// ArrayIterator.prototype.next
	DefineBuiltinFunction(runtime, prototype, "next", ArrayIteratorPrototypeNext, 0)

	// %Symbol.toStringTag%
	completion := prototype.DefineOwnProperty(runtime.SymbolToStringTag, &DataPropertyDescriptor{
		Value:        NewStringValue("Array Iterator"),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: DefineOwnProperty threw an unexpected error in ArrayIterator.prototype constructor.")
	}

	return prototype
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
