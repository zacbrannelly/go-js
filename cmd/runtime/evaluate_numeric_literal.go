package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateNumericLiteral(runtime *Runtime, numericLiteral *ast.NumericLiteralNode) *Completion {
	return NewNormalCompletion(NewNumberValue(numericLiteral.Value, false))
}
