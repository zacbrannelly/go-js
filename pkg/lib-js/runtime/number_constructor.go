package runtime

import (
	"math"
	"strconv"
)

func NewNumberConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		NumberConstructor,
		1,
		NewStringValue("Number"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// Number.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicNumberPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return constructor
}

func DefineNumberConstructorProperties(runtime *Runtime, constructor ObjectInterface) {
	// Number.MAX_SAFE_INTEGER
	constructor.DefineOwnProperty(runtime, NewStringValue("MAX_SAFE_INTEGER"), &DataPropertyDescriptor{
		Value:        NewNumberValue(math.Pow(2, 53)-1, false),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// Number.parseInt
	constructor.DefineOwnProperty(runtime, NewStringValue("parseInt"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, runtime.GetRunningRealm().GetIntrinsic(IntrinsicParseIntFunction)),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
}

func NumberConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	var numberValue *JavaScriptValue = nil
	if len(arguments) == 0 {
		numberValue = NewNumberValue(0, false)
	} else {
		completion := ToNumeric(runtime, arguments[0])
		if completion.Type != Normal {
			return completion
		}
		numberValue = completion.Value.(*JavaScriptValue)

		if numberValue.Type == TypeBigInt {
			numberValue = NewNumberValue(float64(numberValue.Value.(*BigInt).Value.Int64()), false)
		}
	}

	if newTarget == nil || newTarget.Type == TypeUndefined {
		return NewNormalCompletion(numberValue)
	}

	newTargetObj := newTarget.Value.(FunctionInterface)
	completion := OrdinaryCreateFromConstructor(runtime, newTargetObj, IntrinsicNumberPrototype)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(*Object)
	object.NumberData = numberValue

	return NewNormalCompletion(objectVal)
}

func NewParseIntFunction(runtime *Runtime) ObjectInterface {
	parseIntFunc := CreateBuiltinFunction(
		runtime,
		NumberParseInt,
		2,
		NewStringValue("parseInt"),
		runtime.GetRunningRealm(),
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype),
	)
	return parseIntFunc
}

// TODO: This is not spec compliant.
func NumberParseInt(
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

	completion := ToString(runtime, arguments[0])
	if completion.Type != Normal {
		return completion
	}

	if arguments[1].Type != TypeUndefined {
		panic("TODO: Implement Number.parseInt with radix.")
	}

	strValue := completion.Value.(*JavaScriptValue).Value.(*String).Value
	value, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return NewNormalCompletion(NewNumberValue(0, true))
	}

	return NewNormalCompletion(NewNumberValue(value, false))
}
