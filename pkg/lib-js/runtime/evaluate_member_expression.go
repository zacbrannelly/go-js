package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateMemberExpression(runtime *Runtime, memberExpression *ast.MemberExpressionNode) *Completion {
	baseRefCompletion := Evaluate(runtime, memberExpression.GetObject())
	if baseRefCompletion.Type != Normal {
		return baseRefCompletion
	}

	baseRef := baseRefCompletion.Value.(*JavaScriptValue)

	baseValCompletion := GetValue(baseRef)
	if baseValCompletion.Type != Normal {
		return baseValCompletion
	}

	baseVal := baseValCompletion.Value.(*JavaScriptValue)

	if memberExpression.Super {
		panic("TODO: Support super member expressions.")
	}

	strict := analyzer.IsStrictMode(memberExpression)

	if memberExpression.PropertyIdentifier != "" {
		return EvaluatePropertyAccessorWithIdentifierKey(baseVal, memberExpression.PropertyIdentifier, strict)
	}

	return EvaluatePropertyAccessorWithComputedKey(runtime, baseVal, memberExpression.GetProperty(), strict)
}

func EvaluatePropertyAccessorWithIdentifierKey(baseVal *JavaScriptValue, identifier string, strict bool) *Completion {
	var baseObj ObjectInterface = nil
	if baseVal.Type != TypeUndefined {
		baseObj = baseVal.Value.(ObjectInterface)
	}
	return NewNormalCompletion(NewReferenceValueForObject(baseObj, identifier, strict, nil))
}

func EvaluatePropertyAccessorWithComputedKey(runtime *Runtime, baseVal *JavaScriptValue, expression ast.Node, strict bool) *Completion {
	propertyNameRefCompletion := Evaluate(runtime, expression)
	if propertyNameRefCompletion.Type != Normal {
		return propertyNameRefCompletion
	}

	propertyNameRef := propertyNameRefCompletion.Value.(*JavaScriptValue)

	propertyNameValCompletion := GetValue(propertyNameRef)
	if propertyNameValCompletion.Type != Normal {
		return propertyNameValCompletion
	}

	propertyNameVal := propertyNameValCompletion.Value.(*JavaScriptValue)

	baseObj := baseVal.Value.(ObjectInterface)
	return NewNormalCompletion(NewReferenceValueForObjectProperty(baseObj, propertyNameVal, strict, nil))
}
