package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateCoverParenthesizedExpressionAndArrowParameterList(runtime *Runtime, cover *ast.BasicNode) *Completion {
	return Evaluate(runtime, cover.GetChildren()[0])
}
