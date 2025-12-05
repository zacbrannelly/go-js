package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func EvaluateCallExpression(runtime *Runtime, callExpression *ast.CallExpressionNode) *Completion {
	refCompletion := Evaluate(runtime, callExpression.GetCallee())
	if refCompletion.Type != Normal {
		return refCompletion
	}

	ref := refCompletion.Value.(*JavaScriptValue)

	funcValCompletion := GetValue(ref)
	if funcValCompletion.Type != Normal {
		return funcValCompletion
	}

	funcVal := funcValCompletion.Value.(*JavaScriptValue)

	if refRecord, ok := ref.Value.(*Reference); ok {
		isPropertyReference := refRecord.BaseObject != nil

		if !isPropertyReference && refRecord.ReferenceName.Type == TypeString && refRecord.ReferenceName.Value.(*String).Value == "eval" {
			panic("TODO: Support calling eval global function.")
		}
	}

	tailPosition := IsInTailPosition(callExpression)

	return EvaluateCall(runtime, funcVal, ref, callExpression.GetArguments(), tailPosition)
}

func EvaluateCall(
	runtime *Runtime,
	function *JavaScriptValue,
	ref *JavaScriptValue,
	arguments []ast.Node,
	tailPosition bool,
) *Completion {
	var thisValue *JavaScriptValue

	if refRecord, ok := ref.Value.(*Reference); ok {
		isPropertyReference := refRecord.BaseObject != nil

		if isPropertyReference {
			thisValue = refRecord.GetThisValue()
		} else {
			if refRecord.BaseEnv == nil {
				panic("Assert failed: EvaluateCall called on an unresolvable reference.")
			}

			thisValue = refRecord.BaseEnv.WithBaseObject()
		}
	} else {
		thisValue = NewUndefinedValue()
	}

	argListCompletion := ArgumentListEvaluation(runtime, arguments)
	if argListCompletion.Type != Normal {
		return argListCompletion
	}

	argList := argListCompletion.Value.([]*JavaScriptValue)

	if function.Type != TypeObject {
		return NewThrowCompletion(NewTypeError("Not a function"))
	}

	functionObject, isFunctionObject := function.Value.(*FunctionObject)

	if !isFunctionObject {
		return NewThrowCompletion(NewTypeError("Not a function"))
	}

	if tailPosition {
		PrepareForTailCall()
	}

	return functionObject.Call(runtime, thisValue, argList)
}

func PrepareForTailCall() {
	// TODO: Discard the running execution context's associated resources?
}

func ArgumentListEvaluation(runtime *Runtime, arguments []ast.Node) *Completion {
	result := make([]*JavaScriptValue, 0)
	for _, argument := range arguments {
		if argument.GetNodeType() == ast.SpreadElement {
			panic("TODO: Support spread arguments.")
		}

		refCompletion := Evaluate(runtime, argument)
		if refCompletion.Type != Normal {
			return refCompletion
		}

		ref := refCompletion.Value.(*JavaScriptValue)
		valCompletion := GetValue(ref)
		if valCompletion.Type != Normal {
			return valCompletion
		}

		result = append(result, valCompletion.Value.(*JavaScriptValue))
	}

	return NewNormalCompletion(result)
}

func IsInTailPosition(call ast.Node) bool {
	if !analyzer.IsStrictMode(call) {
		return false
	}

	panic("TODO: Implement IsInTailPosition")
}
