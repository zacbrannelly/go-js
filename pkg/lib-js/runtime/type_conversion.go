package runtime

import (
	"math"
	"strconv"
)

func ToNumeric(value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		panic("TODO: ToNumeric for Object values is not implemented.")
	}

	if value.Type == TypeBigInt {
		return NewNormalCompletion(value)
	}

	return ToNumber(value)
}

func ToNumber(value *JavaScriptValue) *Completion {
	if value.Type == TypeNumber {
		return NewNormalCompletion(value)
	}

	if value.Type == TypeUndefined {
		// undefined -> NaN
		return NewNormalCompletion(NewNumberValue(0, true))
	}

	if value.Type == TypeString {
		// TODO: Implement parser for StringNumericLiteral.
		number, err := strconv.ParseFloat(value.Value.(*String).Value, 64)
		if err != nil {
			// NaN
			return NewNormalCompletion(NewNumberValue(0, true))
		}

		return NewNormalCompletion(NewNumberValue(number, false))
	}

	panic("TODO: ToNumber for non-Number values is not implemented.")
}

func ToString(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeString {
		return NewNormalCompletion(value)
	}

	if value.Type == TypeUndefined {
		return NewNormalCompletion(NewStringValue("undefined"))
	}

	if value.Type == TypeNull {
		return NewNormalCompletion(NewStringValue("null"))
	}

	if value.Type == TypeNumber {
		return NewNormalCompletion(NumberToString(value.Value.(*Number), 10))
	}

	if value.Type == TypeObject {
		return ToPrimitiveWithPreferredType(runtime, value, PreferredTypeString)
	}

	panic("TODO: ToString for non-String values is not implemented.")
}

type PreferredType int

const (
	PreferredTypeUndefined PreferredType = iota
	PreferredTypeNumber
	PreferredTypeString
)

func ToPrimitive(runtime *Runtime, value *JavaScriptValue) *Completion {
	return ToPrimitiveWithPreferredType(runtime, value, PreferredTypeUndefined)
}

func ToPrimitiveWithPreferredType(runtime *Runtime, value *JavaScriptValue, preferredType PreferredType) *Completion {
	if value.Type == TypeObject {
		completion := GetMethod(runtime, value, runtime.SymbolToPrimitive)
		if completion.Type != Normal {
			return completion
		}

		method := completion.Value.(*JavaScriptValue)
		if method.Type != TypeUndefined {
			panic("TODO: ToPrimitive for Object values with Symbol.toPrimitive is not implemented.")
		}

		if preferredType == PreferredTypeUndefined {
			preferredType = PreferredTypeNumber
		}

		return OrdinaryToPrimitive(runtime, value, preferredType)
	}

	return NewNormalCompletion(value)
}

func OrdinaryToPrimitive(runtime *Runtime, value *JavaScriptValue, hint PreferredType) *Completion {
	object := value.Value.(ObjectInterface)

	var methodNames []string
	if hint == PreferredTypeNumber {
		methodNames = []string{"valueOf", "toString"}
	} else {
		methodNames = []string{"toString", "valueOf"}
	}

	for _, methodName := range methodNames {
		methodKey := NewStringValue(methodName)
		completion := object.Get(runtime, methodKey, value)
		if completion.Type != Normal {
			return completion
		}

		method := completion.Value.(*JavaScriptValue)
		if IsCallable(method) {
			completion := Call(runtime, method, value, []*JavaScriptValue{})
			if completion.Type != Normal {
				return completion
			}

			result := completion.Value.(*JavaScriptValue)
			if result.Type != TypeObject {
				return completion
			}
		}
	}

	return NewThrowCompletion(NewTypeError(runtime, "Cannot convert object to primitive value."))
}

func ToBoolean(value *JavaScriptValue) *Completion {
	if value.Type == TypeBoolean {
		return NewNormalCompletion(value)
	}

	// Null and undefined are falsy.
	if value.Type == TypeNull || value.Type == TypeUndefined {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	// 0 is falsy.
	if value.Type == TypeNumber {
		return NewNormalCompletion(NewBooleanValue(value.Value.(*Number).Value != 0))
	}

	// Empty string is falsy.
	if value.Type == TypeString {
		return NewNormalCompletion(NewBooleanValue(value.Value.(*String).Value != ""))
	}

	// Otherwise, true.
	return NewNormalCompletion(NewBooleanValue(true))
}

func ToObject(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		return NewNormalCompletion(value)
	}

	if value.Type == TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot convert undefined to an object"))
	}

	if value.Type == TypeNull {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot convert null to an object"))
	}

	if value.Type == TypeString {
		proto := runtime.GetRunningRealm().GetIntrinsic(IntrinsicStringPrototype)
		stringObj := StringCreate(runtime, value, proto)
		return NewNormalCompletion(NewJavaScriptValue(TypeObject, stringObj))
	}

	panic("TODO: ToObject for non-Object values is not implemented.")
}

func ToUint32(value *JavaScriptValue) *Completion {
	numberCompletion := ToNumber(value)
	if numberCompletion.Type != Normal {
		return numberCompletion
	}

	finalValue := math.Trunc(numberCompletion.Value.(*JavaScriptValue).Value.(*Number).Value)
	finalValueUint64 := uint64(finalValue) % uint64(math.Pow(2, 32))

	return NewNormalCompletion(NewNumberValue(float64(finalValueUint64), false))
}

func ToLength(value *JavaScriptValue) *Completion {
	lenCompletion := ToIntegerOrInfinity(value)
	if lenCompletion.Type != Normal {
		return lenCompletion
	}

	len := lenCompletion.Value.(*JavaScriptValue).Value.(*Number).Value
	if len <= 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	return NewNormalCompletion(NewNumberValue(math.Min(len, math.Pow(2, 53)-1), false))
}

func ToIntegerOrInfinity(value *JavaScriptValue) *Completion {
	numberCompletion := ToNumber(value)
	if numberCompletion.Type != Normal {
		return numberCompletion
	}

	number := numberCompletion.Value.(*JavaScriptValue).Value.(*Number)
	if number.NaN || number.Value == 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	if number.Value == math.Inf(1) {
		return NewNormalCompletion(NewNumberValue(math.Inf(1), false))
	}

	if number.Value == math.Inf(-1) {
		return NewNormalCompletion(NewNumberValue(math.Inf(-1), false))
	}

	return NewNormalCompletion(NewNumberValue(truncate(number.Value), false))
}

func truncate(value float64) float64 {
	if value < 0 {
		return -math.Floor(-value)
	}
	return math.Floor(value)
}

func CanonicalNumericIndexString(runtime *Runtime, value *JavaScriptValue) *JavaScriptValue {
	if value.Type != TypeString {
		panic("Assert failed: CanonicalNumericIndexString value is not a string.")
	}

	valueString := value.Value.(*String).Value

	if valueString == "-0" {
		return NewNumberValue(-0, false)
	}

	completion := ToNumber(value)
	if completion.Type != Normal {
		panic("Assert failed: CanonicalNumericIndexString ToNumber threw an error.")
	}

	numberVal := completion.Value.(*JavaScriptValue)

	completion = ToString(runtime, numberVal)
	if completion.Type != Normal {
		panic("Assert failed: CanonicalNumericIndexString ToString threw an error.")
	}

	toString := completion.Value.(*JavaScriptValue).Value.(*String).Value

	if toString == valueString {
		return numberVal
	}

	return NewUndefinedValue()
}
