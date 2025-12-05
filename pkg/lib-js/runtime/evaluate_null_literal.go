package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateNullLiteral(runtime *Runtime, nullLiteral *ast.BasicNode) *Completion {
	return NewNormalCompletion(NewNullValue())
}
