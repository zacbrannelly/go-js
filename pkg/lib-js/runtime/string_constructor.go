package runtime

func NewStringConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		StringConstructor,
		1,
		NewStringValue("String"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// String.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicStringPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// TODO: Define other properties.

	return constructor
}

func StringConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		arguments = append(arguments, NewStringValue(""))
	} else {
		if newTarget != nil && newTarget.Type == TypeUndefined && arguments[0].Type == TypeSymbol {
			panic("TODO: Support Symbol wrapping in String constructor.")
		}
		completion := ToString(runtime, arguments[0])
		if completion.Type != Normal {
			return completion
		}
		arguments[0] = completion.Value.(*JavaScriptValue)
	}

	value := arguments[0]

	// Do casting when called as a function.
	if newTarget == nil || newTarget.Type == TypeUndefined {
		return NewNormalCompletion(value)
	}

	constructor := newTarget.Value.(FunctionInterface)

	completion := GetPrototypeFromConstructor(runtime, constructor, IntrinsicStringPrototype)
	if completion.Type != Normal {
		return completion
	}

	prototype := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	stringObject := StringCreate(runtime, value, prototype)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, stringObject))
}
