package runtime

var LengthString = NewStringValue("length")

func NewArrayPrototype(runtime *Runtime) ObjectInterface {
	obj := NewArrayObject(runtime, 0)
	obj.Prototype = runtime.GetRunningRealm().Intrinsics[IntrinsicObjectPrototype]

	// Array.prototype.at
	DefineBuiltinFunction(runtime, obj, "at", ArrayPrototypeAt, 1)

	// Array.prototype.values
	DefineBuiltinFunction(runtime, obj, "values", ArrayPrototypeValues, 0)

	// Array.prototype[%Symbol.iterator%]
	DefineBuiltinSymbolFunction(runtime, obj, runtime.SymbolIterator, ArrayPrototypeValues, 0)

	// TODO: Implement other methods.

	return obj
}

func ArrayPrototypeAt(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	objectCompletion := ToObject(thisArg)
	if objectCompletion.Type != Normal {
		return objectCompletion
	}

	object := objectCompletion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	lenCompletion := LengthOfArrayLike(runtime, object)
	if lenCompletion.Type != Normal {
		return lenCompletion
	}

	len := lenCompletion.Value.(*JavaScriptValue).Value.(*Number).Value

	relativeIndexCompletion := ToIntegerOrInfinity(arguments[0])
	if relativeIndexCompletion.Type != Normal {
		return relativeIndexCompletion
	}

	relativeIndex := relativeIndexCompletion.Value.(*JavaScriptValue).Value.(*Number).Value

	// e.g. -1 -> len - 1
	if relativeIndex < 0 {
		relativeIndex += len
	}

	// Return undefined if out of bounds.
	if relativeIndex < 0 || relativeIndex >= len {
		return NewNormalCompletion(NewUndefinedValue())
	}

	// Convert the index to a string.
	keyCompletion := ToString(NewNumberValue(float64(relativeIndex), false))
	if keyCompletion.Type != Normal {
		return keyCompletion
	}
	key := keyCompletion.Value.(*JavaScriptValue)

	// Get the element at the index.
	return object.Get(runtime, key, NewJavaScriptValue(TypeObject, object))
}

func ArrayPrototypeValues(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	iterator := CreateArrayIterator(runtime, object, ArrayIteratorKindValue)

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, iterator))
}

func LengthOfArrayLike(runtime *Runtime, object ObjectInterface) *Completion {
	// Get the length property.
	objectValue := NewJavaScriptValue(TypeObject, object)
	lenCompletion := object.Get(runtime, LengthString, objectValue)
	if lenCompletion.Type != Normal {
		return lenCompletion
	}

	// Coerce the value to be a integer length.
	len := lenCompletion.Value.(*JavaScriptValue)
	lenCompletion = ToLength(len)
	return lenCompletion
}
