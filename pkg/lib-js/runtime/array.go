package runtime

import (
	"math"
	"strconv"
)

var (
	constructorString = NewStringValue("constructor")
)

type ArrayObject struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool
	PrivateElements  []*PrivateElement
}

func NewArrayObject(runtime *Runtime, length uint) *ArrayObject {
	obj := &ArrayObject{
		Prototype:        runtime.GetRunningRealm().GetIntrinsic(IntrinsicArrayPrototype),
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		Extensible:       true,
		PrivateElements:  make([]*PrivateElement, 0),
	}
	OrdinaryDefineOwnProperty(runtime, obj, NewStringValue("length"), &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(length), false),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return obj
}

func ArrayCreate(runtime *Runtime, length uint) *Completion {
	if float64(length) > math.Pow(2, 32)-1 {
		return NewThrowCompletion(NewRangeError(runtime, "Array length too large"))
	}

	arrayObject := NewArrayObject(runtime, length)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, arrayObject))
}

func ArrayCreateWithPrototype(runtime *Runtime, length uint, prototype ObjectInterface) *Completion {
	if float64(length) > math.Pow(2, 32)-1 {
		return NewThrowCompletion(NewRangeError(runtime, "Array length too large"))
	}

	arrayObject := NewArrayObject(runtime, length)
	arrayObject.Prototype = prototype
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, arrayObject))
}

func ArraySpeciesCreate(runtime *Runtime, originalArray *JavaScriptValue, length uint) *Completion {
	completion := IsArray(runtime, originalArray)
	if completion.Type != Normal {
		return completion
	}

	isArray := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !isArray {
		return ArrayCreate(runtime, length)
	}

	object := originalArray.Value.(ObjectInterface)

	completion = object.Get(runtime, constructorString, originalArray)
	if completion.Type != Normal {
		return completion
	}

	constructor := completion.Value.(*JavaScriptValue)

	if constructorObj, ok := constructor.Value.(FunctionInterface); ok && constructorObj.HasConstructMethod() {
		thisRealm := runtime.GetRunningRealm()
		completion = GetFunctionRealm(runtime, constructorObj)
		if completion.Type != Normal {
			return completion
		}

		realm := completion.Value.(*Realm)

		if thisRealm != realm {
			intrinsicConstructor := NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicArrayConstructor))
			completion = SameValue(constructor, intrinsicConstructor)
			if completion.Type != Normal {
				return completion
			}

			if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				constructor = NewUndefinedValue()
			}
		}
	}

	if constructor.Type == TypeObject {
		constructorObj := constructor.Value.(ObjectInterface)

		completion = constructorObj.Get(runtime, runtime.SymbolSpecies, constructor)
		if completion.Type != Normal {
			return completion
		}

		constructor = completion.Value.(*JavaScriptValue)
		if constructor.Type == TypeNull {
			constructor = NewUndefinedValue()
		}
	}

	if constructor.Type == TypeUndefined {
		return ArrayCreate(runtime, length)
	}

	constructorObj, ok := constructor.Value.(FunctionInterface)
	if !ok || !constructorObj.HasConstructMethod() {
		return NewThrowCompletion(NewTypeError(runtime, "Array species constructor is not a constructor"))
	}

	lengthVal := NewNumberValue(float64(length), false)
	return Construct(runtime, constructorObj, []*JavaScriptValue{lengthVal}, nil)
}

func ArraySetLength(runtime *Runtime, array *ArrayObject, descriptor PropertyDescriptor) *Completion {
	lengthStr := NewStringValue("length")
	dataDescriptor := descriptor.(*DataPropertyDescriptor)
	if dataDescriptor.Value == nil {
		return OrdinaryDefineOwnProperty(runtime, array, lengthStr, descriptor)
	}

	newLenDescriptor := dataDescriptor.Copy().(*DataPropertyDescriptor)
	newLenCompletion := ToUint32(runtime, newLenDescriptor.Value)

	if newLenCompletion.Type != Normal {
		return newLenCompletion
	}

	newLen := newLenCompletion.Value.(*JavaScriptValue)

	numberLenCompletion := ToNumber(runtime, newLenDescriptor.Value)
	if numberLenCompletion.Type != Normal {
		return numberLenCompletion
	}

	numberLen := numberLenCompletion.Value.(*JavaScriptValue)

	completion := SameValueZero(newLen, numberLen)
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewRangeError(runtime, "Invalid array length"))
	}

	newLenDescriptor.Value = newLen

	oldLenDescriptorCompletion := OrdinaryGetOwnProperty(runtime, array, lengthStr)
	if oldLenDescriptorCompletion.Type != Normal {
		return oldLenDescriptorCompletion
	}

	if oldLenDescriptorCompletion.Value == nil {
		panic("Assert failed: Length descriptor is undefined inside of ArraySetLength.")
	}

	oldLenDescriptor := oldLenDescriptorCompletion.Value.(*DataPropertyDescriptor)

	oldLen := oldLenDescriptor.Value.Value.(*Number).Value

	// If extending the array, just define the new length.
	if newLen.Value.(*Number).Value >= oldLen {
		return OrdinaryDefineOwnProperty(runtime, array, lengthStr, newLenDescriptor)
	}

	if !oldLenDescriptor.Writable {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	var newWritable bool = newLenDescriptor.Writable
	if !newWritable {
		newLenDescriptor.Writable = true
	}

	completion = OrdinaryDefineOwnProperty(runtime, array, lengthStr, newLenDescriptor)
	if completion.Type != Normal {
		return completion
	}

	successVal := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !successVal {
		return completion
	}

	// TODO: Remove elements if the new length is less than the current length.
	for i := uint(oldLen); i > uint(newLen.Value.(*Number).Value); i-- {
		deleteSuccessCompletion := array.Delete(runtime, NewStringValue(strconv.FormatInt(int64(i), 10)))
		if deleteSuccessCompletion.Type == Normal && !deleteSuccessCompletion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			newLenDescriptor.Value = NewNumberValue(float64(i+1), false)
			if !newWritable {
				newLenDescriptor.Writable = false
			}

			OrdinaryDefineOwnProperty(runtime, array, lengthStr, newLenDescriptor)
			return NewNormalCompletion(NewBooleanValue(false))
		}
	}

	if !newWritable {
		// TODO: This is a deviation from the spec.
		// The intent of the below is to set Writable to false by only providing the Writable field.
		// But we haven't implemented the merge logic yet, so we're just providing the entire descriptor.
		success := OrdinaryDefineOwnProperty(runtime, array, lengthStr, &DataPropertyDescriptor{
			Writable:     false,
			Enumerable:   newLenDescriptor.Enumerable,
			Configurable: newLenDescriptor.Configurable,
			Value:        newLenDescriptor.Value,
		})
		if success.Type == Normal && !success.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: Setting length descriptor failed when it shouldn't have.")
		}
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func CreateArrayFromList(runtime *Runtime, list []*JavaScriptValue) ObjectInterface {
	array := NewArrayObject(runtime, 0)
	for i, value := range list {
		completion := CreateDataProperty(runtime, array, NewStringValue(strconv.FormatInt(int64(i), 10)), value)
		if completion.Type != Normal || !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: CreateArrayFromList CreateDataProperty threw an unexpected error.")
		}
	}

	return array
}

func (o *ArrayObject) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	if key.Type == TypeSymbol {
		return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
	}

	if key.Type != TypeString {
		panic("Assert failed: ArrayObject DefineOwnProperty key is not a string or symbol.")
	}

	keyString := key.Value.(*String).Value
	if keyString == "length" {
		return ArraySetLength(runtime, o, descriptor)
	}

	index, err := strconv.ParseUint(keyString, 10, 64)
	if err == nil && float64(index) <= math.Pow(2, 32)-1 {
		lengthDescriptorCompletion := o.GetOwnProperty(runtime, NewStringValue("length"))
		if lengthDescriptorCompletion.Type != Normal {
			return lengthDescriptorCompletion
		}

		lengthDescriptor, ok := lengthDescriptorCompletion.Value.(*DataPropertyDescriptor)
		if !ok {
			panic("Assert failed: Length descriptor is not a data property descriptor.")
		}

		lengthCastCompletion := ToUint32(runtime, lengthDescriptor.Value)
		if lengthCastCompletion.Type != Normal {
			return lengthCastCompletion
		}

		length := uint32(lengthCastCompletion.Value.(*JavaScriptValue).Value.(*Number).Value)

		if uint32(index) >= length && !lengthDescriptor.GetWritable() {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		completion := OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
		if completion.Type != Normal {
			return completion
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return completion
		}

		if uint32(index) >= length {
			lengthDescriptor.Value = NewNumberValue(float64(index+1), false)
			completion = OrdinaryDefineOwnProperty(runtime, o, NewStringValue("length"), lengthDescriptor)
			if completion.Type != Normal {
				return completion
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				panic("Assert failed: Length descriptor is not writable.")
			}
		}

		return NewNormalCompletion(NewBooleanValue(true))
	}

	return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
}

func (o *ArrayObject) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *ArrayObject) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *ArrayObject) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *ArrayObject) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *ArrayObject) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *ArrayObject) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
}

func (o *ArrayObject) IsExtensible(runtime *Runtime) *Completion {
	return NewNormalCompletion(NewBooleanValue(o.Extensible))
}

func (o *ArrayObject) GetPrototypeOf(runtime *Runtime) *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *ArrayObject) SetPrototypeOf(runtime *Runtime, prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(runtime, o, prototype)
}

func (o *ArrayObject) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(runtime, o, key)
}

func (o *ArrayObject) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *ArrayObject) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *ArrayObject) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *ArrayObject) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryDelete(runtime, o, key)
}

func (o *ArrayObject) OwnPropertyKeys(runtime *Runtime) *Completion {
	return NewNormalCompletion(OrdinaryOwnPropertyKeys(o))
}

func (o *ArrayObject) PreventExtensions(runtime *Runtime) *Completion {
	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *ArrayObject) GetPrivateElements() []*PrivateElement {
	return o.PrivateElements
}

func (o *ArrayObject) SetPrivateElements(privateElements []*PrivateElement) {
	o.PrivateElements = privateElements
}

func (o *ArrayObject) GetLength() int {
	desc, ok := o.Properties["length"]
	if !ok {
		panic("Assert failed: Length property is not defined.")
	}

	dataDesc, ok := desc.(*DataPropertyDescriptor)
	if !ok {
		panic("Assert failed: Length property is not a data property descriptor in ArrayObject GetLength.")
	}

	numberVal, ok := dataDesc.Value.Value.(*Number)
	if !ok {
		panic("Assert failed: Length property is not a number in ArrayObject GetLength.")
	}

	return int(numberVal.Value)
}
