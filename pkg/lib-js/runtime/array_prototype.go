package runtime

import (
	"math"
)

var lengthStr = NewStringValue("length")

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

	// Array.prototype.flat
	DefineBuiltinFunction(runtime, obj, "flat", ArrayPrototypeFlat, 0)

	// Array.prototype.flatMap
	DefineBuiltinFunction(runtime, obj, "flatMap", ArrayPrototypeFlatMap, 1)

	// Array.prototype.forEach
	DefineBuiltinFunction(runtime, obj, "forEach", ArrayPrototypeForEach, 1)

	// Array.prototype.includes
	DefineBuiltinFunction(runtime, obj, "includes", ArrayPrototypeIncludes, 1)

	// Array.prototype.indexOf
	DefineBuiltinFunction(runtime, obj, "indexOf", ArrayPrototypeIndexOf, 1)

	// Array.prototype.join
	DefineBuiltinFunction(runtime, obj, "join", ArrayPrototypeJoin, 1)

	// Array.prototype.keys
	DefineBuiltinFunction(runtime, obj, "keys", ArrayPrototypeKeys, 0)

	// Array.prototype.lastIndexOf
	DefineBuiltinFunction(runtime, obj, "lastIndexOf", ArrayPrototypeLastIndexOf, 1)

	// Array.prototype.map
	DefineBuiltinFunction(runtime, obj, "map", ArrayPrototypeMap, 1)

	// Array.prototype.pop
	DefineBuiltinFunction(runtime, obj, "pop", ArrayPrototypePop, 0)

	// Array.prototype.push
	DefineBuiltinFunction(runtime, obj, "push", ArrayPrototypePush, 1)

	// Array.prototype.reduce
	DefineBuiltinFunction(runtime, obj, "reduce", ArrayPrototypeReduce, 1)

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
	lenCompletion := object.Get(runtime, lengthStr, objectValue)
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

func ArrayPrototypeFlat(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = append(arguments, NewUndefinedValue())
	}

	depth := arguments[0]

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

	depthNum := 1.0
	if depth.Type != TypeUndefined {
		completion = ToIntegerOrInfinity(depth)
		if completion.Type != Normal {
			return completion
		}
		depthNum = completion.Value.(*JavaScriptValue).Value.(*Number).Value

		if depthNum < 0 {
			depthNum = 0
		}
	}

	completion = ArraySpeciesCreate(runtime, objectVal, 0)
	if completion.Type != Normal {
		return completion
	}

	array := completion.Value.(*JavaScriptValue)
	arrayObject := array.Value.(ObjectInterface)

	completion = FlattenIntoArray(runtime, arrayObject, object, uint(len), 0, depthNum, nil, nil)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(array)
}

func ArrayPrototypeFlatMap(
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

	mapperFunction := arguments[0]
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

	completion = ArraySpeciesCreate(runtime, objectVal, 0)
	if completion.Type != Normal {
		return completion
	}

	array := completion.Value.(*JavaScriptValue)
	arrayObject := array.Value.(ObjectInterface)

	completion = FlattenIntoArray(runtime, arrayObject, object, uint(len), 0, 1, mapperFunction, thisArgument)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(array)
}

func ArrayPrototypeForEach(
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
		return NewThrowCompletion(NewTypeError("Callback is not a function."))
	}

	for idx := range int(len) {
		kNumber := NewNumberValue(float64(idx), false)
		completion := ToString(kNumber)
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

		completion = callbackFunc.Call(runtime, thisArgument, []*JavaScriptValue{value, kNumber, objectVal})
		if completion.Type != Normal {
			return completion
		}
	}

	return NewNormalCompletion(NewUndefinedValue())
}

func ArrayPrototypeIncludes(
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

	searchElement := arguments[0]
	fromIndex := arguments[1]

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

	if len == 0 {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	completion = ToIntegerOrInfinity(fromIndex)
	if completion.Type != Normal {
		return completion
	}

	n := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	if n == math.Inf(1) {
		return NewNormalCompletion(NewBooleanValue(false))
	} else if n == math.Inf(-1) {
		n = 0
	}

	var k float64
	if n >= 0 {
		k = n
	} else {
		k = len + n
		if k < 0 {
			k = 0
		}
	}

	for k < len {
		completion = ToString(NewNumberValue(k, false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		key := completion.Value.(*JavaScriptValue)

		completion = object.Get(runtime, key, objectVal)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		completion = SameValueZero(value, searchElement)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(NewBooleanValue(true))
		}

		k++
	}

	return NewNormalCompletion(NewBooleanValue(false))
}

func ArrayPrototypeIndexOf(
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

	searchElement := arguments[0]
	fromIndex := arguments[1]

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

	if len == 0 {
		return NewNormalCompletion(NewNumberValue(-1, false))
	}

	completion = ToIntegerOrInfinity(fromIndex)
	if completion.Type != Normal {
		return completion
	}

	n := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	if n == math.Inf(1) {
		return NewNormalCompletion(NewNumberValue(-1, false))
	} else if n == math.Inf(-1) {
		n = 0
	}

	var k float64
	if n >= 0 {
		k = n
	} else {
		k = len + n
		if k < 0 {
			k = 0
		}
	}

	for k < len {
		kNumber := NewNumberValue(k, false)
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
			k++
			continue
		}

		completion = object.Get(runtime, key, objectVal)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		completion = IsStrictlyEqual(value, searchElement)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(kNumber)
		}

		k++
	}

	return NewNormalCompletion(NewNumberValue(-1, false))
}

func ArrayPrototypeJoin(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = append(arguments, NewUndefinedValue())
	}

	separator := arguments[0]

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

	if separator.Type == TypeUndefined {
		separator = NewStringValue(",")
	} else {
		completion = ToString(separator)
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		separator = completion.Value.(*JavaScriptValue)
	}

	resultString := ""

	for idx := range int(len) {
		if idx > 0 {
			resultString += separator.Value.(*String).Value
		}

		kNumber := NewNumberValue(float64(idx), false)
		completion = ToString(kNumber)
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		key := completion.Value.(*JavaScriptValue)

		completion = object.Get(runtime, key, objectVal)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)
		completion = ToString(value)
		if completion.Type != Normal {
			return completion
		}

		valueStr := completion.Value.(*JavaScriptValue).Value.(*String).Value
		resultString += valueStr
	}

	return NewNormalCompletion(NewStringValue(resultString))
}

func ArrayPrototypeKeys(
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

	iterator := CreateArrayIterator(runtime, object, ArrayIteratorKindKey)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, iterator))
}

func ArrayPrototypeLastIndexOf(
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

	searchElement := arguments[0]
	fromIndex := arguments[1]

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

	if len == 0 {
		return NewNormalCompletion(NewNumberValue(-1, false))
	}

	var n float64
	if fromIndex.Type != TypeUndefined {
		completion = ToIntegerOrInfinity(fromIndex)
		if completion.Type != Normal {
			return completion
		}
		n = completion.Value.(*JavaScriptValue).Value.(*Number).Value
	} else {
		n = len - 1
	}

	if n == math.Inf(-1) {
		return NewNormalCompletion(NewNumberValue(-1, false))
	}

	var k float64
	if n >= 0 {
		k = math.Min(n, len-1)
	} else {
		k = len + n
	}

	for k >= 0 {
		kNumber := NewNumberValue(k, false)
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
			k--
			continue
		}

		completion = object.Get(runtime, key, objectVal)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		completion = IsStrictlyEqual(value, searchElement)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(kNumber)
		}

		k--
	}

	return NewNormalCompletion(NewNumberValue(-1, false))
}

func ArrayPrototypeMap(
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
		return NewThrowCompletion(NewTypeError("Callback is not a function."))
	}

	completion = ArraySpeciesCreate(runtime, objectVal, uint(len))
	if completion.Type != Normal {
		return completion
	}

	array := completion.Value.(*JavaScriptValue)
	arrayObject := array.Value.(ObjectInterface)

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

		completion = callbackFunc.Call(runtime, thisArgument, []*JavaScriptValue{value, kNumber, objectVal})
		if completion.Type != Normal {
			return completion
		}

		value = completion.Value.(*JavaScriptValue)

		completion = CreateDataProperty(arrayObject, key, value)
		if completion.Type != Normal {
			return completion
		}
		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError("Failed to create data property."))
		}
	}

	return NewNormalCompletion(array)
}

func ArrayPrototypePop(
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

	completion = LengthOfArrayLike(runtime, object)
	if completion.Type != Normal {
		return completion
	}

	len := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	if len == 0 {
		completion = object.Set(runtime, lengthStr, NewNumberValue(0, false), objectVal)
		if completion.Type != Normal {
			return completion
		}
		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError("Failed to set length property."))
		}
		return NewNormalCompletion(NewUndefinedValue())
	}

	newLen := NewNumberValue(len-1, false)
	completion = ToString(newLen)
	if completion.Type != Normal {
		panic("Assert failed: ToString threw an unexpected error.")
	}
	newLenKey := completion.Value.(*JavaScriptValue)

	completion = object.Get(runtime, newLenKey, objectVal)
	if completion.Type != Normal {
		return completion
	}

	element := completion.Value.(*JavaScriptValue)

	completion = object.Delete(newLenKey)
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError("Failed to delete property."))
	}

	completion = object.Set(runtime, lengthStr, newLen, objectVal)
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError("Failed to set length property."))
	}

	return NewNormalCompletion(element)
}

func ArrayPrototypePush(
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

	completion = LengthOfArrayLike(runtime, object)
	if completion.Type != Normal {
		return completion
	}

	length := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	argCount := len(arguments)
	if float64(argCount)+length > 2^53-1 {
		return NewThrowCompletion(NewTypeError("Array length too large."))
	}

	for _, arg := range arguments {
		completion = ToString(NewNumberValue(float64(length), false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}

		key := completion.Value.(*JavaScriptValue)

		completion = object.Set(runtime, key, arg, objectVal)
		if completion.Type != Normal {
			return completion
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError("Failed to set property."))
		}

		length++
	}

	newLength := NewNumberValue(length, false)
	completion = object.Set(runtime, lengthStr, newLength, objectVal)
	if completion.Type != Normal {
		return completion
	}
	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError("Failed to set length property."))
	}

	return NewNormalCompletion(newLength)
}

func ArrayPrototypeReduce(
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
	initialValue := arguments[1]

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

	length := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	callbackFunc, ok := callback.Value.(*FunctionObject)
	if !ok {
		return NewThrowCompletion(NewTypeError("Callback is not a function."))
	}

	if length == 0 && initialValue.Type != TypeUndefined {
		return NewThrowCompletion(NewTypeError("Array is empty and no initial value was provided."))
	}

	k := 0.0
	accumulator := initialValue

	if initialValue.Type == TypeUndefined {
		isPresent := false

		for !isPresent && k < length {
			completion = ToString(NewNumberValue(k, false))
			if completion.Type != Normal {
				panic("Assert failed: ToString threw an unexpected error.")
			}

			key := completion.Value.(*JavaScriptValue)

			completion = object.HasProperty(key)
			if completion.Type != Normal {
				return completion
			}

			isPresent := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
			if isPresent {
				completion = object.Get(runtime, key, objectVal)
				if completion.Type != Normal {
					return completion
				}

				accumulator = completion.Value.(*JavaScriptValue)
			}

			k++
		}

		if !isPresent {
			return NewThrowCompletion(NewTypeError("Unable to find initial value."))
		}
	}

	for k < length {
		kNumber := NewNumberValue(k, false)
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
			k++
			continue
		}

		completion = object.Get(runtime, key, objectVal)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		completion = callbackFunc.Call(runtime, NewUndefinedValue(), []*JavaScriptValue{accumulator, value, kNumber, objectVal})
		if completion.Type != Normal {
			return completion
		}

		accumulator = completion.Value.(*JavaScriptValue)
		k++
	}

	return NewNormalCompletion(accumulator)
}

func FlattenIntoArray(
	runtime *Runtime,
	target ObjectInterface,
	source ObjectInterface,
	sourceLength uint,
	start uint,
	depth float64,
	mapperFunction *JavaScriptValue,
	thisArg *JavaScriptValue,
) *Completion {
	if mapperFunction != nil {
		if _, ok := mapperFunction.Value.(*FunctionObject); !ok {
			return NewThrowCompletion(NewTypeError("Mapper function is callable."))
		}
	}

	targetIndex := float64(start)
	sourceIndex := 0.0

	sourceVal := NewJavaScriptValue(TypeObject, source)

	inf := math.Inf(1)

	for uint(sourceIndex) < sourceLength {
		sourceIndexNumber := NewNumberValue(sourceIndex, false)
		completion := ToString(sourceIndexNumber)
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}
		propertyKey := completion.Value.(*JavaScriptValue)

		completion = source.HasProperty(propertyKey)
		if completion.Type != Normal {
			return completion
		}

		hasProperty := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
		if !hasProperty {
			sourceIndex++
			continue
		}

		completion = source.Get(runtime, propertyKey, sourceVal)
		if completion.Type != Normal {
			return completion
		}

		elementValue := completion.Value.(*JavaScriptValue)

		if mapperFunction != nil {
			mapperFunctionObj := mapperFunction.Value.(*FunctionObject)
			completion = mapperFunctionObj.Call(runtime, thisArg, []*JavaScriptValue{elementValue, sourceIndexNumber, sourceVal})
			if completion.Type != Normal {
				return completion
			}

			elementValue = completion.Value.(*JavaScriptValue)
		}

		shouldFlatten := false

		if depth > 0 {
			completion = IsArray(elementValue)
			if completion.Type != Normal {
				return completion
			}

			shouldFlatten = completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
		}

		if shouldFlatten {
			newDepth := inf
			if depth != math.Inf(1) {
				newDepth = depth - 1
			}

			element := elementValue.Value.(ObjectInterface)
			completion = LengthOfArrayLike(runtime, element)
			if completion.Type != Normal {
				return completion
			}

			elementLen := completion.Value.(*JavaScriptValue).Value.(*Number).Value

			completion = FlattenIntoArray(
				runtime,
				target,
				element,
				uint(elementLen),
				uint(targetIndex),
				newDepth,
				nil,
				nil,
			)
			if completion.Type != Normal {
				return completion
			}

			targetIndex = completion.Value.(*JavaScriptValue).Value.(*Number).Value
		} else {
			if targetIndex >= 2^53-1 {
				return NewThrowCompletion(NewTypeError("Array length too large."))
			}

			completion = ToString(NewNumberValue(targetIndex, false))
			if completion.Type != Normal {
				panic("Assert failed: ToString threw an unexpected error.")
			}
			targetKey := completion.Value.(*JavaScriptValue)

			completion = CreateDataProperty(target, targetKey, elementValue)
			if completion.Type != Normal {
				return completion
			}
			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError("Failed to create data property."))
			}

			targetIndex++
		}

		sourceIndex++
	}

	return NewNormalCompletion(NewNumberValue(targetIndex, false))
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
