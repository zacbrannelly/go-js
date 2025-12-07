package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateObjectLiteral(runtime *Runtime, objectLiteral *ast.ObjectLiteralNode) *Completion {
	object := OrdinaryObjectCreate(runtime.GetRunningRealm().Intrinsics[IntrinsicObjectPrototype])

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
			propValueCompletion := GetValue(ref)
			if propValueCompletion.Type != Normal {
				return propValueCompletion
			}
			propValue := propValueCompletion.Value.(*JavaScriptValue)

			completion := CreateDataProperty(object, propName, propValue)
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

				propValueCompletion := GetValue(propValueEvalCompletion.Value.(*JavaScriptValue))
				if propValueCompletion.Type != Normal {
					return propValueCompletion
				}
				propValue = propValueCompletion.Value.(*JavaScriptValue)
			}

			if isProtoSetter {
				if propValue.Type == TypeObject || propValue.Type == TypeNull {
					completion := object.SetPrototypeOf(propValue)
					if completion.Type != Normal {
						panic("Assert failed: SetPrototypeOf threw an unexpected error in PropertyDefinitionEvaluation.")
					}
				}
				continue
			}

			completion := CreateDataProperty(object, propKey, propValue)
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
			valueCompletion := GetValue(maybeRef)
			if valueCompletion.Type != Normal {
				return valueCompletion
			}

			value := valueCompletion.Value.(*JavaScriptValue)

			completion := CopyDataProperties(object, value, nil)
			if completion.Type != Normal {
				return completion
			}
			continue
		}

		// PropertyDefinition : MethodDefinition
		if _, ok := propertyNode.(*ast.MethodDefinitionNode); ok {
			panic("TODO: Implement PropertyDefinitionEvaluation for method definitions.")
		}

		panic(fmt.Sprintf("Assert failed: PropertyDefinitionEvaluation received an unexpected property node: %s", ast.NodeTypeToString[propertyNode.GetNodeType()]))
	}
	return NewUnusedCompletion()
}

func IsComputedPropertyKey(node ast.Node) bool {
	switch node.GetNodeType() {
	case ast.IdentifierName, ast.StringLiteral, ast.NumericLiteral:
		return false
	default:
		return true
	}
}
