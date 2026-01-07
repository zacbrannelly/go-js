package runtime

func NewConcreteTypedArrayPrototype(runtime *Runtime, typedArrayName TypedArrayName) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicTypedArrayPrototype))

	elementSize, ok := TypedArrayElementSizes[typedArrayName]
	if !ok {
		panic("Assert failed: Provided typedArrayName is not mapped in TypedArrayElementSizes.")
	}

	// BYTES_PER_ELEMENT
	prototype.DefineOwnProperty(runtime, NewStringValue("BYTES_PER_ELEMENT"), &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(elementSize), false),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return prototype
}

func NewTypedArrayPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
	return prototype
}

func DefineTypedArrayPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// TypedArray.prototype.length
	DefineBuiltinAccessorFunction(runtime, prototype, "length", TypedArrayPrototypeLengthGetter, nil, &AccessorPropertyDescriptor{
		Enumerable:   false,
		Configurable: true,
	})

	// TODO: Define other properties.
}

func TypedArrayPrototypeLengthGetter(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	typedArrayObj, ok := thisArg.Value.(*TypedArrayObject)
	if !ok {
		return NewThrowCompletion(NewTypeError(runtime, "This is not a TypedArray object."))
	}

	taRecord := MakeTypedArrayWithBufferWitness(typedArrayObj, false)
	if taRecord.IsTypedArrayOutOfBounds() {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	length := taRecord.TypedArrayLength()
	return NewNormalCompletion(NewNumberValue(float64(length), false))
}
