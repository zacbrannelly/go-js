package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateStringOrNumericBinaryExpression(runtime *Runtime, operatorExpression ast.OperatorNode) *Completion {
	leftRef := Evaluate(runtime, operatorExpression.GetLeft())
	rightRef := Evaluate(runtime, operatorExpression.GetRight())

	if leftRef.Type == Throw {
		return leftRef
	}

	if rightRef.Type == Throw {
		return rightRef
	}

	// Resolve references to their values (if references).
	leftValCompletion := GetValue(leftRef.Value.(*JavaScriptValue))
	rightValCompletion := GetValue(rightRef.Value.(*JavaScriptValue))

	if leftValCompletion.Type == Throw {
		return leftValCompletion
	}
	if rightValCompletion.Type == Throw {
		return rightValCompletion
	}

	leftVal := leftValCompletion.Value.(*JavaScriptValue)
	rightVal := rightValCompletion.Value.(*JavaScriptValue)

	return ApplyStringOrNumericBinaryOperation(
		runtime,
		leftVal,
		operatorExpression.GetOperator().Value,
		rightVal,
	)
}

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

			leftString := leftStringCompletion.Value.(*JavaScriptValue).Value.(*String)
			rightString := rightStringCompletion.Value.(*JavaScriptValue).Value.(*String)

			return NewNormalCompletion(NewJavaScriptValue(TypeString, StringAdd(leftString, rightString)))
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

	if leftNumeric.Type != TypeNumber {
		panic("Assert failed: Left numeric is not a number in binary operation.")
	}

	switch opText {
	case "+":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberAdd(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "-":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberSub(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "*":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberMul(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "/":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberDiv(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "**":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberExponentiate(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "%":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberRemainder(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "<<":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberLeftShift(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case ">>":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberSignedRightShift(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case ">>>":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberUnsignedRightShift(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "&":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberBitwiseAnd(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "|":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberBitwiseOr(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case "^":
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberBitwiseXor(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	}

	panic(fmt.Sprintf("Assert failed: Unsupported operator: %s", opText))
}
