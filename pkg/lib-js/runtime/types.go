package runtime

import (
	"fmt"
	"strconv"
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
	TypePrivateName
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
	TypePrivateName:        "private name",
}

type JavaScriptValue struct {
	Type  JavaScriptType
	Value any
}

func ArrayToString(runtime *Runtime, v *JavaScriptValue) (string, error) {
	array := v.Value.(*ArrayObject)
	length := array.GetLength()
	elements := []string{}
	for i := range length {
		element := array.Get(runtime, NewStringValue(strconv.Itoa(i)), v)
		if element.Type != Normal {
			return "error", element.Value.(error)
		}

		elementVal := element.Value.(*JavaScriptValue)
		elementStr, err := elementVal.ToString(runtime)
		if err != nil {
			return "error", err
		}

		elements = append(elements, elementStr)
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", ")), nil
}

func ObjectToString(runtime *Runtime, v *JavaScriptValue) (string, error) {
	if _, ok := v.Value.(*ArrayObject); ok {
		return ArrayToString(runtime, v)
	}

	object := v.Value.(ObjectInterface)
	properties := []string{}

	propertyToString := func(key string, value PropertyDescriptor) error {
		// Skip the constructor property to avoid infinite recursion.
		if key == "constructor" {
			return nil
		}

		if dataDescriptor, ok := value.(*DataPropertyDescriptor); ok {
			valueString, err := dataDescriptor.Value.ToString(runtime)
			if err != nil {
				return err
			}
			properties = append(properties, fmt.Sprintf("%s: %s", key, valueString))
		} else {
			// TODO: Support accessor property descriptors.
		}

		return nil
	}

	for key, value := range object.GetProperties() {
		err := propertyToString(key, value)
		if err != nil {
			return "error", err
		}
	}

	for key, value := range object.GetSymbolProperties() {
		err := propertyToString(key.Description, value)
		if err != nil {
			return "error", err
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(properties, ", ")), nil
}

func ReferenceToString(runtime *Runtime, v *JavaScriptValue) (string, error) {
	referenceVal := GetValue(runtime, v)
	if referenceVal.Type != Normal {
		return "error", referenceVal.Value.(error)
	}
	return referenceVal.Value.(*JavaScriptValue).ToString(runtime)
}

func (v *JavaScriptValue) ToString(runtime *Runtime) (string, error) {
	switch v.Type {
	case TypeString:
		return fmt.Sprintf("'%s'", v.Value.(*String).Value), nil
	case TypeSymbol:
		return fmt.Sprintf("Symbol(%s)", v.Value.(*Symbol).Description), nil
	case TypeNumber:
		return fmt.Sprintf("%f", v.Value.(*Number).Value), nil
	case TypeBoolean:
		return fmt.Sprintf("%t", v.Value.(*Boolean).Value), nil
	case TypeNull:
		return "null", nil
	case TypeUndefined:
		return "undefined", nil
	case TypeObject:
		return ObjectToString(runtime, v)
	case TypeReference:
		return ReferenceToString(runtime, v)
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
