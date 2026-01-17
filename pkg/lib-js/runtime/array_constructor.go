package runtime

func NewArrayConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ArrayConstructor,
		1,
		NewStringValue("Array"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// Array.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicArrayPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// Array.isArray
	DefineBuiltinFunction(runtime, constructor, "isArray", ArrayConstructorIsArray, 1)

	// TODO: Define other properties.

	return constructor
}

func ArrayConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	var newTargetObj FunctionInterface = nil
	if newTarget == nil || newTarget.Type == TypeUndefined {
		newTarget = NewJavaScriptValue(TypeObject, runtime.GetRunningExecutionContext().Function)
	}

	newTargetObj = newTarget.Value.(FunctionInterface)

	completion := GetPrototypeFromConstructor(runtime, newTargetObj, IntrinsicArrayPrototype)
	if completion.Type != Normal {
		return completion
	}

	prototype := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	if len(arguments) == 0 {
		return ArrayCreateWithPrototype(runtime, 0, prototype)
	} else if len(arguments) == 1 {
		completion := ArrayCreateWithPrototype(runtime, 0, prototype)
		if completion.Type != Normal {
			panic("Assert failed: ArrayCreate threw an unexpected error.")
		}

		arrayVal := completion.Value.(*JavaScriptValue)
		arrayObj := arrayVal.Value.(ObjectInterface)

		var intLen *JavaScriptValue
		if arguments[0].Type != TypeNumber {
			completion = CreateDataProperty(runtime, arrayObj, zeroString, arguments[0])
			if completion.Type != Normal {
				panic("Assert failed: CreateDataProperty threw an unexpected error.")
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				panic("Assert failed: CreateDataProperty returned false.")
			}

			intLen = NewNumberValue(1, false)
		} else {
			completion := ToUint32(runtime, arguments[0])
			if completion.Type != Normal {
				panic("Assert failed: ToUint32 threw an unexpected error.")
			}

			intLen = completion.Value.(*JavaScriptValue)

			completion = SameValueZero(intLen, arguments[0])
			if completion.Type != Normal {
				panic("Assert failed: SameValueZero threw an unexpected error.")
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewRangeError(runtime, "Array length must be a non-negative integer."))
			}
		}

		completion = arrayObj.Set(runtime, lengthStr, intLen, arrayVal)
		if completion.Type != Normal {
			panic("Assert failed: ArrayObject Set threw an unexpected error.")
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: ArrayObject Set returned false.")
		}

		return NewNormalCompletion(arrayVal)
	}

	completion = ArrayCreateWithPrototype(runtime, uint(len(arguments)), prototype)
	if completion.Type != Normal {
		return completion
	}

	arrayVal := completion.Value.(*JavaScriptValue)
	arrayObj := arrayVal.Value.(ObjectInterface)

	for i, arg := range arguments {
		completion = ToString(runtime, NewNumberValue(float64(i), false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}

		key := completion.Value.(*JavaScriptValue)

		completion = CreateDataProperty(runtime, arrayObj, key, arg)
		if completion.Type != Normal {
			panic("Assert failed: CreateDataProperty threw an unexpected error.")
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: CreateDataProperty returned false.")
		}
	}

	return NewNormalCompletion(arrayVal)
}

func ArrayConstructorIsArray(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		arguments = append(arguments, NewUndefinedValue())
	}

	return IsArray(runtime, arguments[0])
}
