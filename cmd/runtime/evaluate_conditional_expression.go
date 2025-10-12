package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateConditionalExpression(runtime *Runtime, conditionalExpression *ast.ConditionalExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, conditionalExpression.GetCondition())
	if lRefCompletion.Type == Throw {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	conditionValCompletion := GetValue(lRef)
	if conditionValCompletion.Type == Throw {
		return conditionValCompletion
	}

	conditionVal := conditionValCompletion.Value.(*JavaScriptValue)

	conditionBoolValCompletion := ToBoolean(runtime, conditionVal)
	if conditionBoolValCompletion.Type == Throw {
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
	if evalCompletion.Type == Throw {
		return evalCompletion
	}

	evalRef := evalCompletion.Value.(*JavaScriptValue)
	return GetValue(evalRef)
}
