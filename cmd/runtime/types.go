package runtime

import "fmt"

type JavaScriptType int

const (
	TypeUndefined JavaScriptType = iota
	TypeSymbol
	TypeString
	TypeObject
	TypeNumber
	TypeBigInt
	TypeReference
	TypePropertyDescriptor
)

var TypeNames = map[JavaScriptType]string{
	TypeUndefined:          "undefined",
	TypeSymbol:             "symbol",
	TypeString:             "string",
	TypeObject:             "object",
	TypeNumber:             "number",
	TypeBigInt:             "bigint",
	TypeReference:          "reference",
	TypePropertyDescriptor: "property descriptor",
}

type JavaScriptValue struct {
	Type  JavaScriptType
	Value any
}

func (v *JavaScriptValue) ToString() string {
	switch v.Type {
	case TypeString:
		return v.Value.(*String).Value
	case TypeSymbol:
		return v.Value.(*Symbol).Name
	case TypeNumber:
		return fmt.Sprintf("%f", v.Value.(*Number).Value)
	default:
		return "unknown"
	}
}

func NewJavaScriptValue(valueType JavaScriptType, value any) *JavaScriptValue {
	return &JavaScriptValue{
		Type:  valueType,
		Value: value,
	}
}

func NewUndefinedValue() *JavaScriptValue {
	return NewJavaScriptValue(TypeUndefined, nil)
}
