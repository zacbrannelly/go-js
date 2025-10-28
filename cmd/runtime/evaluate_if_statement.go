package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateIfStatement(runtime *Runtime, ifStatement *ast.IfStatementNode) *Completion {
	conditionRefCompletion := Evaluate(runtime, ifStatement.GetCondition())
	if conditionRefCompletion.Type != Normal {
		return conditionRefCompletion
	}

	conditionRef := conditionRefCompletion.Value.(*JavaScriptValue)

	conditionValCompletion := GetValue(conditionRef)
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
