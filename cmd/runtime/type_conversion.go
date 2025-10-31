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

func ToObject(value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		return NewNormalCompletion(value)
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
