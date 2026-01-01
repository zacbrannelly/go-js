package runtime

func NewStringPrototype(runtime *Runtime) ObjectInterface {
	prototype := &StringObject{
		Prototype:        runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype),
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		StringData:       NewStringValue(""),
		Extensible:       true,
	}

	DefinePropertyOrThrow(runtime, prototype, lengthStr, &DataPropertyDescriptor{
		Value:        NewNumberValue(0, false),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return prototype
}

func DefineStringPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// TODO: Define other properties.
}
