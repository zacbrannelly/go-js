package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateDoWhileStatement(runtime *Runtime, doWhileStatement *ast.DoWhileStatementNode) *Completion {
	var value *JavaScriptValue = NewUndefinedValue()
	for {
		statementCompletion := Evaluate(runtime, doWhileStatement.GetStatement())
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

		// Evaluate the condition.
		expressionCompletion := Evaluate(runtime, doWhileStatement.GetCondition())
		if expressionCompletion.Type != Normal {
			return expressionCompletion
		}

		// Convert the condition to a boolean.
		expressionValue := expressionCompletion.Value.(*JavaScriptValue)
		expressionBoolValueCompletion := ToBoolean(runtime, expressionValue)
		if expressionBoolValueCompletion.Type != Normal {
			return expressionBoolValueCompletion
		}

		// Check if the condition is falsy. Return the latest value if it is.
		expressionBoolValue := expressionBoolValueCompletion.Value.(*JavaScriptValue)
		if !expressionBoolValue.Value.(*Boolean).Value {
			return NewNormalCompletion(value)
		}
	}
}
