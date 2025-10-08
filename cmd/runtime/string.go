package runtime

type String struct {
	Value string
}

func NewStringValue(value string) *JavaScriptValue {
	return NewJavaScriptValue(TypeString, &String{
		Value: value,
	})
}
