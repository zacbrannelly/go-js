package runtime

func NewMathObject(runtime *Runtime) ObjectInterface {
	mathObj := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))

	// Math.pow
	DefineBuiltinFunction(runtime, mathObj, "pow", MathPow, 2)

	// TODO: Define properties.

	return mathObj
}

func MathPow(
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

	completion := ToNumber(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	base := completion.Value.(*JavaScriptValue).Value.(*Number)

	completion = ToNumber(arguments[1])
	if completion.Type != Normal {
		return completion
	}

	exponent := completion.Value.(*JavaScriptValue).Value.(*Number)

	result := NumberExponentiate(base, exponent)
	return NewNormalCompletion(NewJavaScriptValue(TypeNumber, result))
}
