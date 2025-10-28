package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateWhileStatement(runtime *Runtime, whileStatement *ast.WhileStatementNode) *Completion {
	var value *JavaScriptValue = NewUndefinedValue()
	for {
		// Evaluate the condition.
		expressionCompletion := Evaluate(runtime, whileStatement.GetCondition())
		if expressionCompletion.Type != Normal {
			return expressionCompletion
		}

		// Get the value of the condition (resolve any references).
		expressionValueCompletion := GetValue(expressionCompletion.Value.(*JavaScriptValue))
		if expressionValueCompletion.Type != Normal {
			return expressionValueCompletion
		}

		// Convert the condition to a boolean
		expressionValue := expressionValueCompletion.Value.(*JavaScriptValue)
		expressionBoolValueCompletion := ToBoolean(expressionValue)
		if expressionBoolValueCompletion.Type != Normal {
			return expressionBoolValueCompletion
		}

		// Check if the condition is falsy. Return the latest value if it is.
		expressionBoolValue := expressionBoolValueCompletion.Value.(*JavaScriptValue)
		if !expressionBoolValue.Value.(*Boolean).Value {
			return NewNormalCompletion(value)
		}

		statementCompletion := Evaluate(runtime, whileStatement.GetStatement())
		if !LoopContinues(runtime, statementCompletion) {
			if statementCompletion.Type != Normal {
				return statementCompletion
			}

			if statementCompletion.Value == nil {
				statementCompletion.Value = value
			}

			return statementCompletion
		}

		if statementCompletion.Value != nil {
			// TODO: Check type.
			value = statementCompletion.Value.(*JavaScriptValue)
		}
	}
}
