package runtime

import (
	"fmt"
	"slices"

	"zbrannelly.dev/go-js/cmd/analyzer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateFunctionExpression(runtime *Runtime, functionExpression *ast.FunctionExpressionNode) *Completion {
	if functionExpression.Declaration {
		// This is EMPTY in the spec, unsure if this matters.
		return NewUnusedCompletion()
	}

	panic("TODO: Implement EvaluateFunctionExpression")
}

func EvaluateBody(
	runtime *Runtime,
	body ast.Node,
	function *FunctionObject,
	arguments []*JavaScriptValue,
) *Completion {
	if body.GetParent() != nil && body.GetParent().GetNodeType() == ast.FunctionExpression {
		functionExpression := body.GetParent().(*ast.FunctionExpressionNode)

		// AsyncGeneratorBody
		if functionExpression.Async && functionExpression.Generator {
			panic("TODO: Implement Async Generator Body")
		}

		// AsyncConciseBody
		if functionExpression.Async && functionExpression.Arrow {
			panic("TODO: Implement Async Arrow Body")
		}

		// AsyncFunctionBody
		if functionExpression.Async {
			panic("TODO: Implement Async Body")
		}

		// GeneratorBody
		if functionExpression.Generator {
			panic("TODO: Implement Generator Body")
		}

		// ConciseBody
		if functionExpression.Arrow {
			panic("TODO: Implement Arrow Body")
		}

		// FunctionBody
		return EvaluateFunctionBody(runtime, body, function, arguments)
	}

	panic("TODO: Implement EvaluateBody")
}

func EvaluateFunctionBody(
	runtime *Runtime,
	body ast.Node,
	function *FunctionObject,
	arguments []*JavaScriptValue,
) *Completion {
	completion := FunctionDeclarationInstantiation(runtime, function, arguments)
	if completion.Type != Normal {
		return completion
	}

	completion = Evaluate(runtime, body)
	if completion.Type != Normal {
		return completion
	}

	return NewReturnCompletion(NewUndefinedValue())
}

func FunctionDeclarationInstantiation(runtime *Runtime, function *FunctionObject, arguments []*JavaScriptValue) *Completion {
	calleeContext := runtime.GetRunningExecutionContext()
	code := function.ScriptCode
	strict := function.Strict
	formals := function.FormalParameters

	parameterNames := make([]string, 0)
	for _, formal := range formals {
		parameterNames = append(parameterNames, BoundNames(formal)...)
	}

	nameMap := make(map[string]int)
	hasDuplicates := false
	for _, name := range parameterNames {
		if _, ok := nameMap[name]; ok {
			nameMap[name]++
			hasDuplicates = true
		} else {
			nameMap[name] = 1
		}
	}

	// TODO: This is needed when building the arguments object.
	// TODO: simpleParameterList := IsSimpleParameterList(formals)
	hasParameterExpressions := ContainsExpression(formals)

	varNames := VarDeclaredNames(code)
	varDeclarations := VarScopedDeclarations(code)
	lexicalNames := LexicallyDeclaredNames(code)

	functionNames := make([]string, 0)
	functionsToInitialize := make([]ast.Node, 0)

	for idx := len(varDeclarations) - 1; idx >= 0; idx-- {
		declaration := varDeclarations[idx]

		if functionExpression, ok := declaration.(*ast.FunctionExpressionNode); ok {
			if !functionExpression.Declaration {
				panic("Assert failed: Unexpected non-declaration function in var declarations.")
			}

			fnNames := BoundNames(functionExpression)
			if fnNames == nil || len(fnNames) != 1 {
				panic("Assert failed: Unexpected number of bound names for function in var declarations.")
			}

			functionName := fnNames[0]
			if slices.Contains(functionNames, functionName) {
				continue
			}

			functionNames = append(functionNames, functionName)
			functionsToInitialize = append(functionsToInitialize, functionExpression)
		}
	}

	argumentsObjectNeeded := true

	if function.ThisMode == ThisModeLexical {
		argumentsObjectNeeded = false
	} else if slices.Contains(parameterNames, "arguments") {
		argumentsObjectNeeded = false
	} else if hasParameterExpressions {
		if slices.Contains(functionNames, "arguments") || slices.Contains(lexicalNames, "arguments") {
			argumentsObjectNeeded = false
		}
	}

	var env Environment

	if strict || !hasParameterExpressions {
		env = calleeContext.LexicalEnvironment
	} else {
		env = NewDeclarativeEnvironment(calleeContext.LexicalEnvironment)

		if calleeContext.LexicalEnvironment != calleeContext.VariableEnvironment {
			panic("Assert failed: LexicalEnvironment and VariableEnvironment are not the same.")
		}

		calleeContext.LexicalEnvironment = env
	}

	parameterBindings := make([]string, 0)
	parameterBindings = append(parameterBindings, parameterNames...)

	for _, paramName := range parameterNames {
		alreadyDeclared := env.HasBinding(paramName)

		if !alreadyDeclared {
			completion := env.CreateMutableBinding(paramName, false)
			if completion.Type != Normal {
				panic("Assert failed: CreateMutableBinding threw an unexpected error in FunctionDeclarationInstantiation.")
			}

			if hasDuplicates {
				completion = env.InitializeBinding(paramName, NewUndefinedValue())
				if completion.Type != Normal {
					panic("Assert failed: InitializeBinding threw an unexpected error in FunctionDeclarationInstantiation.")
				}
			}
		}
	}

	if argumentsObjectNeeded {
		// TODO: Create the arguments object.
		// TODO: Add "arguments" to parameterBindings.
	}

	// Initialize the parameters with either the value passed in, the default value or undefined.
	if hasDuplicates {
		completion := SimpleBindingInitialization(runtime, formals, arguments, nil)
		if completion.Type != Normal {
			return completion
		}
	} else {
		completion := SimpleBindingInitialization(runtime, formals, arguments, env)
		if completion.Type != Normal {
			return completion
		}
	}

	var varEnv Environment = nil
	instantiatedVarNames := make(map[string]bool)

	if !hasParameterExpressions {
		for _, varName := range varNames {
			_, alreadyDeclared := instantiatedVarNames[varName]
			if !slices.Contains(parameterBindings, varName) && !alreadyDeclared {
				instantiatedVarNames[varName] = true

				completion := env.CreateMutableBinding(varName, false)
				if completion.Type != Normal {
					panic("Assert failed: CreateMutableBinding threw an unexpected error in FunctionDeclarationInstantiation.")
				}

				completion = env.InitializeBinding(varName, NewUndefinedValue())
				if completion.Type != Normal {
					panic("Assert failed: InitializeBinding threw an unexpected error in FunctionDeclarationInstantiation.")
				}
			}
		}

		varEnv = env
	} else {
		varEnv = NewDeclarativeEnvironment(env)

		for _, varName := range varNames {
			_, alreadyInstantiated := instantiatedVarNames[varName]
			if alreadyInstantiated {
				continue
			}
			instantiatedVarNames[varName] = true

			completion := varEnv.CreateMutableBinding(varName, false)
			if completion.Type != Normal {
				panic("Assert failed: CreateMutableBinding threw an unexpected error in FunctionDeclarationInstantiation.")
			}

			var initialValue *JavaScriptValue = nil
			if !slices.Contains(parameterBindings, varName) || slices.Contains(functionNames, varName) {
				initialValue = NewUndefinedValue()
			} else {
				completion = env.GetBindingValue(varName, false)
				if completion.Type != Normal {
					panic("Assert failed: GetBindingValue threw an unexpected error in FunctionDeclarationInstantiation.")
				}

				initialValue = completion.Value.(*JavaScriptValue)
			}

			completion = varEnv.InitializeBinding(varName, initialValue)
			if completion.Type != Normal {
				panic("Assert failed: InitializeBinding threw an unexpected error in FunctionDeclarationInstantiation.")
			}
		}
	}

	lexEnv := varEnv
	if !strict {
		lexEnv = NewDeclarativeEnvironment(varEnv)
	}

	calleeContext.LexicalEnvironment = lexEnv

	lexDeclarations := LexicallyScopedDeclarations(code)

	for _, declaration := range lexDeclarations {
		boundNames := BoundNames(declaration)
		for _, name := range boundNames {
			if IsConstantDeclaration(declaration) {
				completion := lexEnv.CreateImmutableBinding(name, true)
				if completion.Type != Normal {
					panic("Assert failed: CreateImmutableBinding threw an unexpected error in FunctionDeclarationInstantiation.")
				}
			} else {
				completion := lexEnv.CreateMutableBinding(name, false)
				if completion.Type != Normal {
					panic("Assert failed: CreateMutableBinding threw an unexpected error in FunctionDeclarationInstantiation.")
				}
			}
		}
	}

	privateEnv := calleeContext.PrivateEnvironment

	for _, function := range functionsToInitialize {
		boundNames := BoundNames(function)
		if len(boundNames) != 1 {
			panic("Assert failed: Unexpected number of bound names for function to initialize in FunctionDeclarationInstantiation.")
		}

		functionExpression, ok := function.(*ast.FunctionExpressionNode)
		if !ok {
			panic("Assert failed: Expected a function expression node to initialize in FunctionDeclarationInstantiation.")
		}

		functionName := boundNames[0]
		functionObject := InstantiateFunctionObject(runtime, functionExpression, lexEnv, privateEnv)

		completion := varEnv.SetMutableBinding(functionName, NewJavaScriptValue(TypeObject, functionObject), false)
		if completion.Type != Normal {
			panic("Assert failed: SetMutableBinding threw an unexpected error in FunctionDeclarationInstantiation.")
		}
	}

	return NewUnusedCompletion()
}

// NOTE: This implements the semantics of IteratorBindingInitialization, without using iterators.
func SimpleBindingInitialization(runtime *Runtime, formals []ast.Node, argumentValues []*JavaScriptValue, env Environment) *Completion {
	var finalCompletion *Completion = NewUnusedCompletion()

	argIdx := 0
	for _, formal := range formals {
		if bindingElement, ok := formal.(*ast.BindingElementNode); ok {
			if bindingIdentifier, ok := bindingElement.GetTarget().(*ast.BindingIdentifierNode); ok {
				paramName := bindingIdentifier.Identifier
				isStrictMode := analyzer.IsStrictMode(bindingIdentifier)

				// Resolve the binding in the environment.
				var lhsCompletion *Completion
				if env == nil {
					lhsCompletion = ResolveBindingFromCurrentContext(paramName, runtime, isStrictMode)
				} else {
					lhsCompletion = ResolveBinding(paramName, env, isStrictMode)
				}

				if lhsCompletion.Type != Normal {
					return lhsCompletion
				}

				lhsRef := lhsCompletion.Value.(*JavaScriptValue)

				var value *JavaScriptValue = nil

				if argIdx < len(argumentValues) {
					value = argumentValues[argIdx]
					argIdx++
				}

				// If no value is provided, use the default value.
				if value == nil && bindingElement.GetInitializer() != nil {
					functionExpr, ok := bindingElement.GetInitializer().(*ast.FunctionExpressionNode)
					if ok && functionExpr.Declaration && functionExpr.GetName() == nil {
						panic("TODO: Handle anonymous function definitions differently in SimpleBindingInitialization.")
					}

					defaultValueEval := Evaluate(runtime, bindingElement.GetInitializer())
					if defaultValueEval.Type != Normal {
						return defaultValueEval
					}

					defaultValueCompletion := GetValue(defaultValueEval.Value.(*JavaScriptValue))
					if defaultValueCompletion.Type != Normal {
						return defaultValueCompletion
					}

					value = defaultValueCompletion.Value.(*JavaScriptValue)
				}

				if value == nil {
					value = NewUndefinedValue()
				}

				if env == nil {
					finalCompletion = PutValue(runtime, lhsRef, value)
					if finalCompletion.Type != Normal {
						return finalCompletion
					}
					continue
				}
				finalCompletion = lhsRef.Value.(*Reference).InitializeReferencedBinding(value)
				if finalCompletion.Type != Normal {
					return finalCompletion
				}
			}
		}

		if bindingRest, ok := formal.(*ast.BindingRestNode); ok {
			if bindingRest.GetBindingPattern() != nil {
				panic("TODO: Implement BindingRestNode with binding pattern for SimpleBindingInitialization.")
			}

			if bindingIdentifier, ok := bindingRest.GetIdentifier().(*ast.BindingIdentifierNode); ok {
				array := NewArrayObject(0)

				paramName := bindingIdentifier.Identifier
				isStrictMode := analyzer.IsStrictMode(bindingIdentifier)

				// Resolve the binding in the environment.
				var lhsCompletion *Completion
				if env == nil {
					lhsCompletion = ResolveBindingFromCurrentContext(paramName, runtime, isStrictMode)
				} else {
					lhsCompletion = ResolveBinding(paramName, env, isStrictMode)
				}

				if lhsCompletion.Type != Normal {
					return lhsCompletion
				}

				lhsRef := lhsCompletion.Value.(*JavaScriptValue)

				arrayIdx := 0
				for {
					if argIdx < len(argumentValues) {
						value := argumentValues[argIdx]
						success := CreateDataProperty(array, NewStringValue(fmt.Sprintf("%d", arrayIdx)), value)
						if success.Type != Normal {
							panic("Assert failed: CreateDataProperty threw an unexpected error in SimpleBindingInitialization.")
						}
						argIdx++
						arrayIdx++
					} else {
						break
					}
				}

				arrayObj := NewJavaScriptValue(TypeObject, array)

				if env == nil {
					finalCompletion = PutValue(runtime, lhsRef, arrayObj)
					if finalCompletion.Type != Normal {
						return finalCompletion
					}
					continue
				}
				finalCompletion = lhsRef.Value.(*Reference).InitializeReferencedBinding(arrayObj)
				if finalCompletion.Type != Normal {
					return finalCompletion
				}
			}
		}
	}

	return finalCompletion
}

func IsSimpleParameterList(formals []ast.Node) bool {
	if len(formals) == 0 {
		return true
	}

	for _, formal := range formals {
		if bindingElement, ok := formal.(*ast.BindingElementNode); ok {
			if bindingElement.GetInitializer() != nil {
				return false
			}

			if bindingElement.GetTarget().GetNodeType() != ast.BindingIdentifier {
				return false
			}
		}
	}

	return true
}

func ContainsExpression(formals []ast.Node) bool {
	if len(formals) == 0 {
		return false
	}

	for _, formal := range formals {
		if bindingElement, ok := formal.(*ast.BindingElementNode); ok {
			if bindingElement.GetInitializer() != nil {
				return true
			}

			if objectBindingPattern, ok := bindingElement.GetTarget().(*ast.ObjectBindingPatternNode); ok {
				for _, property := range objectBindingPattern.GetProperties() {
					if bindingProperty, ok := property.(*ast.BindingPropertyNode); ok {
						if bindingProperty.GetInitializer() != nil {
							return true
						}
					}
				}
			}
		}
	}

	return false
}
