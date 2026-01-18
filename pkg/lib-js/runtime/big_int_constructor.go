package runtime

func NewBigIntConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		BigIntConstructor,
		1,
		NewStringValue("BigInt"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// BigInt.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicBigIntPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return constructor
}

func BigIntConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		arguments = append(arguments, NewStringValue("0"))
	}

	if newTarget != nil && newTarget.Type != TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, "BigInt is not a constructor"))
	}

	completion := ToPrimitiveWithPreferredType(runtime, arguments[0], PreferredTypeNumber)
	if completion.Type != Normal {
		return completion
	}

	primValue := completion.Value.(*JavaScriptValue)

	if primValue.Type == TypeNumber {
		return NumberToBigInt(runtime, primValue.Value.(*Number))
	}

	return ToBigInt(runtime, primValue)
}
