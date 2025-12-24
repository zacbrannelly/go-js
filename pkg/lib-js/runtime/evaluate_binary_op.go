package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateStringOrNumericBinaryExpression(runtime *Runtime, operatorExpression ast.OperatorNode) *Completion {
	leftRef := Evaluate(runtime, operatorExpression.GetLeft())
	if leftRef.Type != Normal {
		return leftRef
	}

	leftValCompletion := GetValue(runtime, leftRef.Value.(*JavaScriptValue))
	if leftValCompletion.Type != Normal {
		return leftValCompletion
	}

	rightRef := Evaluate(runtime, operatorExpression.GetRight())
	if rightRef.Type != Normal {
		return rightRef
	}

	rightValCompletion := GetValue(runtime, rightRef.Value.(*JavaScriptValue))
	if rightValCompletion.Type != Normal {
		return rightValCompletion
	}

	leftVal := leftValCompletion.Value.(*JavaScriptValue)
	rightVal := rightValCompletion.Value.(*JavaScriptValue)

	return ApplyStringOrNumericBinaryOperation(
		runtime,
		leftVal,
		operatorExpression.GetOperator().Type,
		rightVal,
	)
}

func ApplyStringOrNumericBinaryOperation(
	runtime *Runtime,
	leftRef *JavaScriptValue,
	opType lexer.TokenType,
	rightRef *JavaScriptValue,
) *Completion {
	if opType == lexer.Plus {
		leftPrimitiveCompletion := ToPrimitive(leftRef)
		rightPrimitiveCompletion := ToPrimitive(rightRef)

		if leftPrimitiveCompletion.Type != Normal {
			return leftPrimitiveCompletion
		}
		if rightPrimitiveCompletion.Type != Normal {
			return rightPrimitiveCompletion
		}

		leftPrimitive := leftPrimitiveCompletion.Value.(*JavaScriptValue)
		rightPrimitive := rightPrimitiveCompletion.Value.(*JavaScriptValue)

		// Concatenate strings if either operand is a string.
		if leftPrimitive.Type == TypeString || rightPrimitive.Type == TypeString {
			leftStringCompletion := ToString(leftPrimitive)
			rightStringCompletion := ToString(rightPrimitive)

			if leftStringCompletion.Type != Normal {
				return leftStringCompletion
			}
			if rightStringCompletion.Type != Normal {
				return rightStringCompletion
			}

			leftString := leftStringCompletion.Value.(*JavaScriptValue).Value.(*String)
			rightString := rightStringCompletion.Value.(*JavaScriptValue).Value.(*String)

			return NewNormalCompletion(NewJavaScriptValue(TypeString, StringAdd(leftString, rightString)))
		}

		leftRef = leftPrimitive
		rightRef = rightPrimitive
	}

	leftNumericCompletion := ToNumeric(leftRef)
	rightNumericCompletion := ToNumeric(rightRef)

	if leftNumericCompletion.Type != Normal {
		return leftNumericCompletion
	}

	if rightNumericCompletion.Type != Normal {
		return rightNumericCompletion
	}

	leftNumeric := leftNumericCompletion.Value.(*JavaScriptValue)
	rightNumeric := rightNumericCompletion.Value.(*JavaScriptValue)

	if leftNumeric.Type != rightNumeric.Type {
		return NewThrowCompletion(NewTypeError(runtime, fmt.Sprintf("Cannot apply %s to %s and %s", lexer.OperatorTypeToString[opType], TypeNames[leftNumeric.Type], TypeNames[rightNumeric.Type])))
	}

	if leftNumeric.Type == TypeBigInt {
		panic("TODO: BigInt binary operations are not implemented.")
	}

	if leftNumeric.Type != TypeNumber {
		panic("Assert failed: Left numeric is not a number in binary operation.")
	}

	switch opType {
	case lexer.Plus:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberAdd(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.Minus:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberSub(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.Multiply:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberMul(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.Divide:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberDiv(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.Exponentiation:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberExponentiate(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.Modulo:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberRemainder(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.LeftShift:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberLeftShift(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.RightShift:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberSignedRightShift(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.UnsignedRightShift:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberUnsignedRightShift(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.BitwiseAnd:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberBitwiseAnd(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.BitwiseOr:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberBitwiseOr(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	case lexer.BitwiseXor:
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberBitwiseXor(leftNumeric.Value.(*Number), rightNumeric.Value.(*Number))))
	}

	panic(fmt.Sprintf("Assert failed: Unsupported operator: %s", lexer.OperatorTypeToString[opType]))
}
