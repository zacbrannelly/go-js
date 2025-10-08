package runtime

type Number struct {
	Value float64
	NaN   bool
}

func NewNumberValue(value float64, nan bool) *JavaScriptValue {
	return NewJavaScriptValue(TypeNumber, &Number{
		Value: value,
		NaN:   nan,
	})
}

func NewNaNNumberValue() *JavaScriptValue {
	return NewNumberValue(0, true)
}

func NumberAdd(left *Number, right *Number) *Number {
	if left.NaN || right.NaN {
		return &Number{
			Value: 0,
			NaN:   true,
		}
	}

	return &Number{
		Value: left.Value + right.Value,
		NaN:   false,
	}
}
