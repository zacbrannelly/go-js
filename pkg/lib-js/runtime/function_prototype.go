package runtime

import "math"

func NewFunctionPrototype(runtime *Runtime) ObjectInterface {
	realm := runtime.GetRunningRealm()
	prototype := CreateBuiltinFunction(
		runtime,
		FunctionPrototypeConstructor,
		0,
		NewStringValue(""),
		realm,
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype),
	)

	return prototype
}

func DefineFunctionPrototypeProperties(runtime *Runtime, functionProto ObjectInterface) {
	DefineBuiltinFunction(runtime, functionProto, "call", FunctionPrototypeCall, 1)

	// Function.prototype[Symbol.hasInstance]
	DefineBuiltinSymbolFunction(runtime, functionProto, runtime.SymbolHasInstance, FunctionPrototypeHasInstance, 1)

	// Function.prototype.bind
	DefineBuiltinFunction(runtime, functionProto, "bind", FunctionPrototypeBind, 1)

	// TODO: Define other properties.
}

func FunctionPrototypeConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	return NewNormalCompletion(NewUndefinedValue())
}

func FunctionPrototypeCall(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = append(arguments, NewUndefinedValue())
	}

	callThisArg := arguments[0]
	callArguments := arguments[1:]

	if functionObj, ok := thisArg.Value.(FunctionInterface); ok {
		PrepareForTailCall()
		return functionObj.Call(runtime, callThisArg, callArguments)
	}

	return NewThrowCompletion(NewTypeError(runtime, "'this' is not callable"))
}

func FunctionPrototypeHasInstance(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = append(arguments, NewUndefinedValue())
	}
	return OrdinaryHasInstance(runtime, thisArg, arguments[0])
}

func FunctionPrototypeBind(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = append(arguments, NewUndefinedValue())
	}

	targetVal := thisArg
	if !IsCallable(targetVal) {
		return NewThrowCompletion(NewTypeError(runtime, "'this' is not callable."))
	}

	thisArg = arguments[0]
	target := targetVal.Value.(ObjectInterface)

	completion := BoundFunctionCreate(runtime, target, thisArg, arguments[1:])
	if completion.Type != Normal {
		return completion
	}

	boundFunction := completion.Value.(*JavaScriptValue).Value.(*BoundFunction)

	completion = HasOwnProperty(runtime, target, lengthStr)
	if completion.Type != Normal {
		return completion
	}

	newLength := 0.0

	hasOwnProperty := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if hasOwnProperty {
		completion = target.Get(runtime, lengthStr, targetVal)
		if completion.Type != Normal {
			return completion
		}

		lengthVal := completion.Value.(*JavaScriptValue)

		if number, ok := lengthVal.Value.(*Number); ok {
			if number.Value == math.Inf(1) {
				newLength = math.Inf(1)
			} else if number.Value == math.Inf(-1) {
				newLength = 0.0
			} else {
				completion = ToIntegerOrInfinity(lengthVal)
				if completion.Type != Normal {
					panic("Assert failed: ToIntegerOrInfinity threw an unexpected error.")
				}

				integer := completion.Value.(*JavaScriptValue).Value.(*Number).Value
				newLength = math.Max(integer-float64(len(arguments)-1), 0)
			}
		}
	}

	SetFunctionLength(runtime, boundFunction, int(newLength))

	completion = target.Get(runtime, nameStr, thisArg)
	if completion.Type != Normal {
		return completion
	}

	targetName := completion.Value.(*JavaScriptValue)
	if targetName.Type != TypeString {
		targetName = NewStringValue("")
	}

	SetFunctionNameWithPrefix(runtime, boundFunction, targetName, "bound")

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, boundFunction))
}
