package runtime

type Symbol struct {
	Name        string
	Description string
}

func NewSymbolValue(name string, description string) *JavaScriptValue {
	return NewJavaScriptValue(TypeSymbol, &Symbol{
		Name:        name,
		Description: description,
	})
}
