package runtime

import (
	"fmt"
	"strings"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateUnaryExpression(runtime *Runtime, unaryExpression *ast.UnaryExpressionNode) *Completion {
	if unaryExpression.Operator.Type == lexer.Delete {
		return EvaluateDeleteUnaryExpression(runtime, unaryExpression.GetValue())
	}

	if unaryExpression.Operator.Type == lexer.Void {
		return EvaluateVoidUnaryExpression(runtime, unaryExpression.GetValue())
	}

	if unaryExpression.Operator.Type == lexer.TypeOf {
		return EvaluateTypeOfUnaryExpression(runtime, unaryExpression.GetValue())
	}

	if unaryExpression.Operator.Type == lexer.Plus {
		return EvaluatePlusUnaryExpression(runtime, unaryExpression.GetValue())
	}

	if unaryExpression.Operator.Type == lexer.Minus {
		return EvaluateMinusUnaryExpression(runtime, unaryExpression.GetValue())
	}

	if unaryExpression.Operator.Type == lexer.BitwiseNot {
		return EvaluateBitwiseNotUnaryExpression(runtime, unaryExpression.GetValue())
	}

	if unaryExpression.Operator.Type == lexer.Not {
		return EvaluateLogicalNotUnaryExpression(runtime, unaryExpression.GetValue())
	}

	panic(fmt.Sprintf("Unexpected unary operator: %s", unaryExpression.Operator.Value))
}

func EvaluateDeleteUnaryExpression(runtime *Runtime, value ast.Node) *Completion {
	refCompletion := Evaluate(runtime, value)
	if refCompletion.Type != Normal {
		return refCompletion
	}

	maybeRefVal := refCompletion.Value.(*JavaScriptValue)

	// If not a reference, return true.
	if maybeRefVal.Type != TypeReference {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	refVal := maybeRefVal.Value.(*Reference)

	// If unresolvable, return true.
	if refVal.BaseEnv == nil && refVal.BaseObject == nil {
		if refVal.Strict {
			panic("Assert failed: Cannot evaluate unary expression with strict unresolvable reference")
		}

		return NewNormalCompletion(NewBooleanValue(true))
	}

	// Is property reference?
	if refVal.BaseObject != nil {
		// TODO: This is off spec, unsure if this matters though.
		refNameCompletion := ToPropertyKey(refVal.ReferenceName)
		if refNameCompletion.Type != Normal {
			return refNameCompletion
		}

		refName := refNameCompletion.Value.(*JavaScriptValue)
		refNameString := PropertyKeyToString(refName)
		if strings.HasPrefix(refNameString, "#") {
			panic(fmt.Sprintf("Assert failed: Cannot delete private object property '%s'", refNameString))
		}

		// IsSuperReference?
		if refVal.ThisValue != nil {
			return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Cannot delete property '%s' since it's a super property", refNameString)))
		}

		baseObjectCompletion := ToObject(refVal.BaseObject)
		if baseObjectCompletion.Type != Normal {
			return baseObjectCompletion
		}

		baseObject := baseObjectCompletion.Value.(*JavaScriptValue).Value.(ObjectInterface)

		deleteCompletion := baseObject.Delete(refName)
		if deleteCompletion.Type != Normal {
			return deleteCompletion
		}

		if !deleteCompletion.Value.(*Boolean).Value && refVal.Strict {
			return NewThrowCompletion(NewTypeError(fmt.Sprintf("Cannot delete property '%s' of object", refNameString)))
		}

		return deleteCompletion
	} else {
		if refVal.ReferenceName.Type == TypeSymbol {
			panic("Assert failed: Cannot delete symbol properties.")
		}

		return refVal.BaseEnv.DeleteBinding(refVal.ReferenceName.Value.(*String).Value)
	}
}

func EvaluateVoidUnaryExpression(runtime *Runtime, value ast.Node) *Completion {
	refCompletion := Evaluate(runtime, value)
	if refCompletion.Type != Normal {
		return refCompletion
	}

	refVal := refCompletion.Value.(*JavaScriptValue)
	completion := GetValue(runtime, refVal)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(NewUndefinedValue())
}

func EvaluateTypeOfUnaryExpression(runtime *Runtime, value ast.Node) *Completion {
	refCompletion := Evaluate(runtime, value)
	if refCompletion.Type != Normal {
		return refCompletion
	}

	refVal := refCompletion.Value.(*JavaScriptValue)

	// If unresolvable reference, return "undefined".
	if refVal.Type == TypeReference {
		ref := refVal.Value.(*Reference)
		if ref.BaseObject == nil && ref.BaseEnv == nil {
			return NewNormalCompletion(NewStringValue("undefined"))
		}
	}

	completion := GetValue(runtime, refVal)
	if completion.Type != Normal {
		return completion
	}

	val := completion.Value.(*JavaScriptValue)

	if val.Type == TypeUndefined {
		return NewNormalCompletion(NewStringValue("undefined"))
	}

	if val.Type == TypeNull {
		return NewNormalCompletion(NewStringValue("object"))
	}

	if val.Type == TypeBoolean {
		return NewNormalCompletion(NewStringValue("boolean"))
	}

	if val.Type == TypeNumber {
		return NewNormalCompletion(NewStringValue("number"))
	}

	if val.Type == TypeString {
		return NewNormalCompletion(NewStringValue("string"))
	}

	if val.Type == TypeSymbol {
		return NewNormalCompletion(NewStringValue("symbol"))
	}

	if val.Type == TypeBigInt {
		return NewNormalCompletion(NewStringValue("bigint"))
	}

	if val.Type != TypeObject {
		panic(fmt.Sprintf("Unexpected value type: %d", val.Type))
	}

	// TODO: Return "function" if the object has a [[Call]] internal method.

	return NewNormalCompletion(NewStringValue("object"))
}

func EvaluatePlusUnaryExpression(runtime *Runtime, value ast.Node) *Completion {
	refCompletion := Evaluate(runtime, value)
	if refCompletion.Type != Normal {
		return refCompletion
	}

	refVal := refCompletion.Value.(*JavaScriptValue)
	completion := GetValue(runtime, refVal)
	if completion.Type != Normal {
		return completion
	}

	val := completion.Value.(*JavaScriptValue)
	return ToNumber(val)
}

func EvaluateMinusUnaryExpression(runtime *Runtime, value ast.Node) *Completion {
	refCompletion := Evaluate(runtime, value)
	if refCompletion.Type != Normal {
		return refCompletion
	}

	refVal := refCompletion.Value.(*JavaScriptValue)
	completion := GetValue(runtime, refVal)
	if completion.Type != Normal {
		return completion
	}

	val := completion.Value.(*JavaScriptValue)

	oldValCompletion := ToNumeric(val)
	if oldValCompletion.Type != Normal {
		return oldValCompletion
	}

	oldVal := oldValCompletion.Value.(*JavaScriptValue)
	if oldVal.Type == TypeNumber {
		oldValNumber := oldVal.Value.(*Number)
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberUnaryMinus(oldValNumber)))
	} else {
		panic("TODO: BigInt minus unary expressions are not implemented.")
	}
}

func EvaluateBitwiseNotUnaryExpression(runtime *Runtime, value ast.Node) *Completion {
	refCompletion := Evaluate(runtime, value)
	if refCompletion.Type != Normal {
		return refCompletion
	}

	refVal := refCompletion.Value.(*JavaScriptValue)
	completion := GetValue(runtime, refVal)
	if completion.Type != Normal {
		return completion
	}

	val := completion.Value.(*JavaScriptValue)
	oldValCompletion := ToNumeric(val)
	if oldValCompletion.Type != Normal {
		return oldValCompletion
	}

	oldVal := oldValCompletion.Value.(*JavaScriptValue)
	if oldVal.Type == TypeNumber {
		oldValNumber := oldVal.Value.(*Number)
		return NewNormalCompletion(NewJavaScriptValue(TypeNumber, NumberBitwiseNot(oldValNumber)))
	} else {
		panic("TODO: BigInt bitwise not unary expressions are not implemented.")
	}
}

func EvaluateLogicalNotUnaryExpression(runtime *Runtime, value ast.Node) *Completion {
	refCompletion := Evaluate(runtime, value)
	if refCompletion.Type != Normal {
		return refCompletion
	}

	refVal := refCompletion.Value.(*JavaScriptValue)
	completion := GetValue(runtime, refVal)
	if completion.Type != Normal {
		return completion
	}

	oldVal := completion.Value.(*JavaScriptValue)
	oldValBooleanCompletion := ToBoolean(oldVal)
	if oldValBooleanCompletion.Type != Normal {
		return oldValBooleanCompletion
	}

	oldValBoolean := oldValBooleanCompletion.Value.(*JavaScriptValue)
	oldValBooleanValue := oldValBoolean.Value.(*Boolean).Value

	return NewNormalCompletion(NewBooleanValue(!oldValBooleanValue))
}
