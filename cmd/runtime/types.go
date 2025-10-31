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

func ObjectToString(v *JavaScriptValue) (string, error) {
	object := v.Value.(ObjectInterface)
	properties := []string{}
	for key, value := range object.GetProperties() {
		if dataDescriptor, ok := value.(*DataPropertyDescriptor); ok {
			valueString, err := dataDescriptor.Value.ToString()
			if err != nil {
				return "error", err
			}
			properties = append(properties, fmt.Sprintf("%s: %s", key, valueString))
		} else {
			// TODO: Support accessor property descriptors.
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(properties, ", ")), nil
}

func ReferenceToString(v *JavaScriptValue) (string, error) {
	referenceVal := GetValue(v)
	if referenceVal.Type != Normal {
		return "error", referenceVal.Value.(error)
	}
	return referenceVal.Value.(*JavaScriptValue).ToString()
}

func (v *JavaScriptValue) ToString() (string, error) {
	switch v.Type {
	case TypeString:
		return fmt.Sprintf("'%s'", v.Value.(*String).Value), nil
	case TypeSymbol:
		return fmt.Sprintf("Symbol(%s)", v.Value.(*Symbol).Name), nil
	case TypeNumber:
		return fmt.Sprintf("%f", v.Value.(*Number).Value), nil
	case TypeBoolean:
		return fmt.Sprintf("%t", v.Value.(*Boolean).Value), nil
	case TypeNull:
		return "null", nil
	case TypeUndefined:
		return "undefined", nil
	case TypeObject:
		return ObjectToString(v)
	case TypeReference:
		return ReferenceToString(v)
	default:
		return "unknown", nil
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
