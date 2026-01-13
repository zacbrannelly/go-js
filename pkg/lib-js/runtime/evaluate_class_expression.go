package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateClassExpression(runtime *Runtime, classExpression *ast.ClassExpressionNode) *Completion {
	if classExpression.Declaration {
		completion := BindingClassDeclarationEvaluation(runtime, classExpression)
		if completion.Type != Normal {
			return completion
		}

		return NewUnusedCompletion()
	}

	var className *JavaScriptValue = nil
	var classBinding *JavaScriptValue = nil

	if classExpression.GetName() != nil {
		className = NewStringValue(classExpression.GetName().(*ast.BindingIdentifierNode).Identifier)
	} else {
		className = NewStringValue("")
		classBinding = NewUndefinedValue()
	}

	completion := ClassDefinitionEvaluation(runtime, classExpression, classBinding, className)
	if completion.Type != Normal {
		return completion
	}

	// TODO: Set value.[[SourceText]] to the source text of the class expression.

	return completion
}

func BindingClassDeclarationEvaluation(runtime *Runtime, classDeclaration *ast.ClassExpressionNode) *Completion {
	var name *JavaScriptValue
	if classDeclaration.GetName() != nil {
		name = NewStringValue(classDeclaration.GetName().(*ast.BindingIdentifierNode).Identifier)
	} else {
		name = NewUndefinedValue()
	}

	completion := ClassDefinitionEvaluation(runtime, classDeclaration, name, name)
	if completion.Type != Normal {
		return completion
	}

	value := completion.Value.(*JavaScriptValue)

	// TODO: Set value.[[SourceText]] to the source text of the class declaration.

	if name.Type != TypeUndefined {
		env := runtime.GetRunningExecutionContext().LexicalEnvironment
		isStrict := analyzer.IsStrictMode(classDeclaration)
		completion = InitializeBoundName(
			runtime,
			name.Value.(*String).Value,
			value,
			env,
			isStrict,
		)
		if completion.Type != Normal {
			return completion
		}
	}

	return NewNormalCompletion(value)
}

func ClassDefinitionEvaluation(
	runtime *Runtime,
	classDeclaration *ast.ClassExpressionNode,
	classBinding *JavaScriptValue,
	className *JavaScriptValue,
) *Completion {
	env := runtime.GetRunningExecutionContext().LexicalEnvironment
	classEnv := NewDeclarativeEnvironment(env)

	if classBinding.Type != TypeUndefined {
		completion := classEnv.CreateImmutableBinding(runtime, classBinding.Value.(*String).Value, true)
		if completion.Type != Normal {
			panic("Assert failed: CreateImmutableBinding threw an unexpected error in ClassDefinitionEvaluation.")
		}
	}

	outerPrivateEnvironment := runtime.GetRunningExecutionContext().PrivateEnvironment
	classPrivateEnvironment := NewPrivateEnvironment(outerPrivateEnvironment)

	if len(classDeclaration.GetElements()) > 0 {
		privateIdentifiers := PrivateBoundIdentifiers(classDeclaration.GetElements())
		for _, privateIdentifier := range privateIdentifiers {
			classPrivateEnvironment.Names = append(classPrivateEnvironment.Names, PrivateName{
				Description: privateIdentifier,
			})
		}
	}

	var protoParent ObjectInterface
	var constructorParent ObjectInterface

	if classDeclaration.GetHeritage() == nil {
		protoParent = runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype)
		constructorParent = runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype)
	} else {
		runtime.GetRunningExecutionContext().LexicalEnvironment = classEnv
		runtime.GetRunningExecutionContext().PrivateEnvironment = classPrivateEnvironment

		completion := Evaluate(runtime, classDeclaration.GetHeritage())

		runtime.GetRunningExecutionContext().LexicalEnvironment = env

		if completion.Type != Normal {
			return completion
		}

		superclassRef := completion.Value.(*JavaScriptValue)

		completion = GetValue(runtime, superclassRef)
		if completion.Type != Normal {
			return completion
		}

		superclass := completion.Value.(*JavaScriptValue)

		if superclass.Type == TypeNull {
			protoParent = nil
			constructorParent = runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype)
		} else if maybeConstructor, ok := superclass.Value.(FunctionInterface); !ok || !maybeConstructor.HasConstructMethod() {
			return NewThrowCompletion(NewTypeError(runtime, "Superclass is not a constructor."))
		} else {
			superclassObj := superclass.Value.(ObjectInterface)
			completion = superclassObj.Get(runtime, NewStringValue("prototype"), superclass)
			if completion.Type != Normal {
				return completion
			}

			prototypeVal := completion.Value.(*JavaScriptValue)
			if prototypeVal.Type != TypeObject && prototypeVal.Type != TypeNull {
				return NewThrowCompletion(NewTypeError(runtime, "Superclass prototype is not an object."))
			}

			protoParent = prototypeVal.Value.(ObjectInterface)
			constructorParent = superclassObj
		}
	}

	proto := OrdinaryObjectCreate(protoParent)

	var constructor ast.Node = nil
	if len(classDeclaration.GetElements()) > 0 {
		for _, classElement := range classDeclaration.GetElements() {
			if methodDef, ok := classElement.(*ast.MethodDefinitionNode); ok {
				if identifierName, ok := methodDef.GetName().(*ast.IdentifierNameNode); ok {
					if identifierName.Identifier == "constructor" {
						constructor = methodDef
						break
					}
				}
			}
		}
	}

	runtime.GetRunningExecutionContext().LexicalEnvironment = classEnv
	runtime.GetRunningExecutionContext().PrivateEnvironment = classPrivateEnvironment

	var constructorObj *FunctionObject = nil

	if constructor == nil {
		defaultConstructor := func(
			runtime *Runtime,
			function *FunctionObject,
			thisArg *JavaScriptValue,
			arguments []*JavaScriptValue,
			newTarget *JavaScriptValue,
		) *Completion {
			if newTarget == nil || newTarget.Type == TypeUndefined {
				return NewThrowCompletion(NewTypeError(runtime, "Requires `new` operator to create instances."))
			}

			newTargetObj := newTarget.Value.(FunctionInterface)
			activeFunction := runtime.GetRunningExecutionContext().Function

			var result *JavaScriptValue = nil
			if activeFunction.ConstructorKind == ConstructorKindDerived {
				completion := activeFunction.GetPrototypeOf()
				if completion.Type != Normal {
					panic("Assert failed: GetPrototypeOf threw an unexpected error in ClassDefinitionEvaluation.")
				}

				prototype, ok := completion.Value.(*JavaScriptValue).Value.(FunctionInterface)
				if !ok || !prototype.HasConstructMethod() {
					return NewThrowCompletion(NewTypeError(runtime, "Superclass is not a constructor."))
				}

				completion = Construct(runtime, prototype, arguments, newTarget)
				if completion.Type != Normal {
					return completion
				}

				result = completion.Value.(*JavaScriptValue)
			} else {
				completion := OrdinaryCreateFromConstructor(runtime, newTargetObj, IntrinsicObjectPrototype)
				if completion.Type != Normal {
					return completion
				}

				result = completion.Value.(*JavaScriptValue)
			}

			completion := InitializeInstanceElements(runtime, result.Value.(ObjectInterface), activeFunction)
			if completion.Type != Normal {
				return completion
			}

			return NewNormalCompletion(result)
		}

		constructorObj = CreateBuiltinFunction(
			runtime,
			defaultConstructor,
			0,
			className,
			runtime.GetRunningRealm(),
			constructorParent,
		)
	} else {
		completion := DefineMethod(runtime, constructor.(*ast.MethodDefinitionNode), proto, constructorParent)
		if completion.Type != Normal {
			panic("Assert failed: DefineMethod threw an unexpected error in ClassDefinitionEvaluation.")
		}

		constructorObj = completion.Value.(*DefineMethodResult).Closure

		constructorObj.IsClassConstructor = true
		SetFunctionName(runtime, constructorObj, className)
	}

	constructorObjVal := NewJavaScriptValue(TypeObject, constructorObj)
	MakeConstructorWithPrototype(runtime, constructorObj, false, proto)

	if classDeclaration.GetHeritage() != nil {
		constructorObj.ConstructorKind = ConstructorKindDerived
	}

	completion := DefineMethodProperty(runtime, proto, NewStringValue("constructor"), constructorObj, false)
	if completion.Type != Normal {
		panic("Assert failed: DefineMethodProperty threw an unexpected error in ClassDefinitionEvaluation.")
	}

	instancePrivateMethods := make([]*PrivateElement, 0)
	staticPrivateMethods := make([]*PrivateElement, 0)
	instanceFields := make([]*ClassFieldDefinition, 0)
	staticElements := make([]any, 0)

	for _, classElement := range classDeclaration.GetElements() {
		isStatic := IsClassElementStatic(classElement)
		if !isStatic {
			completion = ClassElementEvaluation(runtime, classElement, proto)
		} else {
			completion = ClassElementEvaluation(runtime, classElement, constructorObj)
		}

		if completion.Type != Normal {
			runtime.GetRunningExecutionContext().LexicalEnvironment = env
			runtime.GetRunningExecutionContext().PrivateEnvironment = outerPrivateEnvironment
			return completion
		}

		if privateElement, ok := completion.Value.(*PrivateElement); ok {
			if privateElement.Kind != PrivateElementKindMethod && privateElement.Kind != PrivateElementKindAccessor {
				panic("Assert failed: PrivateElement is not a method or accessor in ClassDefinitionEvaluation.")
			}

			var container []*PrivateElement
			if isStatic {
				container = staticPrivateMethods
			} else {
				container = instancePrivateMethods
			}

			containsElement := false
			for idx, pe := range container {
				if pe.Key.Description == privateElement.Key.Description {
					if pe.Kind != privateElement.Kind {
						panic("Assert failed: PrivateElement kind mismatch in ClassDefinitionEvaluation.")
					}

					containsElement = true

					var combined *PrivateElement

					if privateElement.Get != nil {
						combined = &PrivateElement{
							Key:  privateElement.Key,
							Kind: PrivateElementKindAccessor,
							Get:  privateElement.Get,
							Set:  pe.Set,
						}
					} else {
						combined = &PrivateElement{
							Key:  privateElement.Key,
							Kind: PrivateElementKindAccessor,
							Get:  pe.Get,
							Set:  privateElement.Set,
						}
					}

					container[idx] = combined
					break
				}
			}

			if !containsElement {
				container = append(container, privateElement)
			}

			if isStatic {
				staticPrivateMethods = container
			} else {
				instancePrivateMethods = container
			}
		} else if classFieldDef, ok := completion.Value.(*ClassFieldDefinition); ok {
			if isStatic {
				staticElements = append(staticElements, classFieldDef)
			} else {
				instanceFields = append(instanceFields, classFieldDef)
			}
		} else if classStaticBlockDef, ok := completion.Value.(*ClassStaticBlockDefinition); ok {
			staticElements = append(staticElements, classStaticBlockDef)
		}
	}

	runtime.GetRunningExecutionContext().LexicalEnvironment = env

	if classBinding.Type != TypeUndefined {
		completion = classEnv.InitializeBinding(
			runtime,
			classBinding.Value.(*String).Value,
			constructorObjVal,
		)
		if completion.Type != Normal {
			panic("Assert failed: InitializeBinding threw an unexpected error in ClassDefinitionEvaluation.")
		}
	}

	constructorObj.PrivateMethods = instancePrivateMethods
	constructorObj.Fields = instanceFields

	for _, privateMethod := range instancePrivateMethods {
		completion = PrivateMethodOrAccessorAdd(runtime, constructorObj, privateMethod)
		if completion.Type != Normal {
			panic("Assert failed: PrivateMethodOrAccessorAdd threw an unexpected error in ClassDefinitionEvaluation.")
		}
	}

	for _, elementRecord := range staticElements {
		if classFieldDef, ok := elementRecord.(*ClassFieldDefinition); ok {
			completion = DefineField(runtime, constructorObj, classFieldDef)
		} else {
			classStaticBlockDef, ok := elementRecord.(*ClassStaticBlockDefinition)
			if !ok {
				panic("Assert failed: Unexpected element type in static elements.")
			}

			completion = Call(
				runtime,
				NewJavaScriptValue(TypeObject, classStaticBlockDef.BodyFunction),
				constructorObjVal,
				[]*JavaScriptValue{},
			)
		}

		if completion.Type != Normal {
			runtime.GetRunningExecutionContext().PrivateEnvironment = outerPrivateEnvironment
			return completion
		}
	}

	runtime.GetRunningExecutionContext().PrivateEnvironment = outerPrivateEnvironment
	return NewNormalCompletion(constructorObjVal)
}

func PrivateBoundIdentifiers(classElements []ast.Node) []string {
	names := make([]string, 0)

	for _, classElement := range classElements {
		// FieldDefinition
		if classElement.GetNodeType() == ast.PropertyDefinition {
			propertyDefinition := classElement.(*ast.PropertyDefinitionNode)
			if identifierName, ok := propertyDefinition.GetKey().(*ast.IdentifierNameNode); ok {
				if identifierName.Identifier[0] == '#' {
					names = append(names, identifierName.Identifier)
				}
			}
		}

		// MethodDefinition
		if classElement.GetNodeType() == ast.MethodDefinition {
			methodDefinition := classElement.(*ast.MethodDefinitionNode)
			if identifierName, ok := methodDefinition.GetName().(*ast.IdentifierNameNode); ok {
				if identifierName.Identifier[0] == '#' {
					names = append(names, identifierName.Identifier)
				}
			}
		}
	}

	return names
}

func ClassElementEvaluation(runtime *Runtime, classElement ast.Node, object ObjectInterface) *Completion {
	if classElement.GetNodeType() == ast.PropertyDefinition {
		classFieldDefinition := classElement.(*ast.PropertyDefinitionNode)
		return ClassFieldDefinitionEvaluation(runtime, classFieldDefinition, object)
	}

	if classElement.GetNodeType() == ast.MethodDefinition {
		methodDefinition := classElement.(*ast.MethodDefinitionNode)
		return MethodDefinitionEvaluation(runtime, methodDefinition, object, false)
	}

	if classElement.GetNodeType() == ast.ClassStaticBlock {
		classStaticBlock := classElement.(*ast.ClassStaticBlockNode)
		return ClassStaticBlockDefinitionEvaluation(runtime, classStaticBlock, object)
	}

	return NewUnusedCompletion()
}

func ClassFieldDefinitionEvaluation(
	runtime *Runtime,
	classFieldDefinition *ast.PropertyDefinitionNode,
	object ObjectInterface,
) *Completion {
	completion := Evaluate(runtime, classFieldDefinition.GetKey())
	if completion.Type != Normal {
		return completion
	}

	name := completion.Value.(*JavaScriptValue)

	var initializer *FunctionObject = nil
	if classFieldDefinition.GetValue() != nil {
		env := runtime.GetRunningExecutionContext().LexicalEnvironment
		privateEnv := runtime.GetRunningExecutionContext().PrivateEnvironment
		initializer = OrdinaryFunctionCreate(
			runtime,
			runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype),
			"",
			[]ast.Node{},
			classFieldDefinition.GetValue(),
			false,
			env,
			privateEnv,
		)
		MakeMethod(initializer, object)
		initializer.ClassFieldInitializerName = name
	}

	return NewNormalCompletion(&ClassFieldDefinition{
		Name:        name,
		Initializer: initializer,
	})
}

func ClassStaticBlockDefinitionEvaluation(
	runtime *Runtime,
	classStaticBlock *ast.ClassStaticBlockNode,
	homeObject ObjectInterface,
) *Completion {
	env := runtime.GetRunningExecutionContext().LexicalEnvironment
	privateEnv := runtime.GetRunningExecutionContext().PrivateEnvironment
	bodyFunction := OrdinaryFunctionCreate(
		runtime,
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype),
		"",
		[]ast.Node{},
		classStaticBlock.GetBody(),
		false,
		env,
		privateEnv,
	)
	MakeMethod(bodyFunction, homeObject)
	return NewNormalCompletion(&ClassStaticBlockDefinition{
		BodyFunction: bodyFunction,
	})
}

func IsClassElementStatic(classElement ast.Node) bool {
	if classElement.GetNodeType() == ast.PropertyDefinition {
		propertyDefinition := classElement.(*ast.PropertyDefinitionNode)
		return propertyDefinition.Static
	}

	if classElement.GetNodeType() == ast.MethodDefinition {
		methodDefinition := classElement.(*ast.MethodDefinitionNode)
		return methodDefinition.Static
	}

	return false
}
