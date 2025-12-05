package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateReturnStatement(runtime *Runtime, returnStatement *ast.ReturnStatementNode) *Completion {
	if returnStatement.GetValue() == nil {
		return NewReturnCompletion(NewUndefinedValue())
	}

	expressionRefCompletion := Evaluate(runtime, returnStatement.GetValue())
	if expressionRefCompletion.Type != Normal {
		return expressionRefCompletion
	}

	expressionRef := expressionRefCompletion.Value.(*JavaScriptValue)
	expressionValCompletion := GetValue(expressionRef)
	if expressionValCompletion.Type != Normal {
		return expressionValCompletion
	}

	// TODO: Implement GetGeneratorKind(), check if its ASYNC, if so, set expressionVal to Await(expressionVal)

	return NewReturnCompletion(expressionValCompletion.Value.(*JavaScriptValue))
}
