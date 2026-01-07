package runtime

func NewArrayBufferConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ArrayBufferConstructor,
		1,
		NewStringValue("ArrayBuffer"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// ArrayBuffer.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicArrayBufferPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// TODO: Define other properties.

	return constructor
}

func ArrayBufferConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	for idx := range 2 {
		if idx >= len(arguments) {
			arguments = append(arguments, NewUndefinedValue())
		}
	}

	if newTarget == nil || newTarget.Type == TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, "ArrayBuffer constructor requires 'new'"))
	}

	newTargetObj := newTarget.Value.(FunctionInterface)

	length := arguments[0]
	options := arguments[1]

	completion := ToIndex(runtime, length)
	if completion.Type != Normal {
		return completion
	}

	byteLength := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	completion = GetArrayBufferMaxByteLengthOption(runtime, options)
	if completion.Type != Normal {
		return completion
	}

	if completion.Value != nil {
		panic("TODO: Implement maxByteLength option.")
	}

	return AllocateArrayBuffer(runtime, newTargetObj, uint(byteLength))
}

func GetArrayBufferMaxByteLengthOption(runtime *Runtime, options *JavaScriptValue) *Completion {
	if options.Type != TypeObject {
		// nil to signal empty.
		return NewNormalCompletion(nil)
	}

	optionsObj := options.Value.(ObjectInterface)

	completion := optionsObj.Get(runtime, NewStringValue("maxByteLength"), options)
	if completion.Type != Normal {
		return completion
	}

	maxByteLengthVal := completion.Value.(*JavaScriptValue)

	if maxByteLengthVal.Type == TypeUndefined {
		// nil to signal empty.
		return NewNormalCompletion(nil)
	}

	return ToIndex(runtime, maxByteLengthVal)
}
