package runtime

var (
	causeStr = NewStringValue("cause")
)

func NewErrorConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ErrorConstructor,
		1,
		NewStringValue("Error"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// Error.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicErrorPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return constructor
}

func ErrorConstructor(
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
		newTarget = NewJavaScriptValue(TypeObject, function)
	}

	newTargetObj := newTarget.Value.(*FunctionObject)

	completion := OrdinaryCreateFromConstructor(runtime, newTargetObj, IntrinsicErrorPrototype)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(*Object)

	// Set [[ErrorData]] internal slot.
	object.IsError = true

	messageVal := arguments[0]

	if messageVal.Type != TypeUndefined {
		completion = ToString(messageVal)
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

func InstallErrorCause(runtime *Runtime, object *Object, options *JavaScriptValue) *Completion {
	if options.Type != TypeObject {
		return NewUnusedCompletion()
	}

	optionsObj := options.Value.(ObjectInterface)
	completion := optionsObj.HasProperty(runtime, causeStr)
	if completion.Type != Normal {
		return completion
	}

	hasCause := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !hasCause {
		return NewUnusedCompletion()
	}

	completion = optionsObj.Get(runtime, causeStr, options)
	if completion.Type != Normal {
		return completion
	}

	causeVal := completion.Value.(*JavaScriptValue)
	completion = object.DefineOwnProperty(runtime, causeStr, &DataPropertyDescriptor{
		Value:        causeVal,
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: Failed to define 'cause' property on Error object.")
	}
	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		panic("Assert failed: Failed to define 'cause' property on Error object.")
	}

	return NewUnusedCompletion()
}
