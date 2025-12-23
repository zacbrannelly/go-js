package runtime

func NewFunctionConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ObjectConstructor,
		1,
		NewStringValue("Function"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(constructor)

	// Function.prototype
	constructor.DefineOwnProperty(NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicFunctionPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// TODO: Define other properties.

	return constructor
}
