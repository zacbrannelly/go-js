package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateForInStatement(runtime *Runtime, forInStatement *ast.ForInStatementNode) *Completion {
	completion := ForInStatementLoopEvaluation(runtime, forInStatement)
	return LabelledEvaluation(runtime, completion)
}

func ForInStatementLoopEvaluation(runtime *Runtime, forInStatement *ast.ForInStatementNode) *Completion {
	uninitializedBoundNames := []string{}
	if forInStatement.GetTarget().GetNodeType() == ast.LexicalBinding {
		uninitializedBoundNames = BoundNames(forInStatement.GetTarget())
	}

	completion := ForInHeadEvaluation(runtime, uninitializedBoundNames, forInStatement.GetIterable())
	if completion.Type != Normal {
		return completion
	}

	iterator := completion.Value.(*ForInIterator)
	return ForInBodyEvaluation(runtime, forInStatement, iterator)
}

func ForInHeadEvaluation(
	runtime *Runtime,
	uninitializedBoundNames []string,
	expression ast.Node,
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

	if expressionValue.Type == TypeNull || expressionValue.Type == TypeUndefined {
		return NewEmptyBreakCompletion()
	}

	completion = ToObject(runtime, expressionValue)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	iterator := NewForInIterator(object)
	return NewNormalCompletion(iterator)
}

func ForInBodyEvaluation(
	runtime *Runtime,
	forInStatement *ast.ForInStatementNode,
	iterator *ForInIterator,
) *Completion {
	oldEnv := runtime.GetRunningExecutionContext().LexicalEnvironment

	value := NewUndefinedValue()

	isDestructuring := false
	if forInStatement.GetTarget().GetNodeType() == ast.VariableDeclaration {
		// var ForBinding
		declaration := forInStatement.GetTarget().GetChildren()[0]
		isDestructuring = declaration.GetNodeType() == ast.ArrayBindingPattern
		isDestructuring = isDestructuring || declaration.GetNodeType() == ast.ObjectBindingPattern
	} else if forInStatement.GetNodeType() == ast.LexicalBinding {
		// ForDeclaration
		binding := forInStatement.GetTarget().(*ast.LexicalBindingNode)
		isDestructuring = binding.GetTarget().GetNodeType() == ast.ArrayBindingPattern
		isDestructuring = isDestructuring || binding.GetTarget().GetNodeType() == ast.ObjectBindingPattern
	} else {
		// LeftHandSideExpression
		isDestructuring = forInStatement.GetTarget().GetNodeType() == ast.ArrayLiteral
		isDestructuring = isDestructuring || forInStatement.GetTarget().GetNodeType() == ast.ObjectLiteral
	}

	for {
		completion := iterator.Next(runtime)
		if completion.Type != Normal {
			return completion
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

		if forInStatement.GetTarget().GetNodeType() != ast.LexicalBinding {
			if isDestructuring {
				panic("TODO: Support destructuring.")
			} else {
				if forInStatement.GetTarget().GetNodeType() == ast.VariableDeclaration {
					declaration := forInStatement.GetTarget().GetChildren()[0]
					completion = Evaluate(runtime, declaration)
				} else {
					completion = Evaluate(runtime, forInStatement.GetTarget())
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

			lexicalBinding := forInStatement.GetTarget().(*ast.LexicalBindingNode)
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
			return status
		}

		completion = Evaluate(runtime, forInStatement.GetBody())
		runtime.GetRunningExecutionContext().LexicalEnvironment = oldEnv

		if !LoopContinues(runtime, completion) {
			if completion.Value == nil {
				completion.Value = value
			}

			return completion
		}

		if completion.Value != nil {
			value = completion.Value.(*JavaScriptValue)
		}
	}
}
