package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateAdditiveExpression(runtime *Runtime, additiveExpression *ast.AdditiveExpressionNode) *Completion {
	leftRef := Evaluate(runtime, additiveExpression.GetLeft())
	rightRef := Evaluate(runtime, additiveExpression.GetRight())

	if leftRef.Type == Throw {
		return leftRef
	}

	if rightRef.Type == Throw {
		return rightRef
	}

	// TODO: Handle reference types.

	leftVal := leftRef.Value.(*JavaScriptValue)
	rightVal := rightRef.Value.(*JavaScriptValue)

	return ApplyStringOrNumericBinaryOperation(runtime, leftVal, additiveExpression.Operator.Value, rightVal)
}
