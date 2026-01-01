package runtime

func NewArrayConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ObjectConstructor,
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

	return IsArray(arguments[0])
}
