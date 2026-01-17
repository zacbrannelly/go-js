package runtime

import (
	"math"
	"sort"
	"strconv"
	"unicode/utf8"
)

type StringObject struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool
	PrivateElements  []*PrivateElement

	StringData *JavaScriptValue
}

func StringCreate(runtime *Runtime, value *JavaScriptValue, prototype ObjectInterface) *StringObject {
	if value.Type != TypeString {
		panic("Assert failed: StringCreate value is not a string.")
	}

	stringObject := &StringObject{
		Prototype:        prototype,
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		Extensible:       true,
		PrivateElements:  make([]*PrivateElement, 0),
		StringData:       value,
	}

	length := utf8.RuneCountInString(value.Value.(*String).Value)
	DefinePropertyOrThrow(runtime, stringObject, lengthStr, &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(length), false),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	return stringObject
}

func (o *StringObject) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *StringObject) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *StringObject) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *StringObject) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *StringObject) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *StringObject) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
}

func (o *StringObject) IsExtensible(runtime *Runtime) *Completion {
	return NewNormalCompletion(NewBooleanValue(o.Extensible))
}

func (o *StringObject) GetPrototypeOf(runtime *Runtime) *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *StringObject) SetPrototypeOf(runtime *Runtime, prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(runtime, o, prototype)
}

func (o *StringObject) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	completion := OrdinaryGetOwnProperty(runtime, o, key)

	if completion.Value != nil {
		return completion
	}

	return StringGetOwnProperty(runtime, o, key)
}

func (o *StringObject) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *StringObject) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	completion := StringGetOwnProperty(runtime, o, key)
	if completion.Type != Normal {
		return completion
	}

	if completion.Value != nil {
		stringDesc := completion.Value.(PropertyDescriptor)
		return NewNormalCompletion(IsCompatiblePropertyDescriptor(o.Extensible, descriptor, stringDesc))
	}

	return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
}

func (o *StringObject) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *StringObject) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *StringObject) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryDelete(runtime, o, key)
}

func (o *StringObject) OwnPropertyKeys(runtime *Runtime) *Completion {
	keys := make([]*JavaScriptValue, 0)

	stringData := o.StringData.Value.(*String).Value

	for i := 0; i < len(stringData); i++ {
		keys = append(keys, NewStringValue(strconv.Itoa(i)))
	}

	arrayIndexKeys := make([]*JavaScriptValue, 0)
	arrayIndices := make([]int, 0)
	stringKeys := make([]*JavaScriptValue, 0)

	for key := range o.Properties {
		arrayKey, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			stringKeys = append(stringKeys, NewStringValue(key))
			continue
		}

		if int(arrayKey) < len(stringData) {
			continue
		}

		if arrayKey < 0 && float64(arrayKey) >= math.Pow(2, 32)-1 {
			continue
		}

		arrayIndexKeys = append(arrayIndexKeys, NewStringValue(key))
		arrayIndices = append(arrayIndices, int(arrayKey))
	}

	sort.Slice(arrayIndexKeys, func(i, j int) bool {
		return arrayIndices[i] < arrayIndices[j]
	})

	keys = append(keys, arrayIndexKeys...)
	keys = append(keys, stringKeys...)

	for key := range o.SymbolProperties {
		keys = append(keys, NewJavaScriptValue(TypeSymbol, key))
	}

	return NewNormalCompletion(keys)
}

func (o *StringObject) PreventExtensions(runtime *Runtime) *Completion {
	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *StringObject) GetPrivateElements() []*PrivateElement {
	return o.PrivateElements
}

func (o *StringObject) SetPrivateElements(privateElements []*PrivateElement) {
	o.PrivateElements = privateElements
}

func StringGetOwnProperty(runtime *Runtime, object *StringObject, key *JavaScriptValue) *Completion {
	if key.Type != TypeString {
		// Nil to signal undefined.
		// TODO: Convert these to use JavaScriptValue.
		return NewNormalCompletion(nil)
	}

	index := CanonicalNumericIndexString(runtime, key)
	if index.Type != TypeNumber {
		// Nil to signal undefined.
		return NewNormalCompletion(nil)
	}

	indexNumber := index.Value.(*Number)
	indexValue := indexNumber.Value

	if indexNumber.NaN || math.Floor(indexValue) != indexValue {
		// Nil to signal undefined.
		return NewNormalCompletion(nil)
	}

	if indexValue < 0 {
		// Nil to signal undefined.
		return NewNormalCompletion(nil)
	}

	if object.StringData == nil || object.StringData.Type != TypeString {
		panic("Assert failed: StringObject.StringData is not a string.")
	}

	stringData := object.StringData.Value.(*String).Value
	if int(indexValue) >= utf8.RuneCountInString(stringData) {
		// Nil to signal undefined.
		return NewNormalCompletion(nil)
	}

	runeVal := []rune(stringData)[int(indexValue)]

	return NewNormalCompletion(&DataPropertyDescriptor{
		Value:        NewStringValue(string(runeVal)),
		Writable:     false,
		Enumerable:   true,
		Configurable: false,
	})
}
