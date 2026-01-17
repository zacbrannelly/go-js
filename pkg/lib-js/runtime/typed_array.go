package runtime

import (
	"math"
	"strconv"
)

type TypedArrayContentType int

const (
	TypedArrayContentTypeNumber TypedArrayContentType = iota
	TypedArrayContentTypeBigInt
)

type TypedArrayName string

const (
	TypedArrayNameInt8         TypedArrayName = "Int8Array"
	TypedArrayNameUint8        TypedArrayName = "Uint8Array"
	TypedArrayNameUint8Clamped TypedArrayName = "Uint8ClampedArray"
	TypedArrayNameInt16        TypedArrayName = "Int16Array"
	TypedArrayNameUint16       TypedArrayName = "Uint16Array"
	TypedArrayNameInt32        TypedArrayName = "Int32Array"
	TypedArrayNameUint32       TypedArrayName = "Uint32Array"
	TypedArrayNameBigInt64     TypedArrayName = "BigInt64Array"
	TypedArrayNameBigUint64    TypedArrayName = "BigUint64Array"
	TypedArrayNameFloat16      TypedArrayName = "Float16Array"
	TypedArrayNameFloat32      TypedArrayName = "Float32Array"
	TypedArrayNameFloat64      TypedArrayName = "Float64Array"
)

var TypedArrayElementSizes = map[TypedArrayName]uint{
	TypedArrayNameInt8:         1,
	TypedArrayNameUint8:        1,
	TypedArrayNameUint8Clamped: 1,
	TypedArrayNameInt16:        2,
	TypedArrayNameUint16:       2,
	TypedArrayNameInt32:        4,
	TypedArrayNameUint32:       4,
	TypedArrayNameBigInt64:     8,
	TypedArrayNameBigUint64:    8,
	TypedArrayNameFloat16:      2,
	TypedArrayNameFloat32:      4,
	TypedArrayNameFloat64:      8,
}

type TypedArrayConversionFunction func(runtime *Runtime, value *JavaScriptValue) *Completion

var TypedArrayConversionFunctions = map[TypedArrayName]TypedArrayConversionFunction{
	TypedArrayNameInt8:         ToInt8,
	TypedArrayNameUint8:        ToUint8,
	TypedArrayNameUint8Clamped: ToUint8Clamped,
	TypedArrayNameInt16:        ToInt16,
	TypedArrayNameUint16:       ToUint16,
	TypedArrayNameInt32:        ToInt32,
	TypedArrayNameUint32:       ToUint32,
}

type TypedArrayObject struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool
	PrivateElements  []*PrivateElement

	ViewedArrayBuffer *Object
	ArrayLengthAuto   bool
	ArrayLength       uint
	ByteOffset        uint
	ByteLength        uint
	ByteLengthAuto    bool
	ContentType       TypedArrayContentType
	TypedArrayName    TypedArrayName
}

func AllocateTypedArray(
	runtime *Runtime,
	typedArrayName TypedArrayName,
	newTarget FunctionInterface,
	defaultProto Intrinsic,
	length uint,
) *Completion {
	completion := GetPrototypeFromConstructor(runtime, newTarget, defaultProto)
	if completion.Type != Normal {
		return completion
	}

	prototype := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	obj := TypedArrayCreate(runtime, prototype)
	obj.TypedArrayName = typedArrayName

	if typedArrayName == TypedArrayNameBigInt64 || typedArrayName == TypedArrayNameBigUint64 {
		obj.ContentType = TypedArrayContentTypeBigInt
	} else {
		obj.ContentType = TypedArrayContentTypeNumber
	}

	completion = AllocateTypedArrayBuffer(runtime, obj, length)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, obj))
}

func AllocateTypedArrayWithNoBuffer(
	runtime *Runtime,
	typedArrayName TypedArrayName,
	newTarget FunctionInterface,
	defaultProto Intrinsic,
) *Completion {
	completion := GetPrototypeFromConstructor(runtime, newTarget, defaultProto)
	if completion.Type != Normal {
		return completion
	}

	prototype := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	obj := TypedArrayCreate(runtime, prototype)
	obj.TypedArrayName = typedArrayName

	if typedArrayName == TypedArrayNameBigInt64 || typedArrayName == TypedArrayNameBigUint64 {
		obj.ContentType = TypedArrayContentTypeBigInt
	} else {
		obj.ContentType = TypedArrayContentTypeNumber
	}

	obj.ArrayLength = 0
	obj.ByteLength = 0
	obj.ByteOffset = 0

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, obj))
}

func AllocateTypedArrayBuffer(runtime *Runtime, obj *TypedArrayObject, length uint) *Completion {
	elementSize := TypedArrayElementSize(obj)
	arrayBufferConstructor := runtime.GetRunningRealm().GetIntrinsic(IntrinsicArrayBufferConstructor).(FunctionInterface)
	byteLength := length * elementSize
	completion := AllocateArrayBuffer(runtime, arrayBufferConstructor, byteLength)
	if completion.Type != Normal {
		return completion
	}

	obj.ViewedArrayBuffer = completion.Value.(*JavaScriptValue).Value.(*Object)
	obj.ByteLength = byteLength
	obj.ByteOffset = 0
	obj.ArrayLength = length

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, obj))
}

func InitializeTypedArrayFromList(runtime *Runtime, obj *TypedArrayObject, values []*JavaScriptValue) *Completion {
	completion := AllocateTypedArrayBuffer(runtime, obj, uint(len(values)))
	if completion.Type != Normal {
		return completion
	}

	objVal := NewJavaScriptValue(TypeObject, obj)

	for i, value := range values {
		completion = ToString(runtime, NewNumberValue(float64(i), false))
		if completion.Type != Normal {
			return completion
		}

		key := completion.Value.(*JavaScriptValue)
		completion = obj.Set(runtime, key, value, objVal)
		if completion.Type != Normal {
			return completion
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError(runtime, "Failed to set property."))
		}
	}

	// Clear the values array - in place.
	// To be compliant with the spec.
	values = values[:0]

	return NewUnusedCompletion()
}

func InitializeTypedArrayFromArrayBuffer(
	runtime *Runtime,
	obj *TypedArrayObject,
	arrayBuffer *Object,
	byteOffset *JavaScriptValue,
	length *JavaScriptValue,
) *Completion {
	elementSize := TypedArrayElementSize(obj)
	completion := ToIndex(runtime, byteOffset)
	if completion.Type != Normal {
		return completion
	}

	offset := uint(completion.Value.(*JavaScriptValue).Value.(*Number).Value)

	if offset%elementSize != 0 {
		return NewThrowCompletion(NewRangeError(runtime, "Byte offset is not aligned to the element size"))
	}

	bufferIsFixedLength := IsFixedLengthArrayBuffer(arrayBuffer)

	if length.Type != TypeUndefined {
		completion = ToIndex(runtime, length)
		if completion.Type != Normal {
			return completion
		}

		length = completion.Value.(*JavaScriptValue)
	}

	if IsDetachedArrayBuffer(arrayBuffer) {
		return NewThrowCompletion(NewTypeError(runtime, "ArrayBuffer is detached"))
	}

	bufferByteLength := ArrayBufferByteLength(arrayBuffer, false)

	if length.Type == TypeUndefined && !bufferIsFixedLength {
		if uint(offset) > bufferByteLength {
			return NewThrowCompletion(NewRangeError(runtime, "Byte offset is out of bounds"))
		}

		obj.ArrayLengthAuto = true
		obj.ByteLengthAuto = true
	} else {
		var newByteLength int

		if length.Type == TypeUndefined {
			if bufferByteLength%elementSize != 0 {
				return NewThrowCompletion(NewRangeError(runtime, "Array length is not a multiple of the element size"))
			}

			newByteLength = int(bufferByteLength) - int(offset)
			if newByteLength < 0 {
				return NewThrowCompletion(NewRangeError(runtime, "Byte offset is out of bounds"))
			}
		} else {
			lengthVal := length.Value.(*Number).Value
			newByteLength = int(lengthVal) * int(elementSize)
			if int(offset)+newByteLength > int(bufferByteLength) {
				return NewThrowCompletion(NewRangeError(runtime, "Byte offset is out of bounds"))
			}
		}

		obj.ByteLength = uint(newByteLength)
		obj.ArrayLength = uint(newByteLength) / elementSize
	}

	obj.ViewedArrayBuffer = arrayBuffer
	obj.ByteOffset = offset

	return NewUnusedCompletion()
}

func TypedArrayCreate(runtime *Runtime, prototype ObjectInterface) *TypedArrayObject {
	return &TypedArrayObject{
		Prototype:         prototype,
		Properties:        make(map[string]PropertyDescriptor),
		SymbolProperties:  make(map[*Symbol]PropertyDescriptor),
		Extensible:        true,
		PrivateElements:   make([]*PrivateElement, 0),
		ViewedArrayBuffer: nil,
		ArrayLengthAuto:   false,
		ArrayLength:       0,
	}
}

func (o *TypedArrayObject) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *TypedArrayObject) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *TypedArrayObject) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *TypedArrayObject) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *TypedArrayObject) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *TypedArrayObject) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
}

func (o *TypedArrayObject) IsExtensible(runtime *Runtime) *Completion {
	return NewNormalCompletion(NewBooleanValue(o.Extensible))
}

func (o *TypedArrayObject) GetPrototypeOf(runtime *Runtime) *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *TypedArrayObject) SetPrototypeOf(runtime *Runtime, prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(runtime, o, prototype)
}

func (o *TypedArrayObject) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	if key.Type == TypeString {
		index := CanonicalNumericIndexString(runtime, key)
		if index.Type != TypeUndefined {
			value := TypedArrayGetElement(runtime, o, index.Value.(*Number))
			if value.Type == TypeUndefined {
				// Nil to signal undefined.
				return NewNormalCompletion(nil)
			}

			return NewNormalCompletion(&DataPropertyDescriptor{
				Value:        value,
				Writable:     true,
				Enumerable:   true,
				Configurable: true,
			})
		}
	}
	return OrdinaryGetOwnProperty(runtime, o, key)
}

func (o *TypedArrayObject) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	if key.Type == TypeString {
		index := CanonicalNumericIndexString(runtime, key)
		if index.Type != TypeUndefined {
			isValid := IsValidIntegerIndex(runtime, o, index.Value.(*Number))
			return NewNormalCompletion(NewBooleanValue(isValid))
		}
	}
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *TypedArrayObject) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	if key.Type == TypeString {
		index := CanonicalNumericIndexString(runtime, key)
		if index.Type != TypeUndefined {
			if !IsValidIntegerIndex(runtime, o, index.Value.(*Number)) {
				return NewNormalCompletion(NewBooleanValue(false))
			}

			if !descriptor.GetConfigurable() || !descriptor.GetEnumerable() {
				return NewNormalCompletion(NewBooleanValue(false))
			}

			if descriptor.GetType() != DataPropertyDescriptorType {
				return NewNormalCompletion(NewBooleanValue(false))
			}

			if !descriptor.(*DataPropertyDescriptor).Writable {
				return NewNormalCompletion(NewBooleanValue(false))
			}

			numericIndex := index.Value.(*Number)
			value := descriptor.(*DataPropertyDescriptor).Value

			completion := TypedArraySetElement(runtime, o, numericIndex, value)
			if completion.Type != Normal {
				return completion
			}

			return NewNormalCompletion(NewBooleanValue(true))
		}
	}

	return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
}

func (o *TypedArrayObject) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	if key.Type == TypeString {
		objVal := NewJavaScriptValue(TypeObject, o)
		index := CanonicalNumericIndexString(runtime, key)

		if index.Type != TypeUndefined {
			completion := SameValue(objVal, receiver)
			if completion.Type != Normal {
				return completion
			}

			isSame := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
			if isSame {
				completion = TypedArraySetElement(runtime, o, index.Value.(*Number), value)
				if completion.Type != Normal {
					return completion
				}

				return NewNormalCompletion(NewBooleanValue(true))
			}

			if !IsValidIntegerIndex(runtime, o, index.Value.(*Number)) {
				return NewNormalCompletion(NewBooleanValue(true))
			}
		}
	}
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *TypedArrayObject) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	if key.Type == TypeString {
		index := CanonicalNumericIndexString(runtime, key)
		if index.Type != TypeUndefined {
			return NewNormalCompletion(TypedArrayGetElement(runtime, o, index.Value.(*Number)))
		}
	}
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *TypedArrayObject) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	if key.Type == TypeString {
		index := CanonicalNumericIndexString(runtime, key)
		if index.Type != TypeUndefined {
			return NewNormalCompletion(NewBooleanValue(!IsValidIntegerIndex(runtime, o, index.Value.(*Number))))
		}
	}
	return OrdinaryDelete(runtime, o, key)
}

func (o *TypedArrayObject) OwnPropertyKeys(runtime *Runtime) *Completion {
	taRecord := MakeTypedArrayWithBufferWitness(o, false)
	keys := make([]*JavaScriptValue, 0)

	if !taRecord.IsTypedArrayOutOfBounds() {
		length := taRecord.TypedArrayLength()
		for i := range length {
			keys = append(keys, NewStringValue(strconv.Itoa(int(i))))
		}
	}

	for key := range o.Properties {
		keys = append(keys, NewStringValue(key))
	}

	for key := range o.SymbolProperties {
		keys = append(keys, NewJavaScriptValue(TypeSymbol, key))
	}

	return NewNormalCompletion(keys)
}

func (o *TypedArrayObject) PreventExtensions(runtime *Runtime) *Completion {
	if !IsTypedArrayFixedLength(o) {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *TypedArrayObject) GetPrivateElements() []*PrivateElement {
	return o.PrivateElements
}

func (o *TypedArrayObject) SetPrivateElements(privateElements []*PrivateElement) {
	o.PrivateElements = privateElements
}

func IsTypedArrayFixedLength(object *TypedArrayObject) bool {
	if object.ArrayLengthAuto {
		return false
	}

	if !IsFixedLengthArrayBuffer(object.ViewedArrayBuffer) && !IsSharedArrayBuffer(object.ViewedArrayBuffer) {
		return false
	}

	return true
}

func TypedArrayGetElement(runtime *Runtime, object *TypedArrayObject, index *Number) *JavaScriptValue {
	if !IsValidIntegerIndex(runtime, object, index) {
		return NewUndefinedValue()
	}

	elementSize := TypedArrayElementSize(object)
	byteIndexInBuffer := uint(index.Value)*elementSize + object.ByteOffset

	return GetValueFromBuffer(
		runtime,
		object.ViewedArrayBuffer,
		byteIndexInBuffer,
		object.TypedArrayName,
		true, /* isTypedArray */
		true, /* unordered */
	)
}

func TypedArraySetElement(
	runtime *Runtime,
	object *TypedArrayObject,
	index *Number,
	value *JavaScriptValue,
) *Completion {
	if object.ContentType == TypedArrayContentTypeBigInt {
		panic("TODO: BigInt is not implemented in TypedArraySetElement.")
	}

	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	numberVal := completion.Value.(*JavaScriptValue).Value.(*Number).Value

	if !IsValidIntegerIndex(runtime, object, index) {
		return NewUnusedCompletion()
	}

	byteIndexInBuffer := uint(index.Value)*TypedArrayElementSize(object) + object.ByteOffset

	SetValueInBuffer(
		runtime,
		object.ViewedArrayBuffer,
		byteIndexInBuffer,
		object.TypedArrayName,
		numberVal,
	)

	return NewUnusedCompletion()
}

func IsValidIntegerIndex(runtime *Runtime, object *TypedArrayObject, index *Number) bool {
	if IsDetachedArrayBuffer(object.ViewedArrayBuffer) {
		return false
	}

	if math.Floor(index.Value) != index.Value {
		return false
	}

	if index.Value < 0 {
		return false
	}

	taRecord := MakeTypedArrayWithBufferWitness(object, true)
	if taRecord.IsTypedArrayOutOfBounds() {
		return false
	}

	if index.Value >= float64(taRecord.TypedArrayLength()) {
		return false
	}

	return true
}

func TypedArrayElementSize(object *TypedArrayObject) uint {
	size, ok := TypedArrayElementSizes[object.TypedArrayName]
	if !ok {
		panic("Assert failed: Provided TypedArrayName is not mapped in TypedArrayElementSizes.")
	}
	return size
}

type TypedArrayWithBufferWitness struct {
	Object                         *TypedArrayObject
	CachedBufferByteLength         uint
	CachedBufferByteLengthDetached bool
}

func MakeTypedArrayWithBufferWitness(object *TypedArrayObject, unordered bool) *TypedArrayWithBufferWitness {
	if IsDetachedArrayBuffer(object.ViewedArrayBuffer) {
		return &TypedArrayWithBufferWitness{
			Object:                         object,
			CachedBufferByteLength:         0,
			CachedBufferByteLengthDetached: true,
		}
	}

	byteLength := ArrayBufferByteLength(object.ViewedArrayBuffer, unordered)
	return &TypedArrayWithBufferWitness{
		Object:                         object,
		CachedBufferByteLength:         byteLength,
		CachedBufferByteLengthDetached: false,
	}
}

func (record *TypedArrayWithBufferWitness) IsTypedArrayOutOfBounds() bool {
	object := record.Object

	if record.CachedBufferByteLengthDetached {
		return true
	}

	byteOffsetStart := object.ByteOffset

	var byteOffsetEnd uint
	if object.ArrayLengthAuto {
		byteOffsetEnd = record.CachedBufferByteLength
	} else {
		byteOffsetEnd = object.ByteOffset + object.ArrayLength*TypedArrayElementSize(object)
	}

	if byteOffsetStart > record.CachedBufferByteLength || byteOffsetEnd > record.CachedBufferByteLength {
		return true
	}

	return false
}

func (record *TypedArrayWithBufferWitness) TypedArrayLength() uint {
	if !record.Object.ArrayLengthAuto {
		return record.Object.ArrayLength
	}

	numerator := record.CachedBufferByteLength - record.Object.ByteOffset
	denominator := TypedArrayElementSize(record.Object)

	return uint(math.Floor(float64(numerator) / float64(denominator)))
}
