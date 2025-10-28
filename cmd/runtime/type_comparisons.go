package runtime

import "slices"

func IsLessThan(x *JavaScriptValue, y *JavaScriptValue, leftFirst bool) *Completion {
	var primitiveX *JavaScriptValue
	var primitiveY *JavaScriptValue

	if leftFirst {
		// TODO: Prefer NUMBER primitive when that is supported.
		primitiveXCompletion := ToPrimitive(x)
		if primitiveXCompletion.Type != Normal {
			return primitiveXCompletion
		}

		primitiveX = primitiveXCompletion.Value.(*JavaScriptValue)

		// TODO: Prefer NUMBER primitive when that is supported.
		primitiveYCompletion := ToPrimitive(y)
		if primitiveYCompletion.Type != Normal {
			return primitiveYCompletion
		}

		primitiveY = primitiveYCompletion.Value.(*JavaScriptValue)
	} else {
		primitiveYCompletion := ToPrimitive(y)
		if primitiveYCompletion.Type != Normal {
			return primitiveYCompletion
		}

		primitiveY = primitiveYCompletion.Value.(*JavaScriptValue)

		primitiveXCompletion := ToPrimitive(x)
		if primitiveXCompletion.Type != Normal {
			return primitiveXCompletion
		}

		primitiveX = primitiveXCompletion.Value.(*JavaScriptValue)
	}

	if primitiveX.Type == TypeString && primitiveY.Type == TypeString {
		panic("TODO: Implement IsLessThan for String < String.")
	}

	if primitiveX.Type == TypeBigInt && primitiveY.Type == TypeString {
		panic("TODO: Implement IsLessThan for BigInt < String.")
	}

	if primitiveX.Type == TypeString && primitiveY.Type == TypeBigInt {
		panic("TODO: Implement IsLessThan for String < BigInt.")
	}

	numericXCompletion := ToNumeric(primitiveX)
	if numericXCompletion.Type != Normal {
		return numericXCompletion
	}

	numericYCompletion := ToNumeric(primitiveY)
	if numericYCompletion.Type != Normal {
		return numericYCompletion
	}

	numericX := numericXCompletion.Value.(*JavaScriptValue)
	numericY := numericYCompletion.Value.(*JavaScriptValue)

	if numericX.Type == numericY.Type {
		if numericX.Type == TypeNumber {
			return NumberLessThan(numericX.Value.(*Number), numericY.Value.(*Number))
		} else {
			panic("TODO: Support IsLessThan for BigInt < BigInt.")
		}
	}

	// From here on, x and y are different types and either Number or BigInt.
	panic("TODO: Support IsLessThan for Number < BigInt or BigInt < Number.")
}

func IsLooselyEqual(x *JavaScriptValue, y *JavaScriptValue) *Completion {
	if x.Type == y.Type {
		return IsStrictlyEqual(x, y)
	}

	// undefined == null
	if (x.Type == TypeUndefined || x.Type == TypeNull) && (y.Type == TypeUndefined || y.Type == TypeNull) {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	// number == string (coerce y to a number)
	if x.Type == TypeNumber && y.Type == TypeString {
		numberCompletion := ToNumber(y)
		if numberCompletion.Type != Normal {
			return numberCompletion
		}

		y = numberCompletion.Value.(*JavaScriptValue)
		return IsLooselyEqual(x, y)
	}

	// string == number (coerce x to a number)
	if x.Type == TypeString && y.Type == TypeNumber {
		numberCompletion := ToNumber(x)
		if numberCompletion.Type != Normal {
			return numberCompletion
		}

		x = numberCompletion.Value.(*JavaScriptValue)
		return IsLooselyEqual(x, y)
	}

	if x.Type == TypeBigInt && y.Type == TypeString {
		panic("TODO: Implement IsLooselyEqual for BigInt == String.")
	}

	if x.Type == TypeString && y.Type == TypeBigInt {
		return IsLooselyEqual(y, x)
	}

	if x.Type == TypeBoolean {
		numberCompletion := ToNumber(x)
		if numberCompletion.Type != Normal {
			return numberCompletion
		}

		x = numberCompletion.Value.(*JavaScriptValue)
		return IsLooselyEqual(x, y)
	}

	if y.Type == TypeBoolean {
		numberCompletion := ToNumber(y)
		if numberCompletion.Type != Normal {
			return numberCompletion
		}

		y = numberCompletion.Value.(*JavaScriptValue)
		return IsLooselyEqual(x, y)
	}

	if slices.Contains([]JavaScriptType{TypeString, TypeNumber, TypeBigInt, TypeSymbol}, x.Type) && y.Type == TypeObject {
		primitiveCompletion := ToPrimitive(y)
		if primitiveCompletion.Type != Normal {
			return primitiveCompletion
		}

		y = primitiveCompletion.Value.(*JavaScriptValue)
		return IsLooselyEqual(x, y)
	}

	if slices.Contains([]JavaScriptType{TypeString, TypeNumber, TypeBigInt, TypeSymbol}, y.Type) && x.Type == TypeObject {
		primitiveCompletion := ToPrimitive(x)
		if primitiveCompletion.Type != Normal {
			return primitiveCompletion
		}

		x = primitiveCompletion.Value.(*JavaScriptValue)
		return IsLooselyEqual(x, y)
	}

	if x.Type == TypeNumber && y.Type == TypeBigInt {
		panic("TODO: Implement IsLooselyEqual for Number == BigInt.")
	}

	if x.Type == TypeBigInt && y.Type == TypeNumber {
		panic("TODO: Implement IsLooselyEqual for BigInt == Number.")
	}

	return NewNormalCompletion(NewBooleanValue(false))
}

func IsStrictlyEqual(x *JavaScriptValue, y *JavaScriptValue) *Completion {
	if x.Type != y.Type {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	if x.Type == TypeNumber {
		return NumberEqual(x.Value.(*Number), y.Value.(*Number))
	}

	return SameValueNonNumber(x, y)
}

func SameValue(x *JavaScriptValue, y *JavaScriptValue) *Completion {
	if x.Type != y.Type {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	if x.Type == TypeNumber {
		return NewNormalCompletion(NewBooleanValue(NumberSameValue(x.Value.(*Number), y.Value.(*Number))))
	}

	return SameValueNonNumber(x, y)
}

func SameValueNonNumber(x *JavaScriptValue, y *JavaScriptValue) *Completion {
	if x.Type == TypeUndefined || x.Type == TypeNull {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	if x.Type == TypeBigInt {
		panic("TODO: Implement SameValueNonNumber for BigInt.")
	}

	if x.Type == TypeString {
		return NewNormalCompletion(NewBooleanValue(x.Value.(*String).Value == y.Value.(*String).Value))
	}

	if x.Type == TypeBoolean {
		return NewNormalCompletion(NewBooleanValue(x.Value.(*Boolean).Value == y.Value.(*Boolean).Value))
	}

	if x.Type == TypeSymbol {
		// Symbols are compared by reference, not value.
		return NewNormalCompletion(NewBooleanValue(x.Value.(*Symbol) == y.Value.(*Symbol)))
	}

	if x.Type == TypeObject {
		// Objects are compared by reference, not value.
		return NewNormalCompletion(NewBooleanValue(x.Value.(*Object) == y.Value.(*Object)))
	}

	panic("Unexpected type in SameValueNonNumber.")
}

func SameValueZero(x *JavaScriptValue, y *JavaScriptValue) *Completion {
	if x.Type != y.Type {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	if x.Type == TypeNumber {
		return NewNormalCompletion(NewBooleanValue(NumberSameValueZero(x.Value.(*Number), y.Value.(*Number))))
	}

	return SameValueNonNumber(x, y)
}
