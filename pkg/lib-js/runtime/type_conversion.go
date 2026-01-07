package runtime

import (
	"math"
	"strconv"
)

func ToNumeric(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		panic("TODO: ToNumeric for Object values is not implemented.")
	}

	if value.Type == TypeBigInt {
		return NewNormalCompletion(value)
	}

	return ToNumber(runtime, value)
}

func ToNumber(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeNumber {
		return NewNormalCompletion(value)
	}

	if value.Type == TypeUndefined {
		// undefined -> NaN
		return NewNormalCompletion(NewNumberValue(0, true))
	}

	if value.Type == TypeNull {
		// null -> +0
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	if value.Type == TypeString {
		// TODO: Implement parser for StringNumericLiteral.
		number, err := strconv.ParseFloat(value.Value.(*String).Value, 64)
		if err != nil {
			// NaN
			return NewNormalCompletion(NewNumberValue(0, true))
		}

		return NewNormalCompletion(NewNumberValue(number, false))
	}

	if value.Type == TypeBoolean && value.Value.(*Boolean).Value {
		// true -> +1
		return NewNormalCompletion(NewNumberValue(1, false))
	}

	if value.Type == TypeBoolean && !value.Value.(*Boolean).Value {
		// false -> +0
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	if value.Type == TypeObject {
		completion := ToPrimitive(runtime, value)
		if completion.Type != Normal {
			return completion
		}

		primValue := completion.Value.(*JavaScriptValue)
		return ToNumber(runtime, primValue)
	}

	if value.Type == TypeSymbol {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot convert a Symbol to a number"))
	}

	panic("TODO: ToNumber for non-Number values is not implemented.")
}

func ToString(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeString {
		return NewNormalCompletion(value)
	}

	if value.Type == TypeUndefined {
		return NewNormalCompletion(NewStringValue("undefined"))
	}

	if value.Type == TypeNull {
		return NewNormalCompletion(NewStringValue("null"))
	}

	if value.Type == TypeNumber {
		return NewNormalCompletion(NumberToString(value.Value.(*Number), 10))
	}

	if value.Type == TypeObject {
		return ToPrimitiveWithPreferredType(runtime, value, PreferredTypeString)
	}

	panic("TODO: ToString for non-String values is not implemented.")
}

type PreferredType int

const (
	PreferredTypeUndefined PreferredType = iota
	PreferredTypeNumber
	PreferredTypeString
)

func ToPrimitive(runtime *Runtime, value *JavaScriptValue) *Completion {
	return ToPrimitiveWithPreferredType(runtime, value, PreferredTypeUndefined)
}

func ToPrimitiveWithPreferredType(runtime *Runtime, value *JavaScriptValue, preferredType PreferredType) *Completion {
	if value.Type == TypeObject {
		completion := GetMethod(runtime, value, runtime.SymbolToPrimitive)
		if completion.Type != Normal {
			return completion
		}

		method := completion.Value.(*JavaScriptValue)
		if method.Type != TypeUndefined {
			panic("TODO: ToPrimitive for Object values with Symbol.toPrimitive is not implemented.")
		}

		if preferredType == PreferredTypeUndefined {
			preferredType = PreferredTypeNumber
		}

		return OrdinaryToPrimitive(runtime, value, preferredType)
	}

	return NewNormalCompletion(value)
}

func OrdinaryToPrimitive(runtime *Runtime, value *JavaScriptValue, hint PreferredType) *Completion {
	object := value.Value.(ObjectInterface)

	var methodNames []string
	if hint == PreferredTypeNumber {
		methodNames = []string{"valueOf", "toString"}
	} else {
		methodNames = []string{"toString", "valueOf"}
	}

	for _, methodName := range methodNames {
		methodKey := NewStringValue(methodName)
		completion := object.Get(runtime, methodKey, value)
		if completion.Type != Normal {
			return completion
		}

		method := completion.Value.(*JavaScriptValue)
		if IsCallable(method) {
			completion := Call(runtime, method, value, []*JavaScriptValue{})
			if completion.Type != Normal {
				return completion
			}

			result := completion.Value.(*JavaScriptValue)
			if result.Type != TypeObject {
				return completion
			}
		}
	}

	return NewThrowCompletion(NewTypeError(runtime, "Cannot convert object to primitive value."))
}

func ToBoolean(value *JavaScriptValue) *Completion {
	if value.Type == TypeBoolean {
		return NewNormalCompletion(value)
	}

	// Null and undefined are falsy.
	if value.Type == TypeNull || value.Type == TypeUndefined {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	// 0 is falsy.
	if value.Type == TypeNumber {
		if value.Value.(*Number).NaN {
			return NewNormalCompletion(NewBooleanValue(false))
		}
		return NewNormalCompletion(NewBooleanValue(value.Value.(*Number).Value != 0))
	}

	// Empty string is falsy.
	if value.Type == TypeString {
		return NewNormalCompletion(NewBooleanValue(value.Value.(*String).Value != ""))
	}

	// Otherwise, true.
	return NewNormalCompletion(NewBooleanValue(true))
}

func ToObject(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		return NewNormalCompletion(value)
	}

	if value.Type == TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot convert undefined to an object"))
	}

	if value.Type == TypeNull {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot convert null to an object"))
	}

	if value.Type == TypeString {
		proto := runtime.GetRunningRealm().GetIntrinsic(IntrinsicStringPrototype)
		stringObj := StringCreate(runtime, value, proto)
		return NewNormalCompletion(NewJavaScriptValue(TypeObject, stringObj))
	}

	if value.Type == TypeNumber {
		object := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicNumberPrototype))
		object.(*Object).NumberData = value
		return NewNormalCompletion(NewJavaScriptValue(TypeObject, object))
	}

	if value.Type == TypeBoolean {
		object := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicBooleanPrototype))
		object.(*Object).BooleanData = value
		return NewNormalCompletion(NewJavaScriptValue(TypeObject, object))
	}

	panic("TODO: ToObject for non-Object values is not implemented.")
}

func ToUint32(runtime *Runtime, value *JavaScriptValue) *Completion {
	numberCompletion := ToNumber(runtime, value)
	if numberCompletion.Type != Normal {
		return numberCompletion
	}

	finalValue := math.Trunc(numberCompletion.Value.(*JavaScriptValue).Value.(*Number).Value)
	finalValueUint64 := uint64(finalValue) % uint64(math.Pow(2, 32))

	return NewNormalCompletion(NewNumberValue(float64(finalValueUint64), false))
}

func ToLength(runtime *Runtime, value *JavaScriptValue) *Completion {
	lenCompletion := ToIntegerOrInfinity(runtime, value)
	if lenCompletion.Type != Normal {
		return lenCompletion
	}

	len := lenCompletion.Value.(*JavaScriptValue).Value.(*Number).Value
	if len <= 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	return NewNormalCompletion(NewNumberValue(math.Min(len, math.Pow(2, 53)-1), false))
}

func ToInt8(runtime *Runtime, value *JavaScriptValue) *Completion {
	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	numberVal := completion.Value.(*JavaScriptValue).Value.(*Number)

	if numberVal.NaN || math.IsInf(numberVal.Value, 0) {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	numberVal.Value = truncate(numberVal.Value)

	int8bit := int(numberVal.Value) % 256
	if int8bit >= 128 {
		int8bit -= 256
	}

	return NewNormalCompletion(NewNumberValue(float64(int8bit), false))
}

func ToUint8(runtime *Runtime, value *JavaScriptValue) *Completion {
	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	numberVal := completion.Value.(*JavaScriptValue).Value.(*Number)

	if numberVal.NaN || math.IsInf(numberVal.Value, 0) || numberVal.Value == 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	numberVal.Value = truncate(numberVal.Value)

	int8bit := int(numberVal.Value) % 256

	return NewNormalCompletion(NewNumberValue(float64(int8bit), false))
}

func ToUint8Clamped(runtime *Runtime, value *JavaScriptValue) *Completion {
	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	numberVal := completion.Value.(*JavaScriptValue).Value.(*Number)

	if numberVal.NaN {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	// Clamp between 0 and 255
	clamped := numberVal.Value
	if clamped < 0 {
		clamped = 0
	} else if clamped > 255 {
		clamped = 255
	}

	f := math.Floor(clamped)

	if clamped < f+0.5 {
		return NewNormalCompletion(NewNumberValue(f, false))
	}

	if clamped > f+0.5 {
		return NewNormalCompletion(NewNumberValue(f+1, false))
	}

	// clamped == f + 0.5, round to even
	if int(f)%2 == 0 {
		return NewNormalCompletion(NewNumberValue(f, false))
	}

	return NewNormalCompletion(NewNumberValue(f+1, false))
}

func ToInt16(runtime *Runtime, value *JavaScriptValue) *Completion {
	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	numberVal := completion.Value.(*JavaScriptValue).Value.(*Number)

	if numberVal.NaN || math.IsInf(numberVal.Value, 0) || numberVal.Value == 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	numberVal.Value = truncate(numberVal.Value)

	int16bit := int(numberVal.Value) % 65536
	if int16bit < 0 {
		int16bit += 65536
	}

	if int16bit >= 32768 {
		return NewNormalCompletion(NewNumberValue(float64(int16bit-65536), false))
	}

	return NewNormalCompletion(NewNumberValue(float64(int16bit), false))
}

func ToUint16(runtime *Runtime, value *JavaScriptValue) *Completion {
	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	numberVal := completion.Value.(*JavaScriptValue).Value.(*Number)

	if numberVal.NaN || math.IsInf(numberVal.Value, 0) || numberVal.Value == 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	numberVal.Value = truncate(numberVal.Value)
	int16bit := int(numberVal.Value) % 65536

	return NewNormalCompletion(NewNumberValue(float64(int16bit), false))
}

func ToInt32(runtime *Runtime, value *JavaScriptValue) *Completion {
	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	numberVal := completion.Value.(*JavaScriptValue).Value.(*Number)

	if numberVal.NaN || math.IsInf(numberVal.Value, 0) || numberVal.Value == 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	numberVal.Value = truncate(numberVal.Value)

	int32bit := int64(numberVal.Value) % (1 << 32)

	if int32bit >= (1 << 31) {
		return NewNormalCompletion(NewNumberValue(float64(int32bit-(1<<32)), false))
	}

	return NewNormalCompletion(NewNumberValue(float64(int32bit), false))
}

func ToIntegerOrInfinity(runtime *Runtime, value *JavaScriptValue) *Completion {
	numberCompletion := ToNumber(runtime, value)
	if numberCompletion.Type != Normal {
		return numberCompletion
	}

	number := numberCompletion.Value.(*JavaScriptValue).Value.(*Number)
	if number.NaN || number.Value == 0 {
		return NewNormalCompletion(NewNumberValue(0, false))
	}

	if number.Value == math.Inf(1) {
		return NewNormalCompletion(NewNumberValue(math.Inf(1), false))
	}

	if number.Value == math.Inf(-1) {
		return NewNormalCompletion(NewNumberValue(math.Inf(-1), false))
	}

	return NewNormalCompletion(NewNumberValue(truncate(number.Value), false))
}

func truncate(value float64) float64 {
	if value < 0 {
		return -math.Floor(-value)
	}
	return math.Floor(value)
}

func CanonicalNumericIndexString(runtime *Runtime, value *JavaScriptValue) *JavaScriptValue {
	if value.Type != TypeString {
		panic("Assert failed: CanonicalNumericIndexString value is not a string.")
	}

	valueString := value.Value.(*String).Value

	if valueString == "-0" {
		return NewNumberValue(-0, false)
	}

	completion := ToNumber(runtime, value)
	if completion.Type != Normal {
		panic("Assert failed: CanonicalNumericIndexString ToNumber threw an error.")
	}

	numberVal := completion.Value.(*JavaScriptValue)

	completion = ToString(runtime, numberVal)
	if completion.Type != Normal {
		panic("Assert failed: CanonicalNumericIndexString ToString threw an error.")
	}

	toString := completion.Value.(*JavaScriptValue).Value.(*String).Value

	if toString == valueString {
		return numberVal
	}

	return NewUndefinedValue()
}

func ToIndex(runtime *Runtime, value *JavaScriptValue) *Completion {
	completion := ToIntegerOrInfinity(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	index := completion.Value.(*JavaScriptValue).Value.(*Number).Value
	if index < 0 {
		return NewThrowCompletion(NewRangeError(runtime, "Index is negative"))
	}
	if index > math.Pow(2, 53)-1 {
		return NewThrowCompletion(NewRangeError(runtime, "Index is too large"))
	}

	return completion
}
