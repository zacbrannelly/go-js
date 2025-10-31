package runtime

import (
	"strings"

	"zbrannelly.dev/go-js/cmd/lexer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateRelationalExpression(runtime *Runtime, relationalExpression *ast.RelationalExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, relationalExpression.GetLeft())
	if lRefCompletion.Type != Normal {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(lRef)
	if lValCompletion.Type != Normal {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	rRefCompletion := Evaluate(runtime, relationalExpression.GetRight())
	if rRefCompletion.Type != Normal {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)

	rValCompletion := GetValue(rRef)
	if rValCompletion.Type != Normal {
		return rValCompletion
	}

	rVal := rValCompletion.Value.(*JavaScriptValue)

	switch relationalExpression.GetOperator().Type {
	case lexer.LessThan:
		return EvaluateLessThan(lVal, rVal, true)
	case lexer.GreaterThan:
		return EvaluateLessThan(rVal, lVal, false)
	case lexer.LessThanEqual:
		return EvaluateLessThanOrEqual(lVal, rVal)
	case lexer.GreaterThanEqual:
		return EvaluateGreaterThanOrEqual(lVal, rVal)
	case lexer.In:
		return EvaluateInExpression(lVal, rVal)
	case lexer.InstanceOf:
		return EvaluateInstanceOfExpression(lVal, rVal)
	}

	panic("Unexpected relational operator.")
}

func EvaluateLessThan(lVal *JavaScriptValue, rVal *JavaScriptValue, leftFirst bool) *Completion {
	resultCompletion := IsLessThan(lVal, rVal, leftFirst)
	if resultCompletion.Type != Normal {
		return resultCompletion
	}

	resultVal := resultCompletion.Value.(*JavaScriptValue)
	if resultVal.Type == TypeUndefined {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	return resultCompletion
}

func EvaluateLessThanOrEqual(lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	resultCompletion := IsLessThan(rVal, lVal, false)
	if resultCompletion.Type != Normal {
		return resultCompletion
	}

	resultVal := resultCompletion.Value.(*JavaScriptValue)
	if resultVal.Type == TypeUndefined {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	if resultVal.Type == TypeBoolean && resultVal.Value.(*Boolean).Value {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func EvaluateGreaterThanOrEqual(lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	resultCompletion := IsLessThan(rVal, lVal, true)
	if resultCompletion.Type != Normal {
		return resultCompletion
	}

	resultVal := resultCompletion.Value.(*JavaScriptValue)
	if resultVal.Type == TypeUndefined {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	if resultVal.Type == TypeBoolean && resultVal.Value.(*Boolean).Value {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func EvaluateInExpression(lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	if rVal.Type != TypeObject {
		return NewThrowCompletion(NewTypeError("Cannot use 'in' operator with a non-object type."))
	}

	rValObj := rVal.Value.(*Object)
	propertyKeyCompletion := ToPropertyKey(lVal)
	if propertyKeyCompletion.Type != Normal {
		return propertyKeyCompletion
	}
	propertyKey := propertyKeyCompletion.Value.(*JavaScriptValue)

	if propertyKey.Type == TypeString && strings.HasPrefix(propertyKey.Value.(*String).Value, "#") {
		panic("TODO: Support private properties.")
	}

	return rValObj.HasProperty(propertyKey)
}

func EvaluateInstanceOfExpression(lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	panic("TODO: Implement EvaluateInstanceOfExpression.")
}
