package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateStatementList(runtime *Runtime, statementList *ast.StatementListNode) *Completion {
	var completion *Completion

	var lastValue any
	for _, statement := range statementList.GetChildren() {
		completion = Evaluate(runtime, statement)
		if completion.Type == Throw {
			return completion
		}

		if completion.Value != nil {
			lastValue = completion.Value
		}
	}

	if lastValue != nil {
		return NewNormalCompletion(lastValue)
	}

	return NewUnusedCompletion()
}
