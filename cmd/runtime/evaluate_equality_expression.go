package runtime

import (
	"zbrannelly.dev/go-js/cmd/lexer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateEqualityExpression(runtime *Runtime, equalityExpression *ast.EqualityExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, equalityExpression.GetLeft())
	if lRefCompletion.Type == Throw {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	rRefCompletion := Evaluate(runtime, equalityExpression.GetRight())
	if rRefCompletion.Type == Throw {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)

	switch equalityExpression.GetOperator().Type {
	case lexer.Equal:
		return IsLooselyEqual(runtime, lRef, rRef)
	case lexer.NotEqual:
		return NegateBooleanValue(IsLooselyEqual(runtime, lRef, rRef))
	case lexer.StrictEqual:
		return IsStrictlyEqual(runtime, lRef, rRef)
	case lexer.StrictNotEqual:
		return NegateBooleanValue(IsStrictlyEqual(runtime, lRef, rRef))
	}

	panic("Unexpected equality operator.")
}
