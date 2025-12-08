package runtime

type Symbol struct {
	Description string
}

func NewSymbolValue(description string) *JavaScriptValue {
	return NewJavaScriptValue(TypeSymbol, &Symbol{
		Description: description,
	})
}
