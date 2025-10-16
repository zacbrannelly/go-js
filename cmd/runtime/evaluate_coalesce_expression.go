package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateCoalesceExpression(runtime *Runtime, coalesceExpression *ast.CoalesceExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, coalesceExpression.GetLeft())
	if lRefCompletion.Type == Throw {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(lRef)
	if lValCompletion.Type == Throw {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	// Only evaluate right side if left is null or undefined
	if lVal.Type == TypeNull || lVal.Type == TypeUndefined {
		rRefCompletion := Evaluate(runtime, coalesceExpression.GetRight())
		if rRefCompletion.Type == Throw {
			return rRefCompletion
		}

		rRef := rRefCompletion.Value.(*JavaScriptValue)
		return GetValue(rRef)
	}

	return lValCompletion
}
