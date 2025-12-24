package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateTryStatement(runtime *Runtime, tryStatement *ast.TryStatementNode) *Completion {
	// Evaluate the try block.
	completion := Evaluate(runtime, tryStatement.GetBlock())

	// Evaluate the catch block if the try block threw.
	if completion.Type == Throw && tryStatement.GetCatch() != nil {
		blockValue := completion.Value.(*JavaScriptValue)
		completion = CatchClauseEvaluation(runtime, tryStatement.GetCatch().(*ast.CatchNode), blockValue)
	}

	// Evaluate the finally block.
	if tryStatement.GetFinally() != nil {
		finallyCompletion := Evaluate(runtime, tryStatement.GetFinally())
		if finallyCompletion.Type != Normal {
			completion = finallyCompletion
		}
	}

	if completion.Type == Normal && completion.Value == nil {
		completion.Value = NewUndefinedValue()
	}

	return completion
}

func CatchClauseEvaluation(runtime *Runtime, catch *ast.CatchNode, thrownValue *JavaScriptValue) *Completion {
	if catch.GetTarget() == nil {
		return Evaluate(runtime, catch.GetBlock())
	}

	oldEnv := runtime.GetRunningExecutionContext().LexicalEnvironment
	catchEnv := NewDeclarativeEnvironment(oldEnv)

	for _, argName := range BoundNames(catch.GetTarget()) {
		completion := catchEnv.CreateMutableBinding(runtime, argName, false)
		if completion.Type != Normal {
			panic("Assert failed: CreateMutableBinding threw an unexpected error in CatchClauseEvaluation.")
		}
	}

	runtime.GetRunningExecutionContext().LexicalEnvironment = catchEnv

	completion := BindingInitialization(runtime, catch.GetTarget(), thrownValue, catchEnv)
	if completion.Type != Normal {
		runtime.GetRunningExecutionContext().LexicalEnvironment = oldEnv
		return completion
	}

	completion = Evaluate(runtime, catch.GetBlock())
	runtime.GetRunningExecutionContext().LexicalEnvironment = oldEnv

	return completion
}
