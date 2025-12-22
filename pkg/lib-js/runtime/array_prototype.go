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

	// Array.prototype.entries
	DefineBuiltinFunction(runtime, obj, "entries", ArrayPrototypeEntries, 0)

	// Array.prototype.every
	DefineBuiltinFunction(runtime, obj, "every", ArrayPrototypeEvery, 1)

	// Array.prototype.fill
	DefineBuiltinFunction(runtime, obj, "fill", ArrayPrototypeFill, 1)

	// Array.prototype.filter
	DefineBuiltinFunction(runtime, obj, "filter", ArrayPrototypeFilter, 1)

	// Array.prototype.find
	DefineBuiltinFunction(runtime, obj, "find", ArrayPrototypeFind, 1)

	// Array.prototype.findIndex
	DefineBuiltinFunction(runtime, obj, "findIndex", ArrayPrototypeFindIndex, 1)

	// Array.prototype.findLast
	DefineBuiltinFunction(runtime, obj, "findLast", ArrayPrototypeFindLast, 1)

	// Array.prototype.findLastIndex
	DefineBuiltinFunction(runtime, obj, "findLastIndex", ArrayPrototypeFindLastIndex, 1)

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

func ArrayPrototypeEntries(
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

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(ObjectInterface)

	iterator := CreateArrayIterator(runtime, object, ArrayIteratorKindEntry)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, iterator))
}

func ArrayPrototypeEvery(
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

	callback := arguments[0]
	callbackThisArg := arguments[1]

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

	callbackFunction, ok := callback.Value.(*FunctionObject)
	if !ok {
		return NewThrowCompletion(NewTypeError("Callback is not a callable."))
	}

	for k := range int(len) {
		kNumber := NewNumberValue(float64(k), false)
		completion = ToString(kNumber)
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		key := completion.Value.(*JavaScriptValue)

		completion = object.HasProperty(key)
		if completion.Type != Normal {
			return completion
		}

		hasProperty := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
		if !hasProperty {
			continue
		}

		completion = object.Get(runtime, key, objectVal)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		completion = callbackFunction.Call(runtime, callbackThisArg, []*JavaScriptValue{value, kNumber, objectVal})
		if completion.Type != Normal {
			return completion
		}

		completion = ToBoolean(completion.Value.(*JavaScriptValue))
		if completion.Type != Normal {
			return completion
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(NewBooleanValue(false))
		}
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func ArrayPrototypeFill(
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

	completion := ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}
	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	objectVal := completion.Value.(*JavaScriptValue)

	completion = LengthOfArrayLike(runtime, object)
	if completion.Type != Normal {
		return completion
	}
	len := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	value := arguments[0]
	start := arguments[1]
	end := arguments[2]

	completion = ToIntegerOrInfinity(start)
	if completion.Type != Normal {
		return completion
	}
	relativeStart := completion.Value.(*JavaScriptValue)

	k := ToRelativeIndex(relativeStart, len)

	var relativeEnd *JavaScriptValue
	if end.Type == TypeUndefined {
		relativeEnd = NewNumberValue(len, false)
	} else {
		completion := ToIntegerOrInfinity(end)
		if completion.Type != Normal {
			return completion
		}
		relativeEnd = completion.Value.(*JavaScriptValue)
	}

	final := ToRelativeIndex(relativeEnd, len)

	for k < final {
		completion = ToString(NewNumberValue(k, false))
		if completion.Type != Normal {
			return completion
		}
		pk := completion.Value.(*JavaScriptValue)

		completion := object.Set(runtime, pk, value, objectVal)
		if completion.Type != Normal {
			return completion
		}
		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError("Failed to set property."))
		}

		k++
	}

	return NewNormalCompletion(objectVal)
}

func ArrayPrototypeFilter(
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

	callback := arguments[0]
	thisArgument := arguments[1]

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

	callbackFunc, ok := callback.Value.(*FunctionObject)
	if !ok {
		return NewThrowCompletion(NewTypeError("Callback is not a function"))
	}

	completion = ArraySpeciesCreate(runtime, objectVal, 0)
	if completion.Type != Normal {
		return completion
	}

	a := completion.Value.(*JavaScriptValue)
	aObject := a.Value.(ObjectInterface)

	to := 0.0

	for k := range int(len) {
		kNumber := NewNumberValue(float64(k), false)
		completion = ToString(kNumber)
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		pk := completion.Value.(*JavaScriptValue)

		completion = object.HasProperty(pk)
		if completion.Type != Normal {
			return completion
		}
		kPresent := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

		if !kPresent {
			continue
		}

		completion = object.Get(runtime, pk, objectVal)
		if completion.Type != Normal {
			return completion
		}
		kValue := completion.Value.(*JavaScriptValue)

		completion = callbackFunc.Call(runtime, thisArgument, []*JavaScriptValue{kValue, kNumber, objectVal})
		if completion.Type != Normal {
			return completion
		}

		completion := ToBoolean(completion.Value.(*JavaScriptValue))
		if completion.Type != Normal {
			return completion
		}
		selected := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

		if !selected {
			continue
		}

		completion = ToString(NewNumberValue(to, false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		toKey := completion.Value.(*JavaScriptValue)

		completion = CreateDataProperty(aObject, toKey, kValue)
		if completion.Type != Normal {
			return completion
		}
		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError("Failed to create data property"))
		}

		to++
	}

	return NewNormalCompletion(a)
}

func ArrayPrototypeFind(
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

	predicate := arguments[0]
	thisArgument := arguments[1]

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

	result := FindViaPredicate(runtime, object, uint(len), false, predicate, thisArgument)
	if result.Type != Normal {
		return result
	}

	resultValue := result.Value.(*FindViaPredicateResult)
	return NewNormalCompletion(resultValue.Value)
}

func ArrayPrototypeFindIndex(
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

	predicate := arguments[0]
	thisArgument := arguments[1]

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

	result := FindViaPredicate(runtime, object, uint(len), false, predicate, thisArgument)
	if result.Type != Normal {
		return result
	}

	resultValue := result.Value.(*FindViaPredicateResult)
	return NewNormalCompletion(NewNumberValue(resultValue.Index, false))
}

func ArrayPrototypeFindLast(
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

	predicate := arguments[0]
	thisArgument := arguments[1]

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

	result := FindViaPredicate(runtime, object, uint(len), true, predicate, thisArgument)
	if result.Type != Normal {
		return result
	}

	resultValue := result.Value.(*FindViaPredicateResult)
	return NewNormalCompletion(resultValue.Value)
}

func ArrayPrototypeFindLastIndex(
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

	predicate := arguments[0]
	thisArgument := arguments[1]

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

	result := FindViaPredicate(runtime, object, uint(len), true, predicate, thisArgument)
	if result.Type != Normal {
		return result
	}

	resultValue := result.Value.(*FindViaPredicateResult)
	return NewNormalCompletion(NewNumberValue(resultValue.Index, false))
}

type FindViaPredicateResult struct {
	Index float64
	Value *JavaScriptValue
}

func FindViaPredicate(
	runtime *Runtime,
	object ObjectInterface,
	length uint,
	isDescending bool,
	predicate *JavaScriptValue,
	thisArg *JavaScriptValue,
) *Completion {
	objectVal := NewJavaScriptValue(TypeObject, object)

	functionObj, ok := predicate.Value.(*FunctionObject)
	if !ok {
		return NewThrowCompletion(NewTypeError("Predicate is not callable."))
	}

	loopBody := func(index uint) *Completion {
		indexNumber := NewNumberValue(float64(index), false)
		completion := ToString(indexNumber)
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		key := completion.Value.(*JavaScriptValue)

		completion = object.Get(runtime, key, objectVal)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		completion = functionObj.Call(runtime, thisArg, []*JavaScriptValue{value, indexNumber, objectVal})
		if completion.Type != Normal {
			return completion
		}

		completion = ToBoolean(completion.Value.(*JavaScriptValue))
		if completion.Type != Normal {
			return completion
		}

		if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(&FindViaPredicateResult{Index: float64(index), Value: value})
		}
		return nil
	}

	if isDescending {
		for idx := range length {
			result := loopBody(length - idx - 1)
			if result != nil {
				return result
			}
		}
	} else {
		for idx := range length {
			result := loopBody(idx)
			if result != nil {
				return result
			}
		}
	}
	return NewNormalCompletion(&FindViaPredicateResult{
		Index: -1,
		Value: NewUndefinedValue(),
	})
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
