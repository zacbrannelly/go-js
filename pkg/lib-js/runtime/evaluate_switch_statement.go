package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateSwitchStatement(runtime *Runtime, switchStatement *ast.SwitchStatementNode) *Completion {
	// Evaluate the switch target value.
	completion := Evaluate(runtime, switchStatement.GetTarget())
	if completion.Type != Normal {
		return completion
	}

	// Get the value of the switch target (in the case of a reference).
	maybeRef := completion.Value.(*JavaScriptValue)

	completion = GetValue(runtime, maybeRef)
	if completion.Type != Normal {
		return completion
	}

	targetValue := completion.Value.(*JavaScriptValue)

	oldEnv := runtime.GetRunningExecutionContext().LexicalEnvironment
	blockEnv := NewDeclarativeEnvironment(oldEnv)

	// Instantiate the block declarations.
	BlockDeclarationInstantiation(runtime, switchStatement, blockEnv)

	runtime.GetRunningExecutionContext().LexicalEnvironment = blockEnv

	// Evaluate the switch block.
	completion = CaseBlockEvaluation(runtime, switchStatement, targetValue)

	runtime.GetRunningExecutionContext().LexicalEnvironment = oldEnv
	return completion
}

func CaseBlockEvaluation(
	runtime *Runtime,
	switchStatement *ast.SwitchStatementNode,
	switchValue *JavaScriptValue,
) *Completion {
	if len(switchStatement.GetChildren()) == 0 {
		return NewNormalCompletion(NewUndefinedValue())
	}

	var v any = NewUndefinedValue()

	found := false
	defaultIdx := -1

	for idx, child := range switchStatement.GetChildren() {
		if child.GetNodeType() == ast.SwitchCase {
			if !found {
				completion := CaseClauseIsSelected(runtime, child.(*ast.SwitchCaseNode), switchValue)
				if completion.Type != Normal {
					return completion
				}

				found = completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
			}
		}

		if found && child.GetNodeType() == ast.StatementList {
			completion := EvaluateStatementList(runtime, child.(*ast.StatementListNode))

			if completion.Value != nil {
				v = completion.Value
			}

			if completion.Type != Normal {
				if completion.Value == nil {
					completion.Value = v
				}

				return completion
			}
		}

		if child.GetNodeType() == ast.SwitchDefault {
			defaultIdx = idx
			break
		}
	}

	var defaultBody *ast.StatementListNode

	if defaultIdx != -1 && defaultIdx+1 < len(switchStatement.GetChildren()) {
		if switchStatement.GetChildren()[defaultIdx+1].GetNodeType() == ast.StatementList {
			defaultBody = switchStatement.GetChildren()[defaultIdx+1].(*ast.StatementListNode)
		}
	}

	foundInB := false

	stepSize := 2
	if defaultBody == nil {
		stepSize = 1
	}

	for idx := defaultIdx + stepSize; defaultIdx != -1 && idx < len(switchStatement.GetChildren()); idx++ {
		child := switchStatement.GetChildren()[idx]

		if child.GetNodeType() == ast.SwitchCase {
			if !foundInB {
				completion := CaseClauseIsSelected(runtime, child.(*ast.SwitchCaseNode), switchValue)
				if completion.Type != Normal {
					return completion
				}

				foundInB = completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
			}
		}

		if foundInB && child.GetNodeType() == ast.StatementList {
			completion := EvaluateStatementList(runtime, child.(*ast.StatementListNode))
			if completion.Value != nil {
				v = completion.Value
			}

			if completion.Type != Normal {
				if completion.Value == nil {
					completion.Value = v
				}

				return completion
			}
		}
	}

	if foundInB {
		return NewNormalCompletion(v)
	}

	if defaultBody != nil {
		completion := EvaluateStatementList(runtime, defaultBody)
		if completion.Value != nil {
			v = completion.Value
		}

		if completion.Type != Normal {
			if completion.Value == nil {
				completion.Value = v
			}

			return completion
		}
	}

	for idx := defaultIdx + stepSize; defaultIdx != -1 && idx < len(switchStatement.GetChildren()); idx++ {
		child := switchStatement.GetChildren()[idx]
		if child.GetNodeType() == ast.StatementList {
			completion := EvaluateStatementList(runtime, child.(*ast.StatementListNode))
			if completion.Value != nil {
				v = completion.Value.(*JavaScriptValue)
			}

			if completion.Type != Normal {
				if completion.Value == nil {
					completion.Value = v
				}

				return completion
			}
		}
	}

	return NewNormalCompletion(v)
}

func CaseClauseIsSelected(
	runtime *Runtime,
	caseClause *ast.SwitchCaseNode,
	switchValue *JavaScriptValue,
) *Completion {
	completion := Evaluate(runtime, caseClause.GetExpression())
	if completion.Type != Normal {
		return completion
	}

	maybeRef := completion.Value.(*JavaScriptValue)

	completion = GetValue(runtime, maybeRef)
	if completion.Type != Normal {
		return completion
	}

	expressionValue := completion.Value.(*JavaScriptValue)
	return IsStrictlyEqual(switchValue, expressionValue)
}
