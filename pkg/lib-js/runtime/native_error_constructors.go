package runtime

func NewNativeErrorConstructor(
	runtime *Runtime,
	errorType NativeErrorType,
	proto Intrinsic,
) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		NativeErrorConstructorWrapper(proto),
		1,
		NewStringValue(string(errorType)),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// NativeError.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(proto)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return constructor
}

func NativeErrorConstructorWrapper(
	prototype Intrinsic,
) NativeFunctionBehaviour {
	return func(
		runtime *Runtime,
		function *FunctionObject,
		thisArg *JavaScriptValue,
		arguments []*JavaScriptValue,
		newTarget *JavaScriptValue,
	) *Completion {
		return NativeErrorConstructor(
			runtime,
			function,
			thisArg,
			arguments,
			newTarget,
			prototype,
		)
	}
}

func NativeErrorConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
	prototype Intrinsic,
) *Completion {
	for idx := range 2 {
		if idx >= len(arguments) {
			arguments = append(arguments, NewUndefinedValue())
		}
	}

	if newTarget == nil || newTarget.Type == TypeUndefined {
		newTarget = NewJavaScriptValue(TypeObject, function)
	}

	newTargetObj := newTarget.Value.(FunctionInterface)

	completion := OrdinaryCreateFromConstructor(runtime, newTargetObj, prototype)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(*Object)

	// Set [[ErrorData]] internal slot.
	object.IsError = true

	messageVal := arguments[0]

	if messageVal.Type != TypeUndefined {
		completion = ToString(runtime, messageVal)
		if completion.Type != Normal {
			return completion
		}

		messageVal = completion.Value.(*JavaScriptValue)

		completion = object.DefineOwnProperty(runtime, messageStr, &DataPropertyDescriptor{
			Value:        messageVal,
			Writable:     true,
			Enumerable:   false,
			Configurable: true,
		})
		if completion.Type != Normal {
			panic("Assert failed: Failed to define 'message' property on Error object.")
		}
		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: Failed to define 'message' property on Error object.")
		}
	}

	messageOptions := arguments[1]
	completion = InstallErrorCause(runtime, object, messageOptions)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(objectVal)
}

func NewSyntaxError(runtime *Runtime, message string) *JavaScriptValue {
	return NewNativeError(runtime, IntrinsicSyntaxErrorConstructor, message)
}

func NewTypeError(runtime *Runtime, message string) *JavaScriptValue {
	return NewNativeError(runtime, IntrinsicTypeErrorConstructor, message)
}

func NewReferenceError(runtime *Runtime, message string) *JavaScriptValue {
	return NewNativeError(runtime, IntrinsicReferenceErrorConstructor, message)
}

func NewRangeError(runtime *Runtime, message string) *JavaScriptValue {
	return NewNativeError(runtime, IntrinsicRangeErrorConstructor, message)
}

func NewEvalError(runtime *Runtime, message string) *JavaScriptValue {
	return NewNativeError(runtime, IntrinsicEvalErrorConstructor, message)
}

func NewURIError(runtime *Runtime, message string) *JavaScriptValue {
	return NewNativeError(runtime, IntrinsicURIErrorConstructor, message)
}

func NewNativeError(runtime *Runtime, errorConstructor Intrinsic, message string) *JavaScriptValue {
	realm := runtime.GetRunningRealm()
	constructor := realm.GetIntrinsic(errorConstructor).(FunctionInterface)
	completion := constructor.Construct(runtime, []*JavaScriptValue{NewStringValue(message)}, nil)
	if completion.Type != Normal {
		panic("Assert failed: Failed to construct NativeError.")
	}
	return completion.Value.(*JavaScriptValue)
}
