package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateNullLiteral(runtime *Runtime, nullLiteral *ast.BasicNode) *Completion {
	return NewNormalCompletion(NewNullValue())
}
