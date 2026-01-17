package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateObjectLiteral(runtime *Runtime, objectLiteral *ast.ObjectLiteralNode) *Completion {
	object := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))

	if len(objectLiteral.GetProperties()) == 0 {
		return NewNormalCompletion(NewJavaScriptValue(TypeObject, object))
	}

	completion := PropertyDefinitionEvaluation(runtime, objectLiteral.GetProperties(), object)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, object))
}

func PropertyDefinitionEvaluation(
	runtime *Runtime,
	propertyNodes []ast.Node,
	object ObjectInterface,
) *Completion {
	for _, propertyNode := range propertyNodes {
		// PropertyDefinition : IdentifierReference
		if identifierReference, ok := propertyNode.(*ast.IdentifierReferenceNode); ok {
			propName := NewStringValue(identifierReference.Identifier)
			refCompletion := EvaluateIdentifierReference(runtime, identifierReference)
			if refCompletion.Type != Normal {
				return refCompletion
			}

			ref := refCompletion.Value.(*JavaScriptValue)
			propValueCompletion := GetValue(runtime, ref)
			if propValueCompletion.Type != Normal {
				return propValueCompletion
			}
			propValue := propValueCompletion.Value.(*JavaScriptValue)

			completion := CreateDataProperty(runtime, object, propName, propValue)
			if completion.Type != Normal {
				panic("Assert failed: PropertyDefinitionEvaluation threw an unexpected error.")
			}
			continue
		}

		// PropertyDefinition : PropertyName : AssignmentExpression
		if propertyDefinition, ok := propertyNode.(*ast.PropertyDefinitionNode); ok {
			var propKey *JavaScriptValue = nil

			if identifierName, ok := propertyDefinition.GetKey().(*ast.IdentifierNameNode); ok {
				propKey = NewStringValue(identifierName.Identifier)
			} else {
				propKeyEvalCompletion := Evaluate(runtime, propertyDefinition.GetKey())
				if propKeyEvalCompletion.Type != Normal {
					return propKeyEvalCompletion
				}
				propKey = propKeyEvalCompletion.Value.(*JavaScriptValue)

				// LiteralPropertyName : NumericLiteral
				if propKey.Type == TypeNumber {
					completion := ToString(runtime, propKey)
					if completion.Type != Normal {
						panic("Assert failed: ToString threw an unexpected error in PropertyDefinitionEvaluation.")
					}
					propKey = completion.Value.(*JavaScriptValue)
				}
			}

			isProtoSetter := false

			if stringVal, ok := propKey.Value.(*String); ok && stringVal.Value == "__proto__" && !IsComputedPropertyKey(propertyDefinition.GetKey()) {
				isProtoSetter = true
			}

			var propValue *JavaScriptValue = nil
			if functionExpression, ok := propertyDefinition.GetValue().(*ast.FunctionExpressionNode); ok && !functionExpression.Declaration && functionExpression.GetName() == nil && !isProtoSetter {
				if functionExpression.Declaration {
					panic("Assert failed: PropertyDefinitionEvaluation received a function declaration.")
				} else if functionExpression.Async {
					panic("TODO: Implement PropertyDefinitionEvaluation for async function expressions.")
				} else if functionExpression.Generator {
					panic("TODO: Implement PropertyDefinitionEvaluation for generator function expressions.")
				} else if functionExpression.Arrow {
					functionObj := InstantiateArrowFunctionExpression(runtime, functionExpression, propKey)
					propValue = NewJavaScriptValue(TypeObject, functionObj)
				} else {
					functionObj := InstantiateOrdinaryFunctionExpression(runtime, functionExpression, propKey)
					propValue = NewJavaScriptValue(TypeObject, functionObj)
				}
			} else {
				propValueEvalCompletion := Evaluate(runtime, propertyDefinition.GetValue())
				if propValueEvalCompletion.Type != Normal {
					return propValueEvalCompletion
				}

				propValueCompletion := GetValue(runtime, propValueEvalCompletion.Value.(*JavaScriptValue))
				if propValueCompletion.Type != Normal {
					return propValueCompletion
				}
				propValue = propValueCompletion.Value.(*JavaScriptValue)
			}

			if isProtoSetter {
				if propValue.Type == TypeObject || propValue.Type == TypeNull {
					completion := object.SetPrototypeOf(runtime, propValue)
					if completion.Type != Normal {
						panic("Assert failed: SetPrototypeOf threw an unexpected error in PropertyDefinitionEvaluation.")
					}
				}
				continue
			}

			completion := CreateDataProperty(runtime, object, propKey, propValue)
			if completion.Type != Normal {
				panic("Assert failed: CreateDataProperty threw an unexpected error in PropertyDefinitionEvaluation.")
			}
			continue
		}

		// PropertyDefinition : ... AssignmentExpression
		if spreadElement, ok := propertyNode.(*ast.SpreadElementNode); ok {
			refCompletion := Evaluate(runtime, spreadElement.GetExpression())
			if refCompletion.Type != Normal {
				return refCompletion
			}

			maybeRef := refCompletion.Value.(*JavaScriptValue)
			valueCompletion := GetValue(runtime, maybeRef)
			if valueCompletion.Type != Normal {
				return valueCompletion
			}

			value := valueCompletion.Value.(*JavaScriptValue)

			completion := CopyDataProperties(runtime, object, value, nil)
			if completion.Type != Normal {
				return completion
			}
			continue
		}

		// PropertyDefinition : MethodDefinition
		if _, ok := propertyNode.(*ast.MethodDefinitionNode); ok {
			completion := MethodDefinitionEvaluation(runtime, propertyNode.(*ast.MethodDefinitionNode), object, true)
			if completion.Type != Normal {
				return completion
			}
			continue
		}

		panic(fmt.Sprintf("Assert failed: PropertyDefinitionEvaluation received an unexpected property node: %s", ast.NodeTypeToString[propertyNode.GetNodeType()]))
	}
	return NewUnusedCompletion()
}

func MethodDefinitionEvaluation(
	runtime *Runtime,
	methodDefinition *ast.MethodDefinitionNode,
	object ObjectInterface,
	enumerable bool,
) *Completion {
	// AsyncMethod
	if methodDefinition.Async && !methodDefinition.Generator {
		panic("TODO: Implement MethodDefinitionEvaluation for async method definitions.")
	}

	// GeneratorMethod
	if methodDefinition.Generator && !methodDefinition.Async {
		panic("TODO: Implement MethodDefinitionEvaluation for generator method definitions.")
	}

	// AsyncGeneratorMethod
	if methodDefinition.Async && methodDefinition.Generator {
		panic("TODO: Implement MethodDefinitionEvaluation for async generator method definitions.")
	}

	if methodDefinition.Getter || methodDefinition.Setter {
		completion := Evaluate(runtime, methodDefinition.GetName())
		if completion.Type != Normal {
			return completion
		}

		propKey := completion.Value.(*JavaScriptValue)

		// TODO: Make this happen during Evaluation of the PropertyName.
		// LiteralPropertyName : NumericLiteral
		if propKey.Type == TypeNumber {
			completion := ToString(runtime, propKey)
			if completion.Type != Normal {
				panic("Assert failed: ToString threw an unexpected error in PropertyDefinitionEvaluation.")
			}
			propKey = completion.Value.(*JavaScriptValue)
		}

		env := runtime.GetRunningExecutionContext().LexicalEnvironment
		privateEnv := runtime.GetRunningExecutionContext().PrivateEnvironment

		closure := OrdinaryFunctionCreate(
			runtime,
			runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype),
			"TODO: Match source text from the method definition.",
			[]ast.Node{},
			methodDefinition.GetBody(),
			false,
			env,
			privateEnv,
		)

		MakeMethod(closure, object)
		if methodDefinition.Getter {
			SetFunctionNameWithPrefix(runtime, closure, propKey, "get")
		} else {
			SetFunctionNameWithPrefix(runtime, closure, propKey, "set")
		}

		if propKey.Type == TypePrivateName {
			panic("TODO: Implement MethodDefinitionEvaluation for private names.")
		}

		descriptor := &AccessorPropertyDescriptor{
			Enumerable:   enumerable,
			Configurable: true,
		}

		if methodDefinition.Setter {
			descriptor.Set = closure
		} else {
			descriptor.Get = closure
		}

		completion = DefinePropertyOrThrow(runtime, object, propKey, descriptor)
		if completion.Type != Normal {
			return completion
		}

		return NewUnusedCompletion()
	}

	completion := DefineMethod(runtime, methodDefinition, object, nil)
	if completion.Type != Normal {
		return completion
	}

	defineMethodResult := completion.Value.(*DefineMethodResult)

	SetFunctionName(runtime, defineMethodResult.Closure, defineMethodResult.Key)

	return DefineMethodProperty(
		runtime,
		object,
		defineMethodResult.Key,
		defineMethodResult.Closure,
		enumerable,
	)
}

func DefineMethodProperty(
	runtime *Runtime,
	homeObject ObjectInterface,
	key *JavaScriptValue,
	closure *FunctionObject,
	enumerable bool,
) *Completion {
	if key.Type == TypePrivateName {
		panic("TODO: Implement DefineMethodProperty for private names.")
	}

	descriptor := &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, closure),
		Writable:     true,
		Enumerable:   enumerable,
		Configurable: true,
	}
	completion := DefinePropertyOrThrow(runtime, homeObject, key, descriptor)
	if completion.Type != Normal {
		return completion
	}

	return NewUnusedCompletion()
}

type DefineMethodResult struct {
	Key     *JavaScriptValue
	Closure *FunctionObject
}

func DefineMethod(
	runtime *Runtime,
	methodDefinition *ast.MethodDefinitionNode,
	object ObjectInterface,
	functionPrototype ObjectInterface,
) *Completion {
	completion := Evaluate(runtime, methodDefinition.GetName())
	if completion.Type != Normal {
		return completion
	}

	name := completion.Value.(*JavaScriptValue)

	env := runtime.GetRunningExecutionContext().LexicalEnvironment
	privateEnv := runtime.GetRunningExecutionContext().PrivateEnvironment

	if functionPrototype == nil {
		functionPrototype = runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype)
	}

	closure := OrdinaryFunctionCreate(
		runtime,
		functionPrototype,
		"TODO: Extract source text from the method definition.",
		methodDefinition.GetParameters(),
		methodDefinition.GetBody(),
		false,
		env,
		privateEnv,
	)
	MakeMethod(closure, object)

	return NewNormalCompletion(&DefineMethodResult{
		Key:     name,
		Closure: closure,
	})
}

func MakeMethod(closure *FunctionObject, object ObjectInterface) {
	closure.HomeObject = object
}

func IsComputedPropertyKey(node ast.Node) bool {
	switch node.GetNodeType() {
	case ast.IdentifierName, ast.StringLiteral, ast.NumericLiteral:
		return false
	default:
		return true
	}
}
