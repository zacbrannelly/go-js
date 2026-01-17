package runtime

func NewProxyObjectConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ProxyObjectConstructor,
		1,
		NewStringValue("Proxy"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// Proxy.revocable
	DefineBuiltinFunction(runtime, constructor, "revocable", ProxyObjectRevocable, 2)

	return constructor
}

func ProxyObjectConstructor(
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
		return NewThrowCompletion(NewTypeError(runtime, "Proxy constructor requires 'new'"))
	}

	target := arguments[0]
	handler := arguments[1]

	return ProxyCreate(runtime, target, handler)
}

func ProxyObjectRevocable(
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

	target := arguments[0]
	handler := arguments[1]

	completion := ProxyCreate(runtime, target, handler)
	if completion.Type != Normal {
		return completion
	}

	proxy := completion.Value.(*JavaScriptValue).Value.(*ProxyObject)

	revokeClosure := func(runtime *Runtime, function *FunctionObject, thisArg *JavaScriptValue, arguments []*JavaScriptValue, newTarget *JavaScriptValue) *Completion {
		activeFunction := runtime.GetRunningExecutionContext().Function
		revocableProxy := activeFunction.RevocableProxy

		if revocableProxy == nil {
			return NewNormalCompletion(NewUndefinedValue())
		}

		activeFunction.RevocableProxy = nil

		revocableProxy.ProxyTarget = NewNullValue()
		revocableProxy.ProxyHandler = NewNullValue()

		return NewNormalCompletion(NewUndefinedValue())
	}

	revoker := CreateBuiltinFunction(runtime, revokeClosure, 0, NewStringValue(""), nil, nil)
	revoker.RevocableProxy = proxy

	result := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))

	// "proxy" property.
	completion = CreateDataProperty(runtime, result, NewStringValue("proxy"), NewJavaScriptValue(TypeObject, proxy))
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError(runtime, "Failed to create data property on result object."))
	}

	// "revoke" property.
	completion = CreateDataProperty(runtime, result, NewStringValue("revoke"), NewJavaScriptValue(TypeObject, revoker))
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError(runtime, "Failed to create data property on result object."))
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, result))
}
