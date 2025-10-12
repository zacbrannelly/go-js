package runtime

import (
	"strings"

	"zbrannelly.dev/go-js/cmd/lexer"
	"zbrannelly.dev/go-js/cmd/parser/ast"
)

func EvaluateRelationalExpression(runtime *Runtime, relationalExpression *ast.RelationalExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, relationalExpression.GetLeft())
	if lRefCompletion.Type == Throw {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(lRef)
	if lValCompletion.Type == Throw {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	rRefCompletion := Evaluate(runtime, relationalExpression.GetRight())
	if rRefCompletion.Type == Throw {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)

	rValCompletion := GetValue(rRef)
	if rValCompletion.Type == Throw {
		return rValCompletion
	}

	rVal := rValCompletion.Value.(*JavaScriptValue)

	switch relationalExpression.GetOperator().Type {
	case lexer.LessThan:
		return EvaluateLessThan(runtime, lVal, rVal, true)
	case lexer.GreaterThan:
		return EvaluateLessThan(runtime, rVal, lVal, false)
	case lexer.LessThanEqual:
		return EvaluateLessThanOrEqual(runtime, lVal, rVal)
	case lexer.GreaterThanEqual:
		return EvaluateGreaterThanOrEqual(runtime, lVal, rVal)
	case lexer.In:
		return EvaluateInExpression(runtime, lVal, rVal)
	case lexer.InstanceOf:
		return EvaluateInstanceOfExpression(runtime, lVal, rVal)
	}

	panic("Unexpected relational operator.")
}

func EvaluateLessThan(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue, leftFirst bool) *Completion {
	resultCompletion := IsLessThan(runtime, lVal, rVal, leftFirst)
	if resultCompletion.Type == Throw {
		return resultCompletion
	}

	resultVal := resultCompletion.Value.(*JavaScriptValue)
	if resultVal.Type == TypeUndefined {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	return resultCompletion
}

func EvaluateLessThanOrEqual(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	resultCompletion := IsLessThan(runtime, rVal, lVal, false)
	if resultCompletion.Type == Throw {
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

func EvaluateGreaterThanOrEqual(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	resultCompletion := IsLessThan(runtime, rVal, lVal, true)
	if resultCompletion.Type == Throw {
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

func EvaluateInExpression(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	if rVal.Type != TypeObject {
		return NewThrowCompletion(NewTypeError("Cannot use 'in' operator with a non-object type."))
	}

	rValObj := rVal.Value.(*Object)
	propertyKeyCompletion := ToPropertyKey(runtime, lVal)
	if propertyKeyCompletion.Type == Throw {
		return propertyKeyCompletion
	}
	propertyKey := propertyKeyCompletion.Value.(*JavaScriptValue)

	if propertyKey.Type == TypeString && strings.HasPrefix(propertyKey.Value.(*String).Value, "#") {
		panic("TODO: Support private properties.")
	}

	return rValObj.HasProperty(propertyKey)
}

func EvaluateInstanceOfExpression(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	panic("TODO: Implement EvaluateInstanceOfExpression.")
}
