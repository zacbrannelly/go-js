package runtime

import (
	"fmt"
	"slices"

	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateFunctionExpression(runtime *Runtime, functionExpression *ast.FunctionExpressionNode) *Completion {
	if functionExpression.Declaration {
		// This is EMPTY in the spec, unsure if this matters.
		return NewUnusedCompletion()
	}

	if functionExpression.Async && functionExpression.Generator {
		panic("TODO: Implement Async Generator Function Expression")
	}

	if functionExpression.Async && functionExpression.Arrow {
		panic("TODO: Implement Async Arrow Function Expression")
	}

	// ArrowFunctionExpression
	if functionExpression.Arrow {
		functionObject := InstantiateArrowFunctionExpression(runtime, functionExpression, nil)
		functionObjectValue := NewJavaScriptValue(TypeObject, functionObject)
		return NewNormalCompletion(functionObjectValue)
	}

	// FunctionExpression
	if !functionExpression.Async && !functionExpression.Generator {
		functionObject := InstantiateOrdinaryFunctionExpression(runtime, functionExpression, nil)
		functionObjectValue := NewJavaScriptValue(TypeObject, functionObject)
		return NewNormalCompletion(functionObjectValue)
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
			return EvaluateConciseBody(runtime, body, function, arguments)
		}

		// FunctionBody
		return EvaluateFunctionBody(runtime, body, function, arguments)
	}

	panic("TODO: Implement EvaluateBody")
}

func EvaluateConciseBody(
	runtime *Runtime,
	body ast.Node,
	function *FunctionObject,
	arguments []*JavaScriptValue,
) *Completion {
	completion := FunctionDeclarationInstantiation(runtime, function, arguments)
	if completion.Type != Normal {
		return completion
	}

	if body.GetNodeType() == ast.StatementList {
		// If () => { ... }
		completion = Evaluate(runtime, body)
		if completion.Type != Normal {
			return completion
		}
	} else {
		// If () => expression
		completion = Evaluate(runtime, body)
		if completion.Type != Normal {
			return completion
		}

		maybeRef := completion.Value.(*JavaScriptValue)
		returnValueCompletion := GetValue(maybeRef)
		if returnValueCompletion.Type != Normal {
			return returnValueCompletion
		}

		returnValue := returnValueCompletion.Value.(*JavaScriptValue)
		return NewReturnCompletion(returnValue)
	}

	return NewReturnCompletion(NewUndefinedValue())
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
		completion := SimpleIteratorBindingInitialization(runtime, formals, arguments, nil)
		if completion.Type != Normal {
			return completion
		}
	} else {
		completion := SimpleIteratorBindingInitialization(runtime, formals, arguments, env)
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
func SimpleIteratorBindingInitialization(runtime *Runtime, formals []ast.Node, argumentValues []*JavaScriptValue, env Environment) *Completion {
	var finalCompletion *Completion = NewUnusedCompletion()

	argIdx := 0
	for _, formal := range formals {
		if bindingElement, ok := formal.(*ast.BindingElementNode); ok {
			// BindingIdentifier
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

				// If a value is provided by the caller, use it.
				if argIdx < len(argumentValues) {
					value = argumentValues[argIdx]
					argIdx++
				}

				// If no value is provided, use the default value.
				if value == nil && bindingElement.GetInitializer() != nil {
					functionExpr, ok := bindingElement.GetInitializer().(*ast.FunctionExpressionNode)
					if ok && !functionExpr.Declaration && functionExpr.GetName() == nil {
						panic("TODO: Handle anonymous function definitions differently in SimpleIteratorBindingInitialization.")
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

			// BindingPattern
			if bindingElement.GetTarget().GetNodeType() == ast.ObjectBindingPattern || bindingElement.GetTarget().GetNodeType() == ast.ArrayBindingPattern {
				var value *JavaScriptValue = nil

				// If a value is provided by the caller, use it.
				if argIdx < len(argumentValues) {
					value = argumentValues[argIdx]
					argIdx++
				}

				// If no value is provided, use the default value.
				if value == nil && bindingElement.GetInitializer() != nil {
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

				return BindingInitialization(runtime, bindingElement.GetTarget(), value, env)
			}
		}

		if bindingRest, ok := formal.(*ast.BindingRestNode); ok {
			if bindingRest.GetBindingPattern() != nil {
				panic("TODO: Implement BindingRestNode with binding pattern for SimpleIteratorBindingInitialization.")
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
							panic("Assert failed: CreateDataProperty threw an unexpected error in SimpleIteratorBindingInitialization.")
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

func BindingInitialization(runtime *Runtime, node ast.Node, value *JavaScriptValue, env Environment) *Completion {
	isStrict := analyzer.IsStrictMode(node)

	// BindingIdentifier : Identifier
	if bindingIdentifier, ok := node.(*ast.BindingIdentifierNode); ok {
		return InitializeBoundName(runtime, bindingIdentifier.Identifier, value, env, isStrict)
	}

	// BindingPattern : ObjectBindingPattern
	if objectBindingPattern, ok := node.(*ast.ObjectBindingPatternNode); ok {
		completion := RequireObjectCoercible(value)
		if completion.Type != Normal {
			return completion
		}

		properties := make([]ast.Node, 0)

		for _, property := range objectBindingPattern.GetProperties() {
			if bindingProperty, ok := property.(*ast.BindingPropertyNode); ok {
				properties = append(properties, bindingProperty)
			}

			if bindingRest, ok := property.(*ast.BindingRestNode); ok {
				var excludedNames []*JavaScriptValue = nil
				if len(properties) > 0 {
					completion := PropertyBindingInitializationForPropertyList(runtime, properties, value, env)
					if completion.Type != Normal {
						return completion
					}

					excludedNames = completion.Value.([]*JavaScriptValue)
				}
				return RestBindingInitialization(runtime, bindingRest, value, env, excludedNames)
			}
		}

		completion = PropertyBindingInitializationForPropertyList(runtime, properties, value, env)
		if completion.Type != Normal {
			return completion
		}

		return NewUnusedCompletion()
	}

	// TODO: BindingPattern : ArrayBindingPattern
	if _, ok := node.(*ast.ArrayBindingPatternNode); ok {
		panic("TODO: Implement ArrayBindingPattern for BindingInitialization.")
	}

	panic("Assert failed: Unknown node type in BindingInitialization.")
}

func RestBindingInitialization(
	runtime *Runtime,
	bindingRest *ast.BindingRestNode,
	value *JavaScriptValue,
	env Environment,
	excludedNames []*JavaScriptValue,
) *Completion {
	// Get the binding of the identifier.
	isStrict := analyzer.IsStrictMode(bindingRest)
	lhsCompletion := ResolveBinding(bindingRest.GetIdentifier().(*ast.BindingIdentifierNode).Identifier, env, isStrict)
	if lhsCompletion.Type != Normal {
		return lhsCompletion
	}

	lhs := lhsCompletion.Value.(*JavaScriptValue)
	lhsRef := lhs.Value.(*Reference)

	// TODO: Set the prototype to %Object.prototype%.
	restObj := OrdinaryObjectCreate(nil)
	restObjVal := NewJavaScriptValue(TypeObject, restObj)

	completion := CopyDataProperties(restObj, value, excludedNames)
	if completion.Type != Normal {
		return completion
	}

	if env == nil {
		return PutValue(runtime, lhs, restObjVal)
	}
	return lhsRef.InitializeReferencedBinding(restObjVal)
}

func PropertyBindingInitializationForPropertyList(
	runtime *Runtime,
	properties []ast.Node,
	value *JavaScriptValue,
	env Environment,
) *Completion {
	names := make([]*JavaScriptValue, 0)

	for _, property := range properties {
		bindingProperty, ok := property.(*ast.BindingPropertyNode)
		if !ok {
			panic("Assert failed: Expected a BindingProperty node in PropertyBindingInitializationForPropertyList.")
		}

		if bindingIdentifier, ok := bindingProperty.GetTarget().(*ast.BindingIdentifierNode); ok {
			propertyKey := NewStringValue(bindingIdentifier.Identifier)
			completion := KeyedBindingInitialization(
				runtime,
				propertyKey,
				bindingProperty.GetTarget(),
				bindingProperty.GetInitializer(),
				value,
				env,
			)
			if completion.Type != Normal {
				return completion
			}

			names = append(names, propertyKey)
			continue
		}

		if bindingElement, ok := bindingProperty.GetBindingElement().(*ast.BindingElementNode); ok {
			var propertyKey *JavaScriptValue = nil

			if numberLiteral, ok := bindingProperty.GetTarget().(*ast.NumericLiteralNode); ok {
				// TODO: Get the NumericValue from the NumberLiteralNode.
				numberValCompletion := EvaluateNumericLiteral(runtime, numberLiteral)
				if numberValCompletion.Type != Normal {
					panic("Assert failed: EvaluateNumericLiteral threw an unexpected error in PropertyBindingInitializationForPropertyList.")
				}

				propertyKeyCompletion := ToString(numberValCompletion.Value.(*JavaScriptValue))
				if propertyKeyCompletion.Type != Normal {
					panic("Assert failed: ToString threw an unexpected error in PropertyBindingInitializationForPropertyList.")
				}
				propertyKey = propertyKeyCompletion.Value.(*JavaScriptValue)
			} else {
				propertyKeyEvalCompletion := Evaluate(runtime, bindingProperty.GetTarget())
				if propertyKeyEvalCompletion.Type != Normal {
					return propertyKeyEvalCompletion
				}
				propertyKey = propertyKeyEvalCompletion.Value.(*JavaScriptValue)
			}

			completion := KeyedBindingInitialization(
				runtime,
				propertyKey,
				bindingElement.GetTarget(),
				bindingElement.GetInitializer(),
				value,
				env,
			)
			if completion.Type != Normal {
				return completion
			}

			names = append(names, propertyKey)
			continue
		}

		panic("Assert failed: Unexpected binding property in PropertyBindingInitializationForPropertyList.")
	}

	return NewNormalCompletion(names)
}

func KeyedBindingInitialization(
	runtime *Runtime,
	propertyKey *JavaScriptValue,
	targetNode ast.Node,
	initializer ast.Node,
	value *JavaScriptValue,
	env Environment,
) *Completion {
	// SingleNameBinding : BindingIdentifier Initializer[opt]
	if bindingIdentifier, ok := targetNode.(*ast.BindingIdentifierNode); ok {
		bindingId := bindingIdentifier.Identifier
		isStrictMode := analyzer.IsStrictMode(bindingIdentifier)

		lhsCompletion := ResolveBinding(bindingId, env, isStrictMode)
		if lhsCompletion.Type != Normal {
			return lhsCompletion
		}

		lhs := lhsCompletion.Value.(*JavaScriptValue)

		objCompletion := ToObject(value)
		if objCompletion.Type != Normal {
			return objCompletion
		}

		obj := objCompletion.Value.(*JavaScriptValue).Value.(ObjectInterface)
		valCompletion := obj.Get(propertyKey, value)
		if valCompletion.Type != Normal {
			return valCompletion
		}

		val := valCompletion.Value.(*JavaScriptValue)

		if val.Type == TypeUndefined && initializer != nil {
			if functionExpr, ok := initializer.(*ast.FunctionExpressionNode); ok && functionExpr.Declaration && functionExpr.GetName() == nil {
				panic("TODO: Handle anonymous function definitions differently in KeyedBindingInitialization.")
			}

			defaultValueEval := Evaluate(runtime, initializer)
			if defaultValueEval.Type != Normal {
				return defaultValueEval
			}

			defaultValueCompletion := GetValue(defaultValueEval.Value.(*JavaScriptValue))
			if defaultValueCompletion.Type != Normal {
				return defaultValueCompletion
			}

			val = defaultValueCompletion.Value.(*JavaScriptValue)
		}

		if val == nil {
			val = NewUndefinedValue()
		}

		if env == nil {
			return PutValue(runtime, lhs, val)
		}
		return lhs.Value.(*Reference).InitializeReferencedBinding(val)
	}

	// BindingElement : BindingPattern Initializer[opt]
	objCompletion := ToObject(value)
	if objCompletion.Type != Normal {
		return objCompletion
	}

	obj := objCompletion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	valCompletion := obj.Get(propertyKey, value)

	if valCompletion.Type != Normal {
		return valCompletion
	}

	val := valCompletion.Value.(*JavaScriptValue)

	if val.Type == TypeUndefined && initializer != nil {
		// If the value is undefined, and there is a default value, use it.
		defaultValueEval := Evaluate(runtime, initializer)
		if defaultValueEval.Type != Normal {
			return defaultValueEval
		}

		defaultValueCompletion := GetValue(defaultValueEval.Value.(*JavaScriptValue))
		if defaultValueCompletion.Type != Normal {
			return defaultValueCompletion
		}

		val = defaultValueCompletion.Value.(*JavaScriptValue)
	}

	if val == nil {
		val = NewUndefinedValue()
	}

	return BindingInitialization(runtime, targetNode, val, env)
}

func RequireObjectCoercible(value *JavaScriptValue) *Completion {
	if value.Type == TypeUndefined || value.Type == TypeNull {
		return NewThrowCompletion(NewTypeError("Cannot convert undefined or null to an object"))
	}

	return NewNormalCompletion(value)
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
