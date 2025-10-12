package runtime

func IsLessThan(runtime *Runtime, x *JavaScriptValue, y *JavaScriptValue, leftFirst bool) *Completion {
	var primitiveX *JavaScriptValue
	var primitiveY *JavaScriptValue

	if leftFirst {
		// TODO: Prefer NUMBER primitive when that is supported.
		primitiveXCompletion := ToPrimitive(runtime, x)
		if primitiveXCompletion.Type == Throw {
			return primitiveXCompletion
		}

		primitiveX = primitiveXCompletion.Value.(*JavaScriptValue)

		// TODO: Prefer NUMBER primitive when that is supported.
		primitiveYCompletion := ToPrimitive(runtime, y)
		if primitiveYCompletion.Type == Throw {
			return primitiveYCompletion
		}

		primitiveY = primitiveYCompletion.Value.(*JavaScriptValue)
	} else {
		primitiveYCompletion := ToPrimitive(runtime, y)
		if primitiveYCompletion.Type == Throw {
			return primitiveYCompletion
		}

		primitiveY = primitiveYCompletion.Value.(*JavaScriptValue)

		primitiveXCompletion := ToPrimitive(runtime, x)
		if primitiveXCompletion.Type == Throw {
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

	numericXCompletion := ToNumeric(runtime, primitiveX)
	if numericXCompletion.Type == Throw {
		return numericXCompletion
	}

	numericYCompletion := ToNumeric(runtime, primitiveY)
	if numericYCompletion.Type == Throw {
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
