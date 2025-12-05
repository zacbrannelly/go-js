package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateContinueStatement(runtime *Runtime, continueStatement *ast.ContinueStatementNode) *Completion {
	if label, ok := continueStatement.GetLabel().(*ast.LabelIdentifierNode); ok {
		return NewContinueCompletion(label.Identifier)
	}

	return NewContinueCompletion("")
}
