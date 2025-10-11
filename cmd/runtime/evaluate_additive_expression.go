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

	return ApplyStringOrNumericBinaryOperation(runtime, leftVal, additiveExpression.Operator.Value, rightVal)
}
