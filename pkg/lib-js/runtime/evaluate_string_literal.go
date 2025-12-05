package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateStringLiteral(runtime *Runtime, stringLiteral *ast.StringLiteralNode) *Completion {
	return NewNormalCompletion(NewStringValue(stringLiteral.Value))
}
