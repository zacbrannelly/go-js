package runtime

import "strconv"

type ArrayObject struct {
	Prototype  ObjectInterface
	Properties map[string]PropertyDescriptor
	Extensible bool
}

func NewArrayObject(runtime *Runtime, length uint) *ArrayObject {
	obj := &ArrayObject{
		Prototype:  runtime.GetRunningRealm().Intrinsics[IntrinsicArrayPrototype],
		Properties: make(map[string]PropertyDescriptor),
		Extensible: true,
	}
	OrdinaryDefineOwnProperty(obj, NewStringValue("length"), &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(length), false),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	return obj
}

func ArrayCreate(runtime *Runtime, length uint) *Completion {
	if length > 2^32-1 {
		return NewThrowCompletion(NewRangeError("Array length too large"))
	}

	return NewNormalCompletion(NewArrayObject(runtime, length))
}

func ArraySetLength(array *ArrayObject, descriptor PropertyDescriptor) *Completion {
	lengthStr := NewStringValue("length")
	dataDescriptor := descriptor.(*DataPropertyDescriptor)
	if dataDescriptor.Value == nil {
		return OrdinaryDefineOwnProperty(array, lengthStr, descriptor)
	}

	newLenDescriptor := dataDescriptor.Copy().(*DataPropertyDescriptor)
	newLenCompletion := ToUint32(newLenDescriptor.Value)

	if newLenCompletion.Type != Normal {
		return newLenCompletion
	}

	newLen := newLenCompletion.Value.(*JavaScriptValue)

	numberLenCompletion := ToNumber(newLenDescriptor.Value)
	if numberLenCompletion.Type != Normal {
		return numberLenCompletion
	}

	numberLen := numberLenCompletion.Value.(*JavaScriptValue)

	completion := SameValueZero(newLen, numberLen)
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewRangeError("Invalid array length"))
	}

	newLenDescriptor.Value = newLen

	oldLenDescriptorCompletion := OrdinaryGetOwnProperty(array, lengthStr)
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
		return OrdinaryDefineOwnProperty(array, lengthStr, newLenDescriptor)
	}

	if !oldLenDescriptor.Writable {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	var newWritable bool = newLenDescriptor.Writable
	if !newWritable {
		newLenDescriptor.Writable = true
	}

	completion = OrdinaryDefineOwnProperty(array, lengthStr, newLenDescriptor)
	if completion.Type != Normal {
		return completion
	}

	successVal := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !successVal {
		return completion
	}

	// TODO: Remove elements if the new length is less than the current length.
	for i := uint(oldLen); i > uint(newLen.Value.(*Number).Value); i-- {
		deleteSuccessCompletion := array.Delete(NewStringValue(strconv.FormatInt(int64(i), 10)))
		if deleteSuccessCompletion.Type == Normal && !deleteSuccessCompletion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			newLenDescriptor.Value = NewNumberValue(float64(i+1), false)
			if !newWritable {
				newLenDescriptor.Writable = false
			}

			OrdinaryDefineOwnProperty(array, lengthStr, newLenDescriptor)
			return NewNormalCompletion(NewBooleanValue(false))
		}
	}

	if !newWritable {
		// TODO: This is a deviation from the spec.
		// The intent of the below is to set Writable to false by only providing the Writable field.
		// But we haven't implemented the merge logic yet, so we're just providing the entire descriptor.
		success := OrdinaryDefineOwnProperty(array, lengthStr, &DataPropertyDescriptor{
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

func (o *ArrayObject) DefineOwnProperty(key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	keyString := PropertyKeyToString(key)
	if keyString == "length" {
		return ArraySetLength(o, descriptor)
	}

	index, err := strconv.ParseUint(keyString, 10, 64)
	if err == nil && index <= 2^32-1 {
		lengthDescriptorCompletion := o.GetOwnProperty(NewStringValue("length"))
		if lengthDescriptorCompletion.Type != Normal {
			return lengthDescriptorCompletion
		}

		lengthDescriptor, ok := lengthDescriptorCompletion.Value.(*DataPropertyDescriptor)
		if !ok {
			panic("Assert failed: Length descriptor is not a data property descriptor.")
		}

		lengthCastCompletion := ToUint32(lengthDescriptor.Value)
		if lengthCastCompletion.Type != Normal {
			return lengthCastCompletion
		}

		length := uint32(lengthCastCompletion.Value.(*JavaScriptValue).Value.(*Number).Value)

		if uint32(index) >= length && !lengthDescriptor.GetWritable() {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		completion := OrdinaryDefineOwnProperty(o, key, descriptor)
		if completion.Type != Normal {
			return completion
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return completion
		}

		if uint32(index) >= length {
			lengthDescriptor.Value = NewNumberValue(float64(index+1), false)
			completion = OrdinaryDefineOwnProperty(o, NewStringValue("length"), lengthDescriptor)
			if completion.Type != Normal {
				return completion
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				panic("Assert failed: Length descriptor is not writable.")
			}
		}

		return NewNormalCompletion(NewBooleanValue(true))
	}

	return OrdinaryDefineOwnProperty(o, key, descriptor)
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

func (o *ArrayObject) GetExtensible() bool {
	return o.Extensible
}

func (o *ArrayObject) SetExtensible(extensible bool) {
	o.Extensible = extensible
}

func (o *ArrayObject) GetPrototypeOf() *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *ArrayObject) SetPrototypeOf(prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(o, prototype)
}

func (o *ArrayObject) GetOwnProperty(key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(o, key)
}

func (o *ArrayObject) HasProperty(key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(o, key)
}

func (o *ArrayObject) Set(key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(o, key, value, receiver)
}

func (o *ArrayObject) Get(key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(o, key, receiver)
}

func (o *ArrayObject) Delete(key *JavaScriptValue) *Completion {
	return OrdinaryDelete(o, key)
}
