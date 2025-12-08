package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateCoalesceExpression(runtime *Runtime, coalesceExpression *ast.CoalesceExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, coalesceExpression.GetLeft())
	if lRefCompletion.Type != Normal {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(runtime, lRef)
	if lValCompletion.Type != Normal {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	// Only evaluate right side if left is null or undefined
	if lVal.Type == TypeNull || lVal.Type == TypeUndefined {
		rRefCompletion := Evaluate(runtime, coalesceExpression.GetRight())
		if rRefCompletion.Type != Normal {
			return rRefCompletion
		}

		rRef := rRefCompletion.Value.(*JavaScriptValue)
		return GetValue(runtime, rRef)
	}

	return lValCompletion
}
