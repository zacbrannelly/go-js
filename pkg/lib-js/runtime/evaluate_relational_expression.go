package runtime

import (
	"strings"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateRelationalExpression(runtime *Runtime, relationalExpression *ast.RelationalExpressionNode) *Completion {
	lRefCompletion := Evaluate(runtime, relationalExpression.GetLeft())
	if lRefCompletion.Type != Normal {
		return lRefCompletion
	}

	lRef := lRefCompletion.Value.(*JavaScriptValue)

	lValCompletion := GetValue(runtime, lRef)
	if lValCompletion.Type != Normal {
		return lValCompletion
	}

	lVal := lValCompletion.Value.(*JavaScriptValue)

	rRefCompletion := Evaluate(runtime, relationalExpression.GetRight())
	if rRefCompletion.Type != Normal {
		return rRefCompletion
	}

	rRef := rRefCompletion.Value.(*JavaScriptValue)

	rValCompletion := GetValue(runtime, rRef)
	if rValCompletion.Type != Normal {
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
	if resultCompletion.Type != Normal {
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

func EvaluateGreaterThanOrEqual(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	resultCompletion := IsLessThan(runtime, lVal, rVal, true)
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

func EvaluateInExpression(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	if rVal.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot use 'in' operator with a non-object type."))
	}

	rValObj := rVal.Value.(ObjectInterface)
	propertyKeyCompletion := ToPropertyKey(runtime, lVal)
	if propertyKeyCompletion.Type != Normal {
		return propertyKeyCompletion
	}
	propertyKey := propertyKeyCompletion.Value.(*JavaScriptValue)

	if propertyKey.Type == TypeString && strings.HasPrefix(propertyKey.Value.(*String).Value, "#") {
		panic("TODO: Support private properties.")
	}

	return rValObj.HasProperty(runtime, propertyKey)
}

func EvaluateInstanceOfExpression(runtime *Runtime, lVal *JavaScriptValue, rVal *JavaScriptValue) *Completion {
	if rVal.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Right-hand side of 'instanceof' is not an object."))
	}

	completion := GetMethod(runtime, rVal, runtime.SymbolHasInstance)
	if completion.Type != Normal {
		return completion
	}

	instOfHandler := completion.Value.(*JavaScriptValue)
	if handleFunc, ok := instOfHandler.Value.(FunctionInterface); ok {
		completion = handleFunc.Call(runtime, rVal, []*JavaScriptValue{lVal})
		if completion.Type != Normal {
			return completion
		}

		return ToBoolean(completion.Value.(*JavaScriptValue))
	}

	if _, ok := rVal.Value.(FunctionInterface); !ok {
		return NewThrowCompletion(NewTypeError(runtime, "Right-hand side of 'instanceof' is not callable."))
	}

	return OrdinaryHasInstance(runtime, rVal, lVal)
}
