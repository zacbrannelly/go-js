package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func CompileIfStatement(runtime *Runtime, ifStatement *ast.IfStatementNode) []Instruction {
	instructions := []Instruction{}

	// Evaluate the condition.
	instructions = append(instructions, EmitEvaluateExpression(ifStatement.GetCondition()))

	// Resolve the condition to a boolean value.
	resolveCondition := EmitEvaluateNativeCallback(func(runtime *Runtime, vm *ExecutionVM) *Completion {
		completion := vm.PopEvaluationStack()
		if completion.Type != Normal {
			return completion
		}

		conditionRef := completion.Value.(*JavaScriptValue)

		completion = GetValue(runtime, conditionRef)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)
		return ToBoolean(value)
	})
	instructions = append(instructions, resolveCondition)

	// Skip the true statement if the condition is false.
	instructions = append(instructions, EmitJumpIfFalse(1))

	// Evaluate the true statement.
	instructions = append(instructions, EmitEvaluateExpression(ifStatement.GetTrueStatement()))

	if ifStatement.GetElseStatement() != nil {
		instructions = append(instructions, EmitEvaluateExpression(ifStatement.GetElseStatement()))
	}

	// Make sure the completion is an undefined value if there is no [[Value]] in the last completion.
	makeUndefined := EmitEvaluateNativeCallback(func(runtime *Runtime, vm *ExecutionVM) *Completion {
		completion := vm.PeekEvaluationStack()
		if completion == nil {
			return NewNormalCompletion(NewUndefinedValue())
		}

		completion = vm.PopEvaluationStack()
		if completion.Type != Normal {
			return completion
		}

		if completion.Value == nil {
			return NewNormalCompletion(NewUndefinedValue())
		}

		return completion
	})
	instructions = append(instructions, makeUndefined)

	return instructions
}

func EvaluateIfStatement(runtime *Runtime, ifStatement *ast.IfStatementNode) *Completion {
	conditionRefCompletion := Evaluate(runtime, ifStatement.GetCondition())
	if conditionRefCompletion.Type != Normal {
		return conditionRefCompletion
	}

	conditionRef := conditionRefCompletion.Value.(*JavaScriptValue)

	conditionValCompletion := GetValue(runtime, conditionRef)
	if conditionValCompletion.Type != Normal {
		return conditionValCompletion
	}

	conditionVal := conditionValCompletion.Value.(*JavaScriptValue)

	conditionBoolValCompletion := ToBoolean(conditionVal)
	if conditionBoolValCompletion.Type != Normal {
		return conditionBoolValCompletion
	}

	conditionBoolVal := conditionBoolValCompletion.Value.(*JavaScriptValue)
	conditionBoolValue := conditionBoolVal.Value.(*Boolean).Value

	var statementCompletion *Completion
	if conditionBoolValue {
		statementCompletion = Evaluate(runtime, ifStatement.GetTrueStatement())
	} else if ifStatement.GetElseStatement() != nil {
		statementCompletion = Evaluate(runtime, ifStatement.GetElseStatement())
	}

	if statementCompletion == nil {
		return NewNormalCompletion(NewUndefinedValue())
	}

	if statementCompletion.Type != Normal {
		return statementCompletion
	}

	// TODO: Does this model the UpdateEmpty operation in the spec?
	if statementCompletion.Value == nil {
		return NewNormalCompletion(NewUndefinedValue())
	}

	return statementCompletion
}
