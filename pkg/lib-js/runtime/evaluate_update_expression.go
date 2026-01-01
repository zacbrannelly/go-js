package runtime

import (
	"fmt"

	"zbrannelly.dev/go-js/pkg/lib-js/lexer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateUpdateExpression(runtime *Runtime, updateExpression *ast.UpdateExpressionNode) *Completion {
	lhsCompletion := Evaluate(runtime, updateExpression.GetValue())
	if lhsCompletion.Type != Normal {
		return lhsCompletion
	}

	lhsRef := lhsCompletion.Value.(*JavaScriptValue)

	lhsValueCompletion := GetValue(runtime, lhsRef)
	if lhsValueCompletion.Type != Normal {
		return lhsValueCompletion
	}

	lhsNumericCompletion := ToNumeric(runtime, lhsValueCompletion.Value.(*JavaScriptValue))
	if lhsNumericCompletion.Type != Normal {
		return lhsNumericCompletion
	}

	lhsNumericVal := lhsNumericCompletion.Value.(*JavaScriptValue)

	var newValue *JavaScriptValue
	if lhsNumericVal.Type == TypeNumber {
		switch updateExpression.Operator.Type {
		case lexer.Increment:
			newValue = NewJavaScriptValue(TypeNumber, NumberAdd(lhsNumericVal.Value.(*Number), &Number{Value: 1, NaN: false}))
		case lexer.Decrement:
			newValue = NewJavaScriptValue(TypeNumber, NumberSub(lhsNumericVal.Value.(*Number), &Number{Value: 1, NaN: false}))
		default:
			panic(fmt.Sprintf("Unexpected update operator: %s", updateExpression.Operator.Value))
		}
	} else {
		panic("TODO: BigInt update expressions are not implemented.")
	}

	completion := PutValue(runtime, lhsRef, newValue)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(newValue)
}
