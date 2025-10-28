package runtime

import "zbrannelly.dev/go-js/cmd/parser/ast"

func EvaluateForStatement(runtime *Runtime, forStatement *ast.ForStatementNode) *Completion {
	// Handle lexical declarations for the initializer differently.
	if forStatement.GetInitializer() != nil && forStatement.GetInitializer().GetNodeType() == ast.LexicalDeclaration {
		return EvaluateForStatementWithLexicalDeclaration(runtime, forStatement)
	}

	// Evaluate the initializer.
	if forStatement.GetInitializer() != nil {
		expressionCompletion := Evaluate(runtime, forStatement.GetInitializer())
		if expressionCompletion.Type != Normal {
			return expressionCompletion
		}

		if forStatement.GetInitializer().GetNodeType() != ast.VariableDeclarationList {
			expressionValue := GetValue(expressionCompletion.Value.(*JavaScriptValue))
			if expressionValue.Type != Normal {
				return expressionValue
			}
		}
	}

	return ForBodyEvaluation(
		runtime,
		forStatement.GetCondition(),
		forStatement.GetUpdate(),
		forStatement.GetBody(),
		make([]string, 0),
	)
}

func EvaluateForStatementWithLexicalDeclaration(runtime *Runtime, forStatement *ast.ForStatementNode) *Completion {
	// Create new lexical environment for the loop.
	runningContext := runtime.GetRunningExecutionContext()
	oldEnv := runningContext.LexicalEnvironment
	loopEnv := NewDeclarativeEnvironment(oldEnv)

	lexicalDeclaration := forStatement.GetInitializer()
	isConst := lexicalDeclaration.GetChildren()[0].(*ast.LexicalBindingNode).Const

	// Get the bound names of the lexical declaration.
	boundNames := BoundNames(forStatement.GetInitializer())
	for _, name := range boundNames {
		if isConst {
			loopEnv.CreateImmutableBinding(name, true)
		} else {
			loopEnv.CreateMutableBinding(name, false)
		}
	}

	runningContext.LexicalEnvironment = loopEnv

	lexicalCompletion := Evaluate(runtime, lexicalDeclaration)
	if lexicalCompletion.Type != Normal {
		runningContext.LexicalEnvironment = oldEnv
		return lexicalCompletion
	}

	var perIterationLets []string = make([]string, 0)
	if !isConst {
		perIterationLets = append(perIterationLets, boundNames...)
	}

	bodyCompletion := ForBodyEvaluation(
		runtime,
		forStatement.GetCondition(),
		forStatement.GetUpdate(),
		forStatement.GetBody(),
		perIterationLets,
	)

	runningContext.LexicalEnvironment = oldEnv
	return bodyCompletion
}

func ForBodyEvaluation(runtime *Runtime, test ast.Node, increment ast.Node, body ast.Node, perIterationLets []string) *Completion {
	value := NewUndefinedValue()

	perIterationEnvCompletion := CreatePerIterationEnvironment(runtime, perIterationLets)
	if perIterationEnvCompletion.Type != Normal {
		return perIterationEnvCompletion
	}

	for {
		// Evaluate the test expression.
		if test != nil {
			testCompletion := Evaluate(runtime, test)
			if testCompletion.Type != Normal {
				return testCompletion
			}

			testValueCompletion := GetValue(testCompletion.Value.(*JavaScriptValue))
			if testValueCompletion.Type != Normal {
				return testValueCompletion
			}

			testValue := testValueCompletion.Value.(*JavaScriptValue)
			testBoolValueCompletion := ToBoolean(testValue)
			if testBoolValueCompletion.Type != Normal {
				return testBoolValueCompletion
			}

			testBoolValue := testBoolValueCompletion.Value.(*JavaScriptValue)
			if !testBoolValue.Value.(*Boolean).Value {
				return NewNormalCompletion(value)
			}
		}

		// Evaluate the body.
		resultCompletion := Evaluate(runtime, body)
		if !LoopContinues(runtime, resultCompletion) {
			if resultCompletion.Value == nil {
				resultCompletion.Value = value
			}
			return resultCompletion
		}

		// Keep the latest value from the body.
		if resultCompletion.Value != nil {
			value = resultCompletion.Value.(*JavaScriptValue)
		}

		perIterationEnvCompletion = CreatePerIterationEnvironment(runtime, perIterationLets)
		if perIterationEnvCompletion.Type != Normal {
			return perIterationEnvCompletion
		}

		// Evaluate the increment expression.
		if increment != nil {
			incrementCompletion := Evaluate(runtime, increment)
			if incrementCompletion.Type != Normal {
				return incrementCompletion
			}

			incrementValueCompletion := GetValue(incrementCompletion.Value.(*JavaScriptValue))
			if incrementValueCompletion.Type != Normal {
				return incrementValueCompletion
			}
		}
	}
}

func CreatePerIterationEnvironment(runtime *Runtime, letNames []string) *Completion {
	if len(letNames) == 0 {
		return NewUnusedCompletion()
	}

	runningContext := runtime.GetRunningExecutionContext()
	lastIterationEnv := runningContext.LexicalEnvironment
	thisIterationEnv := NewDeclarativeEnvironment(lastIterationEnv.GetOuterEnvironment())

	for _, name := range letNames {
		// Create a new binding in this iteration's environment.
		thisIterationEnv.CreateMutableBinding(name, false)

		// Get the value of the binding in the last iteration's environment.
		lastValueCompletion := lastIterationEnv.GetBindingValue(name, true)
		if lastValueCompletion.Type != Normal {
			return lastValueCompletion
		}

		// Initialize the binding in this iteration's environment with the value from the last iteration.
		thisIterationEnv.InitializeBinding(name, lastValueCompletion.Value.(*JavaScriptValue))
	}

	// Set the running context's lexical environment to the new environment.
	runningContext.LexicalEnvironment = thisIterationEnv
	return NewUnusedCompletion()
}
