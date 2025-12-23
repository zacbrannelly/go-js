package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateThrowStatement(runtime *Runtime, throwStatement *ast.ThrowStatementNode) *Completion {
	completion := Evaluate(runtime, throwStatement.GetExpression())
	if completion.Type != Normal {
		return completion
	}

	maybeRef := completion.Value.(*JavaScriptValue)
	completion = GetValue(runtime, maybeRef)
	if completion.Type != Normal {
		return completion
	}

	return NewThrowCompletion(completion.Value.(*JavaScriptValue))
}
