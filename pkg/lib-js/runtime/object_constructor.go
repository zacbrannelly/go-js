package runtime

func NewObjectConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ObjectConstructor,
		1,
		NewStringValue("Object"),
		realm,
		realm.Intrinsics[IntrinsicFunctionPrototype],
	)
	MakeConstructor(constructor)

	// Object.assign
	DefineBuiltinFunction(runtime, constructor, "assign", ObjectAssign, 2)

	return constructor
}

func ObjectConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	activeFunction := runtime.GetRunningExecutionContext().Function
	if newTarget != nil && newTarget.Type != TypeUndefined && newTarget.Value != activeFunction {
		panic("TODO: Support NewTarget in Object constructor.")
	}

	if len(arguments) == 0 || arguments[0].Type == TypeUndefined || arguments[0].Type == TypeNull {
		newObj := OrdinaryObjectCreate(function.Realm.Intrinsics[IntrinsicObjectPrototype])
		return NewNormalCompletion(NewJavaScriptValue(TypeObject, newObj))
	}

	completion := ToObject(arguments[0])
	if completion.Type != Normal {
		panic("Assert failed: ToObject threw an error when it should not have.")
	}

	return completion
}

func ObjectAssign(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	target := arguments[0]
	if len(arguments) == 1 {
		return NewNormalCompletion(target)
	}

	completion := ToObject(target)
	if completion.Type != Normal {
		return completion
	}

	target = completion.Value.(*JavaScriptValue)
	targetObj := target.Value.(ObjectInterface)

	for _, source := range arguments[1:] {
		if source.Type == TypeUndefined || source.Type == TypeNull {
			continue
		}

		completion = ToObject(source)
		if completion.Type != Normal {
			panic("Assert failed: ToObject threw an error when it should not have.")
		}

		fromObj := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

		completion = fromObj.OwnPropertyKeys()
		if completion.Type != Normal {
			return completion
		}

		keys := completion.Value.([]*JavaScriptValue)
		for _, key := range keys {
			completion = fromObj.GetOwnProperty(key)
			if completion.Type != Normal {
				return completion
			}

			if desc, ok := completion.Value.(PropertyDescriptor); !ok || desc == nil || !desc.GetEnumerable() {
				continue
			}

			completion = fromObj.Get(runtime, key, source)
			if completion.Type != Normal {
				return completion
			}

			propValue := completion.Value.(*JavaScriptValue)
			completion = targetObj.Set(runtime, key, propValue, target)
			if completion.Type != Normal {
				return completion
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				// TODO: Improve the error message.
				return NewThrowCompletion(NewTypeError("Failed to set property."))
			}
		}
	}

	return NewNormalCompletion(target)
}
