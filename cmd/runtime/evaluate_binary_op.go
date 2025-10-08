package runtime

import "fmt"

func ApplyStringOrNumericBinaryOperation(
	runtime *Runtime,
	leftRef *JavaScriptValue,
	opText string,
	rightRef *JavaScriptValue,
) *Completion {
	if opText == "+" {
		leftPrimitiveCompletion := ToPrimitive(runtime, leftRef)
		rightPrimitiveCompletion := ToPrimitive(runtime, rightRef)

		if leftPrimitiveCompletion.Type == Throw {
			return leftPrimitiveCompletion
		}
		if rightPrimitiveCompletion.Type == Throw {
			return rightPrimitiveCompletion
		}

		leftPrimitive := leftPrimitiveCompletion.Value.(*JavaScriptValue)
		rightPrimitive := rightPrimitiveCompletion.Value.(*JavaScriptValue)

		// Concatenate strings if either operand is a string.
		if leftPrimitive.Type == TypeString || rightPrimitive.Type == TypeString {
			leftStringCompletion := ToString(runtime, leftPrimitive)
			rightStringCompletion := ToString(runtime, rightPrimitive)

			if leftStringCompletion.Type == Throw {
				return leftStringCompletion
			}
			if rightStringCompletion.Type == Throw {
				return rightStringCompletion
			}

			leftString := leftStringCompletion.Value.(*JavaScriptValue)
			rightString := rightStringCompletion.Value.(*JavaScriptValue)

			return NewNormalCompletion(NewStringValue(leftString.Value.(string) + rightString.Value.(string)))
		}

		leftRef = leftPrimitive
		rightRef = rightPrimitive
	}

	leftNumericCompletion := ToNumeric(runtime, leftRef)
	rightNumericCompletion := ToNumeric(runtime, rightRef)

	if leftNumericCompletion.Type == Throw {
		return leftNumericCompletion
	}

	if rightNumericCompletion.Type == Throw {
		return rightNumericCompletion
	}

	leftNumeric := leftNumericCompletion.Value.(*JavaScriptValue)
	rightNumeric := rightNumericCompletion.Value.(*JavaScriptValue)

	if leftNumeric.Type != rightNumeric.Type {
		return NewThrowCompletion(NewTypeError(fmt.Sprintf("Cannot apply %s to %s and %s", opText, TypeNames[leftNumeric.Type], TypeNames[rightNumeric.Type])))
	}

	if leftNumeric.Type == TypeBigInt {
		panic("TODO: BigInt binary operations are not implemented.")
	}

	switch opText {
	case "+":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberAdd(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	}

	panic(fmt.Sprintf("Assert failed: Unsupported operator: %s", opText))
}

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

	panic("TODO: ToNumber for non-Number values is not implemented.")
}

func ToString(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeString {
		return NewNormalCompletion(value)
	}

	panic("TODO: ToString for non-String values is not implemented.")
}

func ToPrimitive(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type == TypeObject {
		panic("TODO: ToPrimitive for Object values is not implemented.")
	}

	return NewNormalCompletion(value)
}
