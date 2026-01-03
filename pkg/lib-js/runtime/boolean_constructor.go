package runtime

func NewBooleanConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		BooleanConstructor,
		1,
		NewStringValue("Boolean"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// Boolean.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicBooleanPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// TODO: Define other properties.

	return constructor
}

func BooleanConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		arguments = append(arguments, NewUndefinedValue())
	}

	completion := ToBoolean(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	if newTarget == nil || newTarget.Type == TypeUndefined {
		return completion
	}

	value := completion.Value.(*JavaScriptValue)
	newTargetObj := newTarget.Value.(FunctionInterface)

	completion = OrdinaryCreateFromConstructor(runtime, newTargetObj, IntrinsicBooleanPrototype)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(*Object)
	object.BooleanData = value

	return NewNormalCompletion(objectVal)
}
