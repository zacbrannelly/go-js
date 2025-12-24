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

	// TODO: Define other properties.

	return constructor
}
