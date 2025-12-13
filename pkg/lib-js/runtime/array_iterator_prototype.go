package runtime

func NewArrayIteratorPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().Intrinsics[IntrinsicIteratorPrototype])

	// TODO: Define properties.

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
