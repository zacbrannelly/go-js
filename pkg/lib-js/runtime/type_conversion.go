package runtime

import "math"

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

	panic("TODO: ToNumber for non-Number values is not implemented.")
}

func ToString(value *JavaScriptValue) *Completion {
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

	panic("TODO: ToString for non-String values is not implemented.")
}

func ToPrimitive(value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		panic("TODO: ToPrimitive for Object values is not implemented.")
	}

	return NewNormalCompletion(value)
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

	panic("TODO: ToObject for non-Object values is not implemented.")
}

func ToUint32(value *JavaScriptValue) *Completion {
	numberCompletion := ToNumber(value)
	if numberCompletion.Type != Normal {
		return numberCompletion
	}

	finalValue := math.Trunc(numberCompletion.Value.(*JavaScriptValue).Value.(*Number).Value)
	finalValueUint64 := uint64(finalValue) % (2 ^ 32)

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
