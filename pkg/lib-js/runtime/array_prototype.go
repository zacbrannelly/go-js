package runtime

var LengthString = NewStringValue("length")

func NewArrayPrototype(runtime *Runtime) ObjectInterface {
	obj := NewArrayObject(runtime, 0)
	obj.Prototype = runtime.GetRunningRealm().Intrinsics[IntrinsicObjectPrototype]

	// Array.prototype.at
	DefineBuiltinFunction(runtime, obj, "at", ArrayPrototypeAt, 1)

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
	lenCompletion := LengthOfArrayLike(object)
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
	return object.Get(key, NewJavaScriptValue(TypeObject, object))
}

func LengthOfArrayLike(object ObjectInterface) *Completion {
	// Get the length property.
	objectValue := NewJavaScriptValue(TypeObject, object)
	lenCompletion := object.Get(LengthString, objectValue)
	if lenCompletion.Type != Normal {
		return lenCompletion
	}

	// Coerce the value to be a integer length.
	len := lenCompletion.Value.(*JavaScriptValue)
	lenCompletion = ToLength(len)
	return lenCompletion
}
