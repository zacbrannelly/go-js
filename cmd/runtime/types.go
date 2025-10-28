package runtime

import (
	"fmt"
	"strings"
)

type JavaScriptType int

const (
	TypeUndefined JavaScriptType = iota
	TypeNull
	TypeSymbol
	TypeString
	TypeObject
	TypeNumber
	TypeBigInt
	TypeBoolean
	TypeReference
	TypePropertyDescriptor
)

var TypeNames = map[JavaScriptType]string{
	TypeUndefined:          "undefined",
	TypeNull:               "null",
	TypeSymbol:             "symbol",
	TypeString:             "string",
	TypeObject:             "object",
	TypeNumber:             "number",
	TypeBigInt:             "bigint",
	TypeBoolean:            "boolean",
	TypeReference:          "reference",
	TypePropertyDescriptor: "property descriptor",
}

type JavaScriptValue struct {
	Type  JavaScriptType
	Value any
}

func ObjectToString(v *JavaScriptValue) string {
	object := v.Value.(ObjectInterface)
	properties := []string{}
	for key, value := range object.GetProperties() {
		if dataDescriptor, ok := value.(*DataPropertyDescriptor); ok {
			properties = append(properties, fmt.Sprintf("%s: %s", key, dataDescriptor.Value.ToString()))
		} else {
			// TODO: Support accessor property descriptors.
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(properties, ", "))
}

func (v *JavaScriptValue) ToString() string {
	switch v.Type {
	case TypeString:
		return fmt.Sprintf("'%s'", v.Value.(*String).Value)
	case TypeSymbol:
		return fmt.Sprintf("Symbol(%s)", v.Value.(*Symbol).Name)
	case TypeNumber:
		return fmt.Sprintf("%f", v.Value.(*Number).Value)
	case TypeBoolean:
		return fmt.Sprintf("%t", v.Value.(*Boolean).Value)
	case TypeNull:
		return "null"
	case TypeUndefined:
		return "undefined"
	case TypeObject:
		return ObjectToString(v)
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

func NewNullValue() *JavaScriptValue {
	return NewJavaScriptValue(TypeNull, nil)
}
