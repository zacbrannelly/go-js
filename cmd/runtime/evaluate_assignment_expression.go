package runtime

import (
	"fmt"
	"slices"

	"zbrannelly.dev/go-js/cmd/lexer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateAssignmentExpression(runtime *Runtime, assignmentExpression *ast.AssignmentExpressionNode) *Completion {
	lhsNode := assignmentExpression.GetTarget()
	rhsNode := assignmentExpression.GetValue()

	if assignmentExpression.Operator.Type == lexer.Assignment {
		return EvaluateSimpleAssignment(runtime, lhsNode, rhsNode)
	}

	if slices.Contains(lexer.AssignmentOperators, assignmentExpression.Operator.Type) {
		return EvaluateAssignmentOperatorExpression(runtime, lhsNode, assignmentExpression.Operator.Type, rhsNode)
	}

	if assignmentExpression.Operator.Type == lexer.AndAssignment {
		panic("TODO: Support logical assignment operators.")
	}

	if assignmentExpression.Operator.Type == lexer.OrAssignment {
		panic("TODO: Support logical assignment operators.")
	}

	if assignmentExpression.Operator.Type == lexer.NullishCoalescingAssignment {
		panic("TODO: Support logical assignment operators.")
	}

	panic("Unexpected assignment operator.")
}

func EvaluateSimpleAssignment(runtime *Runtime, lhsNode ast.Node, rhsNode ast.Node) *Completion {
	if lhsNode.GetNodeType() != ast.ObjectLiteral && lhsNode.GetNodeType() != ast.ArrayLiteral {
		lhsRefCompletion := Evaluate(runtime, lhsNode)
		if lhsRefCompletion.Type == Throw {
			return lhsRefCompletion
		}

		lhsRef := lhsRefCompletion.Value.(*JavaScriptValue)

		// TODO: Check if anon function definition, if so do something different according to the spec.

		rhsRefCompletion := Evaluate(runtime, rhsNode)
		if rhsRefCompletion.Type == Throw {
			return rhsRefCompletion
		}

		rhsRef := rhsRefCompletion.Value.(*JavaScriptValue)
		rhsValCompletion := GetValue(rhsRef)
		if rhsValCompletion.Type == Throw {
			return rhsValCompletion
		}

		rhsVal := rhsValCompletion.Value.(*JavaScriptValue)

		completion := PutValue(runtime, lhsRef, rhsVal)
		if completion.Type == Throw {
			return completion
		}

		return NewNormalCompletion(rhsVal)
	}

	panic("TODO: Support object & array destructuring.")
}

var AssignmentOpToOpTable = map[lexer.TokenType]lexer.TokenType{
	lexer.MultiplyAssignment:           lexer.Multiply,
	lexer.DivideAssignment:             lexer.Divide,
	lexer.ModuloAssignment:             lexer.Modulo,
	lexer.PlusAssignment:               lexer.Plus,
	lexer.MinusAssignment:              lexer.Minus,
	lexer.LeftShiftAssignment:          lexer.LeftShift,
	lexer.RightShiftAssignment:         lexer.RightShift,
	lexer.UnsignedRightShiftAssignment: lexer.UnsignedRightShift,
	lexer.BitwiseAndAssignment:         lexer.BitwiseAnd,
	lexer.BitwiseOrAssignment:          lexer.BitwiseOr,
	lexer.BitwiseXorAssignment:         lexer.BitwiseXor,
	lexer.ExponentiationAssignment:     lexer.Exponentiation,
}

func EvaluateAssignmentOperatorExpression(runtime *Runtime, lhsNode ast.Node, opType lexer.TokenType, rhsNode ast.Node) *Completion {
	lhsRefCompletion := Evaluate(runtime, lhsNode)
	if lhsRefCompletion.Type == Throw {
		return lhsRefCompletion
	}

	lhsRef := lhsRefCompletion.Value.(*JavaScriptValue)
	leftValCompletion := GetValue(lhsRef)
	if leftValCompletion.Type == Throw {
		return leftValCompletion
	}

	leftVal := leftValCompletion.Value.(*JavaScriptValue)

	rhsRefCompletion := Evaluate(runtime, rhsNode)
	if rhsRefCompletion.Type == Throw {
		return rhsRefCompletion
	}

	rhsRef := rhsRefCompletion.Value.(*JavaScriptValue)
	rhsValCompletion := GetValue(rhsRef)
	if rhsValCompletion.Type == Throw {
		return rhsValCompletion
	}

	rhsVal := rhsValCompletion.Value.(*JavaScriptValue)

	var assignmentOpType lexer.TokenType
	if assignmentOpType = AssignmentOpToOpTable[opType]; assignmentOpType == 0 {
		panic(fmt.Sprintf("Assert failed: Unsupported assignment operator: %s", lexer.OperatorTypeToString[opType]))
	}

	resultCompletion := ApplyStringOrNumericBinaryOperation(runtime, leftVal, assignmentOpType, rhsVal)
	if resultCompletion.Type == Throw {
		return resultCompletion
	}

	resultVal := resultCompletion.Value.(*JavaScriptValue)

	completion := PutValue(runtime, lhsRef, resultVal)
	if completion.Type == Throw {
		return completion
	}

	return resultCompletion
}
