package runtime

func ToNumeric(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		panic("TODO: ToNumeric for Object values is not implemented.")
	}

	if value.Type == TypeBigInt {
		return NewNormalCompletion(value)
	}

	return ToNumber(runtime, value)
}

func ToNumber(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeNumber {
		return NewNormalCompletion(value)
	}

	panic("TODO: ToNumber for non-Number values is not implemented.")
}

func ToString(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeString {
		return NewNormalCompletion(value)
	}

	panic("TODO: ToString for non-String values is not implemented.")
}

func ToPrimitive(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		panic("TODO: ToPrimitive for Object values is not implemented.")
	}

	return NewNormalCompletion(value)
}

func ToBoolean(runtime *Runtime, value *JavaScriptValue) *Completion {
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
