package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func CompileStatementList(runtime *Runtime, statementList *ast.StatementListNode) []Instruction {
	instructions := []Instruction{}

	for _, statement := range statementList.GetChildren() {
		instructions = append(instructions, Compile(runtime, statement)...)
	}

	return instructions
}

func EvaluateStatementList(runtime *Runtime, statementList *ast.StatementListNode) *Completion {
	var completion *Completion

	var lastValue any
	for _, statement := range statementList.GetChildren() {
		completion = Evaluate(runtime, statement)
		if completion.Type != Normal {
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
