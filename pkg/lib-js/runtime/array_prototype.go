package runtime

import (
	"math"
)

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

	// Array.prototype.copyWithin
	DefineBuiltinFunction(runtime, obj, "copyWithin", ArrayPrototypeCopyWithin, 2)

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

func ArrayPrototypeCopyWithin(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	for idx := range 3 {
		if idx >= len(arguments) {
			arguments = append(arguments, NewUndefinedValue())
		}
	}

	targetArg := arguments[0]
	startArg := arguments[1]
	endArg := arguments[2]

	completion := ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(ObjectInterface)

	completion = LengthOfArrayLike(runtime, object)
	if completion.Type != Normal {
		return completion
	}

	len := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	completion = ToIntegerOrInfinity(targetArg)
	if completion.Type != Normal {
		return completion
	}

	relativeTarget := completion.Value.(*JavaScriptValue)
	to := ToRelativeIndex(relativeTarget, len)

	completion = ToIntegerOrInfinity(startArg)
	if completion.Type != Normal {
		return completion
	}
	relativeStart := completion.Value.(*JavaScriptValue)
	from := ToRelativeIndex(relativeStart, len)

	var relativeEnd *JavaScriptValue
	if endArg.Type == TypeUndefined {
		relativeEnd = NewNumberValue(len, false)
	} else {
		completion = ToIntegerOrInfinity(endArg)
		if completion.Type != Normal {
			return completion
		}
		relativeEnd = completion.Value.(*JavaScriptValue)
	}
	final := ToRelativeIndex(relativeEnd, len)

	count := math.Min(final-from, len-to)

	var direction float64
	if from < to && to < from+count {
		direction = -1
		from = from + count - 1
		to = to + count - 1
	} else {
		direction = 1
	}

	for count > 0 {
		completion = ToString(NewNumberValue(float64(from), false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		fromKey := completion.Value.(*JavaScriptValue)

		completion = ToString(NewNumberValue(float64(to), false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		toKey := completion.Value.(*JavaScriptValue)

		completion = object.HasProperty(fromKey)
		if completion.Type != Normal {
			return completion
		}

		hasFrom := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
		if hasFrom {
			completion = object.Get(runtime, fromKey, objectVal)
			if completion.Type != Normal {
				return completion
			}

			fromValue := completion.Value.(*JavaScriptValue)

			completion = object.Set(runtime, toKey, fromValue, objectVal)
			if completion.Type != Normal {
				return completion
			}
			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError("Failed to set property."))
			}
		} else {
			completion = object.Delete(toKey)
			if completion.Type != Normal {
				return completion
			}
			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError("Failed to delete property."))
			}
		}

		from += direction
		to += direction
		count -= 1
	}

	return NewNormalCompletion(objectVal)
}

func ToRelativeIndex(value *JavaScriptValue, length float64) float64 {
	if value.Type != TypeNumber {
		panic("Assert failed: ToRelativeIndex value is not a number.")
	}

	relativeIndex := value.Value.(*Number).Value
	if relativeIndex == math.Inf(-1) {
		return 0
	} else if relativeIndex < 0 {
		return math.Max(length+relativeIndex, 0)
	} else {
		return math.Min(relativeIndex, length)
	}
}
