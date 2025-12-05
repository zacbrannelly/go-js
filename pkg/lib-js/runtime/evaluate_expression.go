package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateExpression(runtime *Runtime, expression *ast.ExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, expression.GetLeft())
	if lRefCompletion.Type != Normal {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(lRef)
	if lValCompletion.Type != Normal {
		return lValCompletion
	}

	rRefCompletion := Evaluate(runtime, expression.GetRight())
	if rRefCompletion.Type != Normal {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)

	return GetValue(rRef)
}
