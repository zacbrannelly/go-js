package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/parser"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

func NewFunctionConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		FunctionConstructor,
		1,
		NewStringValue("Function"),
		realm,
		realm.GetIntrinsic(IntrinsicFunctionPrototype),
	)
	MakeConstructor(runtime, constructor)

	// Function.prototype
	constructor.DefineOwnProperty(runtime, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicFunctionPrototype)),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})

	// TODO: Define other properties.

	return constructor
}

func FunctionConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		arguments = append(arguments, NewStringValue(""))
	}

	activeFunction := runtime.GetRunningExecutionContext().Function

	parameterArgs := arguments[:len(arguments)-1]
	bodyArg := arguments[len(arguments)-1]

	return CreateDynamicFunction(
		runtime,
		activeFunction,
		newTarget,
		DynamicFunctionKindNormal,
		parameterArgs,
		bodyArg,
	)
}

type DynamicFunctionKind int

const (
	DynamicFunctionKindNormal DynamicFunctionKind = iota
	DynamicFunctionKindGenerator
	DynamicFunctionKindAsync
	DynamicFunctionKindAsyncGenerator
)

func CreateDynamicFunction(
	runtime *Runtime,
	constructor *FunctionObject,
	newTarget *JavaScriptValue,
	kind DynamicFunctionKind,
	parameterArgs []*JavaScriptValue,
	bodyArg *JavaScriptValue,
) *Completion {
	if newTarget == nil || newTarget.Type == TypeUndefined {
		newTarget = NewJavaScriptValue(TypeObject, constructor)
	}

	prefix := "function"
	fallbackProto := IntrinsicFunctionPrototype

	if kind != DynamicFunctionKindNormal {
		panic("TODO: Implement other dynamic function kinds.")
	}

	parameterStr := ""

	for idx, param := range parameterArgs {
		completion := ToString(runtime, param)
		if completion.Type != Normal {
			return completion
		}

		paramStr := completion.Value.(*JavaScriptValue).Value.(*String).Value
		if idx == 0 {
			parameterStr += paramStr
		} else {
			parameterStr += "," + paramStr
		}
	}

	completion := ToString(runtime, bodyArg)
	if completion.Type != Normal {
		return completion
	}

	bodyStr := completion.Value.(*JavaScriptValue).Value.(*String).Value
	bodyStr = "\n" + bodyStr + "\n"

	sourceStr := prefix + " anonymous(" + parameterStr + "\n){" + bodyStr + "}"

	formalParameters, err := parser.ParseFormalParameters(parameterStr, false, false)
	if err != nil {
		return NewThrowCompletion(NewSyntaxError(runtime, err.Error()))
	}

	functionBody, err := parser.ParseFunctionBody(bodyStr)
	if err != nil {
		return NewThrowCompletion(NewSyntaxError(runtime, err.Error()))
	}

	_, err = parser.ParseFunctionExpression(sourceStr, false)
	if err != nil {
		return NewThrowCompletion(NewSyntaxError(runtime, err.Error()))
	}

	completion = GetPrototypeFromConstructor(runtime, constructor, fallbackProto)
	if completion.Type != Normal {
		return completion
	}

	proto := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	// The evaluation code expects the function body to have a parent.
	parent := ast.NewFunctionExpressionNode(nil, formalParameters, functionBody)
	functionBody.SetParent(parent)

	functionObj := OrdinaryFunctionCreate(
		runtime,
		proto,
		sourceStr,
		formalParameters,
		functionBody,
		false,
		runtime.GetRunningExecutionContext().Realm.GlobalEnv,
		nil,
	)

	SetFunctionName(runtime, functionObj, NewStringValue("anonymous"))

	if kind == DynamicFunctionKindNormal {
		MakeConstructor(runtime, functionObj)
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, functionObj))
}
