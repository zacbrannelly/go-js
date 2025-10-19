package runtime

import (
	"zbrannelly.dev/go-js/cmd/lexer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateEqualityExpression(runtime *Runtime, equalityExpression *ast.EqualityExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, equalityExpression.GetLeft())
	if lRefCompletion.Type != Normal {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(lRef)
	if lValCompletion.Type != Normal {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	rRefCompletion := Evaluate(runtime, equalityExpression.GetRight())
	if rRefCompletion.Type != Normal {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)

	rValCompletion := GetValue(rRef)
	if rValCompletion.Type != Normal {
		return rValCompletion
	}

	rVal := rValCompletion.Value.(*JavaScriptValue)

	switch equalityExpression.GetOperator().Type {
	case lexer.Equal:
		return IsLooselyEqual(runtime, lVal, rVal)
	case lexer.NotEqual:
		return NegateBooleanValue(IsLooselyEqual(runtime, lVal, rVal))
	case lexer.StrictEqual:
		return IsStrictlyEqual(runtime, lVal, rVal)
	case lexer.StrictNotEqual:
		return NegateBooleanValue(IsStrictlyEqual(runtime, lVal, rVal))
	}

	panic("Unexpected equality operator.")
}
