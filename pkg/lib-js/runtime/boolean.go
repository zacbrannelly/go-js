package runtime

type Boolean struct {
	Value bool
}

func NewBooleanValue(value bool) *JavaScriptValue {
	return NewJavaScriptValue(TypeBoolean, &Boolean{
		Value: value,
	})
}
