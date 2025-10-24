package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateBreakStatement(runtime *Runtime, breakStatement *ast.BreakStatementNode) *Completion {
	if label, ok := breakStatement.GetLabel().(*ast.LabelIdentifierNode); ok {
		return NewBreakCompletion(label.Identifier)
	}

	return NewBreakCompletion("")
}
