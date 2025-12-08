package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateEqualityExpression(runtime *Runtime, equalityExpression *ast.EqualityExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, equalityExpression.GetLeft())
	if lRefCompletion.Type != Normal {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(runtime, lRef)
	if lValCompletion.Type != Normal {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	rRefCompletion := Evaluate(runtime, equalityExpression.GetRight())
	if rRefCompletion.Type != Normal {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)

	rValCompletion := GetValue(runtime, rRef)
	if rValCompletion.Type != Normal {
		return rValCompletion
	}

	rVal := rValCompletion.Value.(*JavaScriptValue)

	switch equalityExpression.GetOperator().Type {
	case lexer.Equal:
		return IsLooselyEqual(lVal, rVal)
	case lexer.NotEqual:
		return NegateBooleanValue(IsLooselyEqual(lVal, rVal))
	case lexer.StrictEqual:
		return IsStrictlyEqual(lVal, rVal)
	case lexer.StrictNotEqual:
		return NegateBooleanValue(IsStrictlyEqual(lVal, rVal))
	}

	panic("Unexpected equality operator.")
}
