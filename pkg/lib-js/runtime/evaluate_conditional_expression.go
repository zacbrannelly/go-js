package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateConditionalExpression(runtime *Runtime, conditionalExpression *ast.ConditionalExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, conditionalExpression.GetCondition())
	if lRefCompletion.Type != Normal {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	conditionValCompletion := GetValue(lRef)
	if conditionValCompletion.Type != Normal {
		return conditionValCompletion
	}

	conditionVal := conditionValCompletion.Value.(*JavaScriptValue)

	conditionBoolValCompletion := ToBoolean(conditionVal)
	if conditionBoolValCompletion.Type != Normal {
		return conditionBoolValCompletion
	}

	conditionBoolVal := conditionBoolValCompletion.Value.(*JavaScriptValue)
	conditionBoolValue := conditionBoolVal.Value.(*Boolean).Value

	var evalNode ast.Node
	if conditionBoolValue {
		evalNode = conditionalExpression.GetTrueExpr()
	} else {
		evalNode = conditionalExpression.GetFalseExpr()
	}

	evalCompletion := Evaluate(runtime, evalNode)
	if evalCompletion.Type != Normal {
		return evalCompletion
	}

	evalRef := evalCompletion.Value.(*JavaScriptValue)
	return GetValue(evalRef)
}
