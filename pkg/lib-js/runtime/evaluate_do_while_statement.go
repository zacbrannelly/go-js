package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

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

		// Get the value of the condition (resolve any references).
		expressionValueCompletion := GetValue(expressionCompletion.Value.(*JavaScriptValue))
		if expressionValueCompletion.Type != Normal {
			return expressionValueCompletion
		}

		// Convert the condition to a boolean.
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
	}
}
