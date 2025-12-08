package runtime

import (
	"zbrannelly.dev/go-js/pkg/lib-js/analyzer"
	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

type ConstructorKind int

const (
	ConstructorKindBase ConstructorKind = iota
	ConstructorKindDerived
)

type ThisMode int

const (
	ThisModeLexical ThisMode = iota
	ThisModeStrict
	ThisModeGlobal
)

type NativeFunctionBehaviour func(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion

type FunctionObject struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool

	Environment               Environment
	PrivateEnvironment        Environment
	FormalParameters          []ast.Node
	ScriptCode                ast.Node
	ConstructorKind           ConstructorKind
	Realm                     *Realm
	Script                    *Script
	ThisMode                  ThisMode
	Strict                    bool
	HomeObject                ObjectInterface
	SourceText                string
	ClassFieldInitializerName *JavaScriptValue
	IsClassConstructor        bool

	// TODO: Module (to cover the Module part of ScriptOrModule)
	// TODO: Fields (to support class fields)
	// TODO: PrivateMethods (to support class private methods)

	// Built-in function specific properties.
	IsNativeFunction       bool
	InitialName            *JavaScriptValue
	NativeFunctionCallback NativeFunctionBehaviour
}

func OrdinaryFunctionCreate(
	runtime *Runtime,
	proto ObjectInterface,
	sourceText string,
	parameters []ast.Node,
	body ast.Node,
	isLexicalThis bool,
	env Environment,
	privateEnv Environment,
) *FunctionObject {
	strict := analyzer.IsStrictMode(body)

	var thisMode ThisMode
	if isLexicalThis {
		thisMode = ThisModeLexical
	} else if strict {
		thisMode = ThisModeStrict
	} else {
		thisMode = ThisModeGlobal
	}

	functionObject := &FunctionObject{
		Prototype:                 proto,
		Properties:                make(map[string]PropertyDescriptor),
		SymbolProperties:          make(map[*Symbol]PropertyDescriptor),
		Extensible:                true,
		SourceText:                sourceText,
		FormalParameters:          parameters,
		ScriptCode:                body,
		Strict:                    strict,
		ThisMode:                  thisMode,
		IsClassConstructor:        false,
		Environment:               env,
		PrivateEnvironment:        privateEnv,
		Script:                    runtime.GetRunningScript(),
		Realm:                     runtime.GetRunningExecutionContext().Realm,
		HomeObject:                nil,
		ClassFieldInitializerName: nil,
		// TODO: Set Fields to empty array.
		// TODO: Set PrivateMethods to empty array.
	}

	length := ExpectedArgumentCount(functionObject.FormalParameters)
	SetFunctionLength(functionObject, length)

	return functionObject
}

// TODO: Add support for prefix.
func CreateBuiltinFunction(
	runtime *Runtime,
	behaviour NativeFunctionBehaviour,
	length int,
	name *JavaScriptValue,
	realm *Realm,
	prototype ObjectInterface,
) *FunctionObject {
	if realm == nil {
		realm = runtime.GetRunningExecutionContext().Realm
	}

	if prototype == nil {
		prototype = realm.Intrinsics[IntrinsicFunctionPrototype]
	}

	functionObject := &FunctionObject{
		Properties:             make(map[string]PropertyDescriptor),
		SymbolProperties:       make(map[*Symbol]PropertyDescriptor),
		Extensible:             true,
		IsNativeFunction:       true,
		NativeFunctionCallback: behaviour,
		InitialName:            NewNullValue(),
		Prototype:              prototype,
		Realm:                  realm,
	}

	SetFunctionName(functionObject, name)
	SetFunctionLength(functionObject, length)

	return functionObject
}

func ExpectedArgumentCount(parameters []ast.Node) int {
	if len(parameters) == 0 {
		return 0
	}

	count := 0
	for _, param := range parameters {
		if param.GetNodeType() == ast.BindingRestProperty {
			break
		}

		if bindingElement, ok := param.(*ast.BindingElementNode); ok && bindingElement.GetInitializer() != nil {
			break
		}

		count++
	}

	return count
}

func SetFunctionLength(function *FunctionObject, length int) {
	completion := DefinePropertyOrThrow(function, NewStringValue("length"), &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(length), false),
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: SetFunctionLength threw an error when it should not have.")
	}
}

func InstantiateFunctionObject(
	runtime *Runtime,
	function *ast.FunctionExpressionNode,
	env Environment,
	privateEnv Environment,
) *FunctionObject {
	if !function.Declaration {
		panic("Assert failed: InstantiateFunctionObject called on a non-declaration function expression.")
	}

	// AsyncGeneratorFunctionDeclaration
	if function.Async && function.Generator {
		panic("TODO: Call InstantiateAsyncGeneratorFunctionObject")
	}

	// AsyncFunctionDeclaration
	if function.Async {
		panic("TODO: Call InstantiateAsyncFunctionObject")
	}

	// GeneratorFunctionDeclaration
	if function.Generator {
		panic("TODO: Call InstantiateGeneratorFunctionObject")
	}

	// FunctionDeclaration
	return InstantiateOrdinaryFunctionObject(runtime, function, env, privateEnv)
}

func InstantiateOrdinaryFunctionObject(
	runtime *Runtime,
	function *ast.FunctionExpressionNode,
	env Environment,
	privateEnv Environment,
) *FunctionObject {
	var name string
	if nameNode, ok := function.GetName().(*ast.BindingIdentifierNode); ok {
		name = nameNode.Identifier
	} else {
		name = "default"
	}

	// TODO: Extract source text from the function expression node.
	sourceText := "TODO: Modify parser to track source text for function expressions."
	functionObject := OrdinaryFunctionCreate(
		runtime,
		runtime.GetRunningRealm().Intrinsics[IntrinsicFunctionPrototype],
		sourceText,
		function.GetParameters(),
		function.GetBody(),
		false,
		env,
		privateEnv,
	)

	SetFunctionName(functionObject, NewStringValue(name))
	MakeConstructor(functionObject)

	return functionObject
}

func InstantiateOrdinaryFunctionExpression(
	runtime *Runtime,
	function *ast.FunctionExpressionNode,
	name *JavaScriptValue,
) *FunctionObject {
	if nameNode, ok := function.GetName().(*ast.BindingIdentifierNode); ok {
		if name != nil {
			panic("Assert failed: InstantiateOrdinaryFunctionExpression received a name for a node with a BindingIdentifierNode.")
		}
		name = NewStringValue(nameNode.Identifier)
	} else if name == nil {
		name = NewStringValue("")
	}

	runningContext := runtime.GetRunningExecutionContext()
	env := runningContext.LexicalEnvironment
	privateEnv := runningContext.PrivateEnvironment

	// TODO: Extract source text from the function expression node.
	sourceText := "TODO: Modify parser to track source text for function expressions."
	functionObject := OrdinaryFunctionCreate(
		runtime,
		runtime.GetRunningRealm().Intrinsics[IntrinsicFunctionPrototype],
		sourceText,
		function.GetParameters(),
		function.GetBody(),
		false,
		env,
		privateEnv,
	)

	SetFunctionName(functionObject, name)
	MakeConstructor(functionObject)

	return functionObject
}

func InstantiateArrowFunctionExpression(
	runtime *Runtime,
	function *ast.FunctionExpressionNode,
	name *JavaScriptValue,
) *FunctionObject {
	runningContext := runtime.GetRunningExecutionContext()
	env := runningContext.LexicalEnvironment
	privateEnv := runningContext.PrivateEnvironment

	// TODO: Extract source text from the function expression node.
	sourceText := "TODO: Modify parser to track source text for function expressions."
	functionObject := OrdinaryFunctionCreate(
		runtime,
		runtime.GetRunningRealm().Intrinsics[IntrinsicFunctionPrototype],
		sourceText,
		function.GetParameters(),
		function.GetBody(),
		true, // Arrow function expressions use LEXICAL-THIS.
		env,
		privateEnv,
	)

	if name == nil {
		name = NewStringValue("")
	}

	SetFunctionName(functionObject, name)
	return functionObject
}

func SetFunctionName(function *FunctionObject, name *JavaScriptValue) {
	if !function.Extensible {
		panic("Assert failed: SetFunctionName called on a non-extensible function object.")
	}

	if function.Properties["name"] != nil {
		panic("Assert failed: SetFunctionName called on a function object with a 'name' property.")
	}

	switch name.Type {
	case TypeSymbol:
		panic("TODO: Support setting function name to a symbol.")
	case TypePrivateName:
		panic("TODO: Support setting function name to a private name.")
	}

	if function.IsNativeFunction {
		function.InitialName = name
	}

	// TODO: Support prefix.

	completion := DefinePropertyOrThrow(function, NewStringValue("name"), &DataPropertyDescriptor{
		Value:        name,
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: SetFunctionName threw an error when it should not have.")
	}
}

func MakeConstructor(function *FunctionObject) {
	function.ConstructorKind = ConstructorKindBase

	prototype := function.Realm.Intrinsics[IntrinsicObjectPrototype]
	completion := DefinePropertyOrThrow(prototype, NewStringValue("constructor"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, function),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: MakeConstructor threw an error when it should not have.")
	}

	completion = DefinePropertyOrThrow(function, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, prototype),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	if completion.Type != Normal {
		panic("Assert failed: MakeConstructor threw an error when it should not have.")
	}
}

func (o *FunctionObject) Call(
	runtime *Runtime,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
) *Completion {
	if o.IsNativeFunction {
		return BuiltinCallOrConstruct(runtime, o, thisArg, arguments, NewUndefinedValue())
	}

	calleeContext := PrepareForOrdinaryCall(runtime, o, NewUndefinedValue())

	if calleeContext != runtime.GetRunningExecutionContext() {
		panic("Assert failed: callee context is not the running context.")
	}

	if o.IsClassConstructor {
		// Error is created in the callee context.
		errorObj := NewTypeError("Cannot call a class constructor.")

		// Pop the callee context.
		runtime.PopExecutionContext()

		return NewThrowCompletion(errorObj)
	}

	OrdinaryCallBindThis(o, calleeContext, thisArg)

	resultCompletion := OrdinaryCallEvaluateBody(runtime, o, arguments)
	runtime.PopExecutionContext()

	if resultCompletion.Type == Return {
		return NewNormalCompletion(resultCompletion.Value)
	}

	if resultCompletion.Type != Throw {
		panic("Assert failed: function result completion is not a return or throw completion.")
	}

	return resultCompletion
}

func BuiltinCallOrConstruct(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if !function.IsNativeFunction {
		panic("Assert failed: BuiltinCallOrConstruct called on a non-native function.")
	}

	calleeContext := &ExecutionContext{
		Function: function,
		Realm:    function.Realm,
	}
	runtime.PushExecutionContext(calleeContext)

	// Call the native function.
	if function.NativeFunctionCallback == nil {
		panic("Assert failed: Native function callback is nil.")
	}
	result := function.NativeFunctionCallback(
		runtime,
		function,
		thisArg,
		arguments,
		newTarget,
	)

	runtime.PopExecutionContext()
	return result
}

func PrepareForOrdinaryCall(
	runtime *Runtime,
	function *FunctionObject,
	newTarget *JavaScriptValue,
) *ExecutionContext {
	localEnv := NewFunctionEnvironment(function, newTarget)
	calleeContext := &ExecutionContext{
		Function:            function,
		Realm:               function.Realm,
		Script:              function.Script,
		LexicalEnvironment:  localEnv,
		VariableEnvironment: localEnv,
		PrivateEnvironment:  function.PrivateEnvironment,
		Labels:              make([]string, 0),
		// TODO: Set Module of function to the Module of the execution context.
	}

	runtime.PushExecutionContext(calleeContext)
	return calleeContext
}

func OrdinaryCallBindThis(
	function *FunctionObject,
	calleeContext *ExecutionContext,
	thisArg *JavaScriptValue,
) {
	if function.ThisMode == ThisModeLexical {
		return
	}

	var thisValue *JavaScriptValue

	if function.ThisMode == ThisModeStrict {
		thisValue = thisArg
	} else {
		if thisArg.Type == TypeUndefined || thisArg.Type == TypeNull {
			thisValue = NewJavaScriptValue(TypeObject, function.Realm.GlobalEnv.GlobalThisValue)
		} else {
			toObjectCompletion := ToObject(thisArg)
			if toObjectCompletion.Type != Normal {
				panic("Assert failed: ToObject threw an error when it should not have.")
			}

			thisValue = toObjectCompletion.Value.(*JavaScriptValue)
		}
	}

	localEnv, ok := calleeContext.LexicalEnvironment.(*DeclarativeEnvironment)
	if !ok || !localEnv.IsFunctionEnvironment {
		panic("Assert failed: OrdinaryCallBindThis called on a non-function environment.")
	}

	bindThisCompletion := BindThisValue(localEnv, thisValue)
	if bindThisCompletion.Type != Normal {
		panic("Assert failed: BindThisValue threw an error when it should not have.")
	}
}

func BindThisValue(env *DeclarativeEnvironment, thisValue *JavaScriptValue) *Completion {
	if env.ThisBindingStatus == ThisBindingStatusLexical {
		panic("Assert failed: BindThisValue called on a lexical environment.")
	}

	if env.ThisBindingStatus == ThisBindingStatusInitialized {
		return NewThrowCompletion(NewReferenceError("Cannot change the value of 'this'"))
	}

	env.ThisValue = thisValue
	env.ThisBindingStatus = ThisBindingStatusInitialized

	return NewUnusedCompletion()
}

func OrdinaryCallEvaluateBody(
	runtime *Runtime,
	function *FunctionObject,
	arguments []*JavaScriptValue,
) *Completion {
	return EvaluateBody(runtime, function.ScriptCode, function, arguments)
}

func (o *FunctionObject) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *FunctionObject) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *FunctionObject) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *FunctionObject) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *FunctionObject) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *FunctionObject) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
}

func (o *FunctionObject) GetExtensible() bool {
	return o.Extensible
}

func (o *FunctionObject) SetExtensible(extensible bool) {
	o.Extensible = extensible
}

func (o *FunctionObject) GetPrototypeOf() *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *FunctionObject) SetPrototypeOf(prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(o, prototype)
}

func (o *FunctionObject) GetOwnProperty(key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(o, key)
}

func (o *FunctionObject) HasProperty(key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(o, key)
}

func (o *FunctionObject) DefineOwnProperty(key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	return OrdinaryDefineOwnProperty(o, key, descriptor)
}

func (o *FunctionObject) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *FunctionObject) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *FunctionObject) Delete(key *JavaScriptValue) *Completion {
	return OrdinaryDelete(o, key)
}
