package runtime

import "fmt"

func NewTypedArrayConstructor(
	runtime *Runtime,
	typedArrayName TypedArrayName,
	prototype Intrinsic,
) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		func(
			runtime *Runtime,
			function *FunctionObject,
			thisArg *JavaScriptValue,
			arguments []*JavaScriptValue,
			newTarget *JavaScriptValue,
		) *Completion {
			return TypedArrayConstructor(
				runtime,
				typedArrayName,
				prototype,
				function,
				thisArg,
				arguments,
				newTarget,
			)
		},
		1,
		NewStringValue(string(typedArrayName)),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(prototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	elementSize, ok := TypedArrayElementSizes[typedArrayName]
	if !ok {
		panic("Assert failed: Provided typedArrayName is not mapped in TypedArrayElementSizes.")
	}

	// BYTES_PER_ELEMENT
	constructor.DefineOwnProperty(runtime, NewStringValue("BYTES_PER_ELEMENT"), &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(elementSize), false),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return constructor
}

func TypedArrayConstructor(
	runtime *Runtime,
	typedArrayName TypedArrayName,
	prototype Intrinsic,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if newTarget == nil || newTarget.Type == TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, fmt.Sprintf("%s constructor requires 'new'", typedArrayName)))
	}

	newTargetObj := newTarget.Value.(FunctionInterface)

	if len(arguments) == 0 {
		return AllocateTypedArray(runtime, typedArrayName, newTargetObj, prototype, 0)
	}

	firstArg := arguments[0]

	if firstArg.Type == TypeObject {
		completion := AllocateTypedArrayWithNoBuffer(runtime, typedArrayName, newTargetObj, prototype)
		if completion.Type != Normal {
			return completion
		}

		obj := completion.Value.(*JavaScriptValue).Value.(*TypedArrayObject)

		if _, ok := firstArg.Value.(*TypedArrayObject); ok {
			panic("TODO: Implement TypedArray constructor with typed array argument.")
		} else if firstArgObj, ok := firstArg.Value.(*Object); ok && firstArgObj.ArrayBufferData != nil {
			var byteOffset *JavaScriptValue
			var length *JavaScriptValue

			if len(arguments) > 1 {
				byteOffset = arguments[1]
			} else {
				byteOffset = NewUndefinedValue()
			}

			if len(arguments) > 2 {
				length = arguments[2]
			} else {
				length = NewUndefinedValue()
			}

			completion = InitializeTypedArrayFromArrayBuffer(runtime, obj, firstArgObj, byteOffset, length)
			if completion.Type != Normal {
				return completion
			}
		} else {
			completion := GetMethod(runtime, firstArg, runtime.SymbolIterator)
			if completion.Type != Normal {
				return completion
			}

			iteratorMethod := completion.Value.(*JavaScriptValue)
			if iteratorMethod.Type != TypeUndefined {
				completion := GetIteratorFromMethod(runtime, iteratorMethod, firstArg)
				if completion.Type != Normal {
					return completion
				}

				iterator := completion.Value.(*Iterator)
				completion = IteratorToList(runtime, iterator)
				if completion.Type != Normal {
					return completion
				}

				values := completion.Value.([]*JavaScriptValue)
				completion = InitializeTypedArrayFromList(runtime, obj, values)
				if completion.Type != Normal {
					return completion
				}
			} else {
				panic("TODO: Implement TypedArray constructor with array-like argument.")
			}
		}

		return NewNormalCompletion(NewJavaScriptValue(TypeObject, obj))
	}

	completion := ToIndex(runtime, firstArg)
	if completion.Type != Normal {
		return completion
	}

	length := completion.Value.(*JavaScriptValue).Value.(*Number).Value
	return AllocateTypedArray(runtime, typedArrayName, newTargetObj, prototype, uint(length))
}
