package runtime

import (
	"encoding/binary"
	"math"
	"slices"

	"github.com/x448/float16"
)

func AllocateArrayBuffer(runtime *Runtime, constructor FunctionInterface, byteLength uint) *Completion {
	completion := OrdinaryCreateFromConstructor(runtime, constructor, IntrinsicArrayBufferPrototype)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	obj := objectVal.Value.(*Object)

	if float64(byteLength) > math.Pow(2, 53)-1 {
		return NewThrowCompletion(NewRangeError(runtime, "ArrayBuffer length too large"))
	}

	obj.ArrayBufferByteLength = byteLength
	obj.ArrayBufferData = make([]byte, byteLength)

	for i := range obj.ArrayBufferData {
		obj.ArrayBufferData[i] = 0
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, obj))
}

func AllocateArrayBufferWithMaxByteLength(runtime *Runtime, constructor FunctionInterface, byteLength uint, maxByteLength uint) *Completion {
	if byteLength > maxByteLength {
		return NewThrowCompletion(NewRangeError(runtime, "ArrayBuffer length exceeds maxByteLength"))
	}

	completion := OrdinaryCreateFromConstructor(runtime, constructor, IntrinsicArrayBufferPrototype)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	obj := objectVal.Value.(*Object)

	if float64(byteLength) > math.Pow(2, 53)-1 {
		return NewThrowCompletion(NewRangeError(runtime, "ArrayBuffer length too large"))
	}

	obj.ArrayBufferByteLength = byteLength
	obj.ArrayBufferData = make([]byte, byteLength)

	for i := range obj.ArrayBufferData {
		obj.ArrayBufferData[i] = 0
	}

	obj.ArrayBufferHasMaxByteLength = true
	obj.ArrayBufferMaxByteLength = maxByteLength

	// TODO: Check if the array buffer max size is supported.

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, obj))
}

func IsFixedLengthArrayBuffer(object *Object) bool {
	return !object.ArrayBufferHasMaxByteLength
}

func IsSharedArrayBuffer(object *Object) bool {
	if object.ArrayBufferData == nil {
		return false
	}
	return object.ArrayBufferDataIsShared
}

func IsDetachedArrayBuffer(object *Object) bool {
	return object.ArrayBufferData == nil
}

func ArrayBufferByteLength(object *Object, unordered bool) uint {
	if IsSharedArrayBuffer(object) {
		panic("TODO: Support shared array buffer in ArrayBufferByteLength.")
	}

	if IsDetachedArrayBuffer(object) {
		panic("Assert failed: ArrayBuffer is detached in ArrayBufferByteLength.")
	}

	return object.ArrayBufferByteLength
}

func GetValueFromBuffer(
	runtime *Runtime,
	arrayBuffer *Object,
	byteIndex uint,
	dataType TypedArrayName,
	isTypedArray bool,
	unordered bool,
) *JavaScriptValue {
	elementSize, ok := TypedArrayElementSizes[dataType]
	if !ok {
		panic("Assert failed: Provided dataType is not mapped in TypedArrayElementSizes.")
	}

	var rawValue []byte
	if IsSharedArrayBuffer(arrayBuffer) {
		panic("TODO: Support shared array buffer in GetValueFromBuffer.")
	} else {
		rawValue = arrayBuffer.ArrayBufferData[byteIndex : byteIndex+elementSize]
	}

	return RawBytesToNumber(dataType, rawValue, runtime.IsLittleEndian())
}

func SetValueInBuffer(
	runtime *Runtime,
	arrayBuffer *Object,
	byteIndex uint,
	dataType TypedArrayName,
	value float64,
) {
	elementSize, ok := TypedArrayElementSizes[dataType]
	if !ok {
		panic("Assert failed: Provided dataType is not mapped in TypedArrayElementSizes.")
	}

	rawBytes := NumericToRawBytes(runtime, dataType, value, runtime.IsLittleEndian())

	if IsSharedArrayBuffer(arrayBuffer) {
		panic("TODO: Support shared array buffer in SetValueInBuffer.")
	} else {
		copy(arrayBuffer.ArrayBufferData[byteIndex:byteIndex+elementSize], rawBytes)
	}
}

func NumericToRawBytes(runtime *Runtime, dataType TypedArrayName, value float64, littleEndian bool) []byte {
	var rawBytes []byte
	switch dataType {
	case TypedArrayNameFloat16:
		bits := float16.Fromfloat32(float32(value)).Bits()
		rawBytes = make([]byte, 2)
		binary.LittleEndian.PutUint16(rawBytes, bits)
	case TypedArrayNameFloat32:
		bits := math.Float32bits(float32(value))
		rawBytes = make([]byte, 4)
		binary.LittleEndian.PutUint32(rawBytes, bits)
	case TypedArrayNameFloat64:
		bits := math.Float64bits(value)
		rawBytes = make([]byte, 8)
		binary.LittleEndian.PutUint64(rawBytes, bits)
	default:
		elementSize, ok := TypedArrayElementSizes[dataType]
		if !ok {
			panic("Assert failed: Provided dataType is not mapped in TypedArrayElementSizes.")
		}
		rawBytes = make([]byte, elementSize)

		conversionFunction, ok := TypedArrayConversionFunctions[dataType]
		if !ok {
			panic("Assert failed: Provided dataType is not mapped in TypedArrayConversionFunctions.")
		}
		completion := conversionFunction(runtime, NewNumberValue(value, false))
		if completion.Type != Normal {
			panic("Assert failed: Conversion function returned an error.")
		}

		numberValue := completion.Value.(*JavaScriptValue).Value.(*Number)

		int64Value := int64(numberValue.Value)

		// Copy elementSize bytes from the 64-bit value
		switch elementSize {
		case 1:
			rawBytes[0] = byte(int64Value)
		case 2:
			binary.LittleEndian.PutUint16(rawBytes, uint16(int64Value))
		case 4:
			binary.LittleEndian.PutUint32(rawBytes, uint32(int64Value))
		case 8:
			binary.LittleEndian.PutUint64(rawBytes, uint64(int64Value))
		default:
			panic("Assert failed: Unsupported element size.")
		}
	}

	if !littleEndian {
		slices.Reverse(rawBytes)
	}

	return rawBytes
}

func RawBytesToNumber(dataType TypedArrayName, rawValue []byte, littleEndian bool) *JavaScriptValue {
	if !littleEndian {
		slices.Reverse(rawValue)
	}

	switch dataType {
	case TypedArrayNameInt8:
		value := int8(rawValue[0])
		return NewNumberValue(float64(value), false)
	case TypedArrayNameUint8:
		value := uint8(rawValue[0])
		return NewNumberValue(float64(value), false)
	case TypedArrayNameUint8Clamped:
		value := uint8(rawValue[0])
		return NewNumberValue(float64(value), false)
	case TypedArrayNameInt16:
		value := binary.LittleEndian.Uint16(rawValue)
		return NewNumberValue(float64(value), false)
	case TypedArrayNameUint16:
		value := binary.LittleEndian.Uint16(rawValue)
		return NewNumberValue(float64(value), false)
	case TypedArrayNameInt32:
		value := binary.LittleEndian.Uint32(rawValue)
		return NewNumberValue(float64(value), false)
	case TypedArrayNameUint32:
		value := binary.LittleEndian.Uint32(rawValue)
		return NewNumberValue(float64(value), false)
	case TypedArrayNameBigInt64:
		panic("TODO: BigInt64 is not implemented in RawBytesToNumber.")
	case TypedArrayNameBigUint64:
		panic("TODO: BigUint64 is not implemented in RawBytesToNumber.")
	case TypedArrayNameFloat16:
		value := float16.Frombits(binary.LittleEndian.Uint16(rawValue))
		valueFloat := float64(value.Float32())
		if math.IsNaN(valueFloat) {
			return NewNumberValue(0, true)
		}
		return NewNumberValue(valueFloat, false)
	case TypedArrayNameFloat32:
		value := math.Float32frombits(binary.LittleEndian.Uint32(rawValue))
		if math.IsNaN(float64(value)) {
			return NewNumberValue(0, true)
		}
		return NewNumberValue(float64(value), false)
	case TypedArrayNameFloat64:
		value := math.Float64frombits(binary.LittleEndian.Uint64(rawValue))
		if math.IsNaN(value) {
			return NewNumberValue(0, true)
		}
		return NewNumberValue(value, false)
	}

	panic("Assert failed: Provided dataType is not mapped in RawBytesToNumber.")
}
