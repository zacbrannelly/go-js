package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateNewExpression(runtime *Runtime, newExpression *ast.NewExpressionNode) *Completion {
	// newExpression.GetConstructor() always returns a CallExpressionNode.
	callExpression, ok := newExpression.GetConstructor().(*ast.CallExpressionNode)
	if !ok {
		panic("Assert failed: newExpression.GetConstructor() did not return a CallExpressionNode.")
	}

	completion := Evaluate(runtime, callExpression.GetCallee())
	if completion.Type != Normal {
		return completion
	}

	calleeRef := completion.Value.(*JavaScriptValue)

	completion = GetValue(runtime, calleeRef)
	if completion.Type != Normal {
		return completion
	}

	callee, ok := completion.Value.(*JavaScriptValue).Value.(*FunctionObject)
	if !ok {
		return NewThrowCompletion(NewTypeError("Not a constructor"))
	}

	completion = ArgumentListEvaluation(runtime, callExpression.GetArguments())
	if completion.Type != Normal {
		return completion
	}

	arguments := completion.Value.([]*JavaScriptValue)

	if !callee.HasConstruct {
		return NewThrowCompletion(NewTypeError("Not a constructor"))
	}

	return Construct(runtime, callee, arguments, nil)
}

func Construct(
	runtime *Runtime,
	constructor *FunctionObject,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if newTarget == nil {
		newTarget = NewJavaScriptValue(TypeObject, constructor)
	}

	return constructor.Construct(runtime, arguments, newTarget)
}
