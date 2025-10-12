package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateBooleanLiteral(runtime *Runtime, booleanLiteral *ast.BooleanLiteralNode) *Completion {
	return NewNormalCompletion(NewBooleanValue(booleanLiteral.Value))
}
