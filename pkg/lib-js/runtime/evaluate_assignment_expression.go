package runtime

import (
	"fmt"
	"slices"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

var LogicalAssignmentOps = []lexer.TokenType{
	lexer.AndAssignment,
	lexer.OrAssignment,
	lexer.NullishCoalescingAssignment,
}

func EvaluateAssignmentExpression(runtime *Runtime, assignmentExpression *ast.AssignmentExpressionNode) *Completion {
	lhsNode := assignmentExpression.GetTarget()
	rhsNode := assignmentExpression.GetValue()

	if assignmentExpression.Operator.Type == lexer.Assignment {
		return EvaluateSimpleAssignment(runtime, lhsNode, rhsNode)
	}

	if slices.Contains(lexer.AssignmentOperators, assignmentExpression.Operator.Type) {
		return EvaluateAssignmentOperatorExpression(runtime, lhsNode, assignmentExpression.Operator.Type, rhsNode)
	}

	if slices.Contains(LogicalAssignmentOps, assignmentExpression.Operator.Type) {
		return EvaluateLogicalAssignmentExpression(runtime, lhsNode, assignmentExpression.Operator.Type, rhsNode)
	}

	panic("Unexpected assignment operator.")
}

func EvaluateSimpleAssignment(runtime *Runtime, lhsNode ast.Node, rhsNode ast.Node) *Completion {
	if lhsNode.GetNodeType() != ast.ObjectLiteral && lhsNode.GetNodeType() != ast.ArrayLiteral {
		lhsRefCompletion := Evaluate(runtime, lhsNode)
		if lhsRefCompletion.Type != Normal {
			return lhsRefCompletion
		}

		lhsRef := lhsRefCompletion.Value.(*JavaScriptValue)

		// TODO: Check if anon function definition, if so do something different according to the spec.

		rhsRefCompletion := Evaluate(runtime, rhsNode)
		if rhsRefCompletion.Type != Normal {
			return rhsRefCompletion
		}

		rhsRef := rhsRefCompletion.Value.(*JavaScriptValue)
		rhsValCompletion := GetValue(runtime, rhsRef)
		if rhsValCompletion.Type != Normal {
			return rhsValCompletion
		}

		rhsVal := rhsValCompletion.Value.(*JavaScriptValue)

		completion := PutValue(runtime, lhsRef, rhsVal)
		if completion.Type != Normal {
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
	if lhsRefCompletion.Type != Normal {
		return lhsRefCompletion
	}

	lhsRef := lhsRefCompletion.Value.(*JavaScriptValue)
	leftValCompletion := GetValue(runtime, lhsRef)
	if leftValCompletion.Type != Normal {
		return leftValCompletion
	}

	leftVal := leftValCompletion.Value.(*JavaScriptValue)

	rhsRefCompletion := Evaluate(runtime, rhsNode)
	if rhsRefCompletion.Type != Normal {
		return rhsRefCompletion
	}

	rhsRef := rhsRefCompletion.Value.(*JavaScriptValue)
	rhsValCompletion := GetValue(runtime, rhsRef)
	if rhsValCompletion.Type != Normal {
		return rhsValCompletion
	}

	rhsVal := rhsValCompletion.Value.(*JavaScriptValue)

	var assignmentOpType lexer.TokenType
	if assignmentOpType = AssignmentOpToOpTable[opType]; assignmentOpType == 0 {
		panic(fmt.Sprintf("Assert failed: Unsupported assignment operator: %s", lexer.OperatorTypeToString[opType]))
	}

	resultCompletion := ApplyStringOrNumericBinaryOperation(runtime, leftVal, assignmentOpType, rhsVal)
	if resultCompletion.Type != Normal {
		return resultCompletion
	}

	resultVal := resultCompletion.Value.(*JavaScriptValue)

	completion := PutValue(runtime, lhsRef, resultVal)
	if completion.Type != Normal {
		return completion
	}

	return resultCompletion
}

func EvaluateLogicalAssignmentExpression(runtime *Runtime, lhsNode ast.Node, opType lexer.TokenType, rhsNode ast.Node) *Completion {
	lhsRefCompletion := Evaluate(runtime, lhsNode)
	if lhsRefCompletion.Type != Normal {
		return lhsRefCompletion
	}

	lhsRef := lhsRefCompletion.Value.(*JavaScriptValue)
	leftValCompletion := GetValue(runtime, lhsRef)
	if leftValCompletion.Type != Normal {
		return leftValCompletion
	}

	leftVal := leftValCompletion.Value.(*JavaScriptValue)

	switch opType {
	case lexer.AndAssignment, lexer.OrAssignment:
		leftValBooleanCompletion := ToBoolean(leftVal)
		if leftValBooleanCompletion.Type != Normal {
			return leftValBooleanCompletion
		}

		leftValBoolean := leftValBooleanCompletion.Value.(*JavaScriptValue)
		leftValBooleanValue := leftValBoolean.Value.(*Boolean).Value

		// Early return when doing an AND assignment and the left value is a falsy value.
		if opType == lexer.AndAssignment && !leftValBooleanValue {
			return NewNormalCompletion(leftValBoolean)
		}

		// Early return when doing an OR assignment and the left value is a truthy value.
		if leftValBooleanValue {
			return NewNormalCompletion(leftValBoolean)
		}
	case lexer.NullishCoalescingAssignment:
		// Early return when doing a nullish coalescing assignment and the left value is not undefined or null.
		if leftVal.Type != TypeUndefined && leftVal.Type != TypeNull {
			return NewNormalCompletion(leftVal)
		}
	default:
		panic("Unexpected logical assignment operator.")
	}

	// TODO: Check if anon function definition, if so do something different according to the spec.

	rhsRefCompletion := Evaluate(runtime, rhsNode)
	if rhsRefCompletion.Type != Normal {
		return rhsRefCompletion
	}

	rhsRef := rhsRefCompletion.Value.(*JavaScriptValue)

	rhsValCompletion := GetValue(runtime, rhsRef)
	if rhsValCompletion.Type != Normal {
		return rhsValCompletion
	}

	rhsVal := rhsValCompletion.Value.(*JavaScriptValue)

	completion := PutValue(runtime, lhsRef, rhsVal)
	if completion.Type != Normal {
		return completion
	}

	return rhsValCompletion
}
