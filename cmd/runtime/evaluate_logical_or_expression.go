package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateLogicalORExpression(runtime *Runtime, logicalORExpression *ast.LogicalORExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, logicalORExpression.GetLeft())
	if lRefCompletion.Type == Throw {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(lRef)
	if lValCompletion.Type == Throw {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	lValBooleanCompletion := ToBoolean(runtime, lVal)
	if lValBooleanCompletion.Type == Throw {
		return lValBooleanCompletion
	}

	lValBoolean := lValBooleanCompletion.Value.(*JavaScriptValue)
	lValBooleanValue := lValBoolean.Value.(*Boolean).Value

	if lValBooleanValue {
		return lValCompletion
	}

	rRefCompletion := Evaluate(runtime, logicalORExpression.GetRight())
	if rRefCompletion.Type == Throw {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)
	return GetValue(rRef)
}
