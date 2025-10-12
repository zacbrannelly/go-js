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
