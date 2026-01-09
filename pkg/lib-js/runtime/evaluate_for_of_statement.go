package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateForOfStatement(runtime *Runtime, forOfStatement *ast.ForOfStatementNode) *Completion {
	completion := ForOfStatementLoopEvaluation(runtime, forOfStatement)
	return LabelledEvaluation(runtime, completion)
}

func ForOfStatementLoopEvaluation(runtime *Runtime, forOfStatement *ast.ForOfStatementNode) *Completion {
	uninitializedBoundNames := []string{}
	if forOfStatement.GetTarget().GetNodeType() == ast.LexicalBinding {
		uninitializedBoundNames = BoundNames(forOfStatement.GetTarget())
	}

	completion := ForOfHeadEvaluation(
		runtime,
		uninitializedBoundNames,
		forOfStatement.GetIterable(),
		forOfStatement.Await,
	)
	if completion.Type != Normal {
		return completion
	}

	iterator := completion.Value.(*Iterator)
	return ForOfBodyEvaluation(runtime, forOfStatement, iterator, forOfStatement.Await)
}

func ForOfHeadEvaluation(
	runtime *Runtime,
	uninitializedBoundNames []string,
	expression ast.Node,
	await bool,
) *Completion {
	oldEnv := runtime.GetRunningExecutionContext().LexicalEnvironment

	if len(uninitializedBoundNames) > 0 {
		newEnv := NewDeclarativeEnvironment(oldEnv)
		for _, boundName := range uninitializedBoundNames {
			newEnv.CreateMutableBinding(runtime, boundName, false)
		}
		runtime.GetRunningExecutionContext().LexicalEnvironment = newEnv
	}

	completion := Evaluate(runtime, expression)
	runtime.GetRunningExecutionContext().LexicalEnvironment = oldEnv

	if completion.Type != Normal {
		return completion
	}

	completion = GetValue(runtime, completion.Value.(*JavaScriptValue))
	if completion.Type != Normal {
		return completion
	}

	expressionValue := completion.Value.(*JavaScriptValue)

	var iteratorKind IteratorKind
	if await {
		iteratorKind = IteratorKindAsync
	} else {
		iteratorKind = IteratorKindSync
	}

	return GetIterator(runtime, expressionValue, iteratorKind)
}

func ForOfBodyEvaluation(
	runtime *Runtime,
	forOfStatement *ast.ForOfStatementNode,
	iterator *Iterator,
	await bool,
) *Completion {
	if await {
		panic("TODO: Implement ForOfBodyEvaluation for await.")
	}

	oldEnv := runtime.GetRunningExecutionContext().LexicalEnvironment

	value := NewUndefinedValue()

	isDestructuring := false
	if forOfStatement.GetTarget().GetNodeType() == ast.VariableDeclaration {
		// var ForBinding
		declaration := forOfStatement.GetTarget().GetChildren()[0]
		isDestructuring = declaration.GetNodeType() == ast.ArrayBindingPattern
		isDestructuring = isDestructuring || declaration.GetNodeType() == ast.ObjectBindingPattern
	} else if forOfStatement.GetNodeType() == ast.LexicalBinding {
		// ForDeclaration
		binding := forOfStatement.GetTarget().(*ast.LexicalBindingNode)
		isDestructuring = binding.GetTarget().GetNodeType() == ast.ArrayBindingPattern
		isDestructuring = isDestructuring || binding.GetTarget().GetNodeType() == ast.ObjectBindingPattern
	} else {
		// LeftHandSideExpression
		isDestructuring = forOfStatement.GetTarget().GetNodeType() == ast.ArrayLiteral
		isDestructuring = isDestructuring || forOfStatement.GetTarget().GetNodeType() == ast.ObjectLiteral
	}

	for {
		completion := Call(runtime, iterator.Next, iterator.Iterator, []*JavaScriptValue{})
		if completion.Type != Normal {
			return completion
		}

		if await {
			panic("TODO: Call Await() here.")
		}

		nextResult := completion.Value.(*JavaScriptValue)
		if nextResult.Type != TypeObject {
			panic("Assert failed: ForInIterator.Next returned a non-object")
		}

		completion = IteratorComplete(runtime, nextResult)
		if completion.Type != Normal {
			return completion
		}

		done := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
		if done {
			return NewNormalCompletion(value)
		}

		completion = IteratorValue(runtime, nextResult)
		if completion.Type != Normal {
			return completion
		}

		nextValue := completion.Value.(*JavaScriptValue)

		var status *Completion

		if forOfStatement.GetTarget().GetNodeType() != ast.LexicalBinding {
			if isDestructuring {
				panic("TODO: Support destructuring.")
			} else {
				if forOfStatement.GetTarget().GetNodeType() == ast.VariableDeclaration {
					declaration := forOfStatement.GetTarget().GetChildren()[0]
					completion = Evaluate(runtime, declaration)
				} else {
					completion = Evaluate(runtime, forOfStatement.GetTarget())
				}

				if completion.Type != Normal {
					status = completion
				} else {
					lhsRef := completion.Value.(*JavaScriptValue)
					status = PutValue(runtime, lhsRef, nextValue)
				}
			}
		} else {
			iterationEnv := NewDeclarativeEnvironment(oldEnv)

			lexicalBinding := forOfStatement.GetTarget().(*ast.LexicalBindingNode)
			forBinding := lexicalBinding.GetTarget()
			boundNames := BoundNames(forBinding)
			for _, name := range boundNames {
				if lexicalBinding.Const {
					iterationEnv.CreateImmutableBinding(runtime, name, true)
				} else {
					iterationEnv.CreateMutableBinding(runtime, name, false)
				}
			}

			runtime.GetRunningExecutionContext().LexicalEnvironment = iterationEnv

			if isDestructuring {
				panic("TODO: Support destructuring.")
			} else {
				if len(boundNames) > 1 {
					panic("Assert failed: Non-destructuring ForDeclaration with lexical binding must have exactly one bound name.")
				}

				isStrictMode := analyzer.IsStrictMode(lexicalBinding)
				completion = ResolveBindingFromCurrentContext(boundNames[0], runtime, isStrictMode)

				if completion.Type != Normal {
					panic("Assert failed: ResolveBindingFromCurrentContext threw an unexpected error in ForInBodyEvaluation.")
				}

				reference := completion.Value.(*JavaScriptValue).Value.(*Reference)
				status = reference.InitializeReferencedBinding(runtime, nextValue)
			}
		}

		if status.Type != Normal {
			runtime.GetRunningExecutionContext().LexicalEnvironment = oldEnv
			return IteratorClose(runtime, iterator, status)
		}

		completion = Evaluate(runtime, forOfStatement.GetBody())
		runtime.GetRunningExecutionContext().LexicalEnvironment = oldEnv

		if !LoopContinues(runtime, completion) {
			if completion.Value == nil {
				completion.Value = value
			}

			if await {
				panic("TODO: Return AsyncIteratorClose here.")
			}

			return IteratorClose(runtime, iterator, completion)
		}

		if completion.Value != nil {
			value = completion.Value.(*JavaScriptValue)
		}
	}
}
