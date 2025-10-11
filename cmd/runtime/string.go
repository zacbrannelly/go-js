package runtime

type String struct {
	Value string
}

func NewStringValue(value string) *JavaScriptValue {
	return NewJavaScriptValue(TypeString, &String{
		Value: value,
	})
}

func StringAdd(left *String, right *String) *String {
	return &String{
		Value: left.Value + right.Value,
	}
}
