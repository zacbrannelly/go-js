package runtime

func NewNumberConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		NumberConstructor,
		1,
		NewStringValue("Number"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// Number.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicNumberPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// TODO: Define other properties.

	return constructor
}

func NumberConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	var numberValue *JavaScriptValue = nil
	if len(arguments) == 0 {
		numberValue = NewNumberValue(0, false)
	} else {
		completion := ToNumber(runtime, arguments[0])
		if completion.Type != Normal {
			return completion
		}
		numberValue = completion.Value.(*JavaScriptValue)

		if numberValue.Type == TypeBigInt {
			panic("TODO: Implement BigInt to Number conversion.")
		}
	}

	if newTarget == nil || newTarget.Type == TypeUndefined {
		return NewNormalCompletion(numberValue)
	}

	newTargetObj := newTarget.Value.(FunctionInterface)
	completion := OrdinaryCreateFromConstructor(runtime, newTargetObj, IntrinsicNumberPrototype)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(*Object)
	object.NumberData = numberValue

	return NewNormalCompletion(objectVal)
}
