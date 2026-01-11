package runtime

import (
	"fmt"

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

type FunctionInterface interface {
	Call(runtime *Runtime, thisArg *JavaScriptValue, arguments []*JavaScriptValue) *Completion
	Construct(runtime *Runtime, arguments []*JavaScriptValue, newTarget *JavaScriptValue) *Completion
	Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion
	HasConstructMethod() bool
}

type ClassFieldDefinition struct {
	Name        *JavaScriptValue
	Initializer *FunctionObject
}

type ClassStaticBlockDefinition struct {
	BodyFunction *FunctionObject
}

type PrivateElementKind int

const (
	PrivateElementKindField PrivateElementKind = iota
	PrivateElementKindMethod
	PrivateElementKindAccessor
)

type PrivateElement struct {
	Key   PrivateName
	Kind  PrivateElementKind
	Value *JavaScriptValue
	Get   FunctionInterface
	Set   FunctionInterface
}

type FunctionObject struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool
	PrivateElements  []*PrivateElement

	Environment               Environment
	PrivateEnvironment        *PrivateEnvironment
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
	PrivateMethods            []*PrivateElement
	Fields                    []*ClassFieldDefinition

	// TODO: Module (to cover the Module part of ScriptOrModule)

	// Built-in function specific properties.
	IsNativeFunction       bool
	InitialName            *JavaScriptValue
	NativeFunctionCallback NativeFunctionBehaviour

	HasConstruct bool
}

func OrdinaryFunctionCreate(
	runtime *Runtime,
	proto ObjectInterface,
	sourceText string,
	parameters []ast.Node,
	body ast.Node,
	isLexicalThis bool,
	env Environment,
	privateEnv *PrivateEnvironment,
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
		PrivateElements:           make([]*PrivateElement, 0),
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
	SetFunctionLength(runtime, functionObject, length)

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
		prototype = realm.GetIntrinsic(IntrinsicFunctionPrototype)
	}

	functionObject := &FunctionObject{
		Properties:             make(map[string]PropertyDescriptor),
		SymbolProperties:       make(map[*Symbol]PropertyDescriptor),
		Extensible:             true,
		PrivateElements:        make([]*PrivateElement, 0),
		IsNativeFunction:       true,
		NativeFunctionCallback: behaviour,
		InitialName:            NewNullValue(),
		Prototype:              prototype,
		Realm:                  realm,
	}

	SetFunctionName(runtime, functionObject, name)
	SetFunctionLength(runtime, functionObject, length)

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

func SetFunctionLength(runtime *Runtime, function ObjectInterface, length int) {
	completion := DefinePropertyOrThrow(runtime, function, NewStringValue("length"), &DataPropertyDescriptor{
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
	privateEnv *PrivateEnvironment,
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
	privateEnv *PrivateEnvironment,
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
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype),
		sourceText,
		function.GetParameters(),
		function.GetBody(),
		false,
		env,
		privateEnv,
	)

	SetFunctionName(runtime, functionObject, NewStringValue(name))
	MakeConstructor(runtime, functionObject)

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

	// Add binding for the function name if present in the expression as a binding identifier.
	if function.GetName() != nil {
		env = NewDeclarativeEnvironment(env)
		completion := env.CreateImmutableBinding(runtime, name.Value.(*String).Value, false)
		if completion.Type != Normal {
			panic("Assert failed: CreateImmutableBinding threw an unexpected error in InstantiateOrdinaryFunctionExpression.")
		}
	}

	// TODO: Extract source text from the function expression node.
	sourceText := "TODO: Modify parser to track source text for function expressions."
	functionObject := OrdinaryFunctionCreate(
		runtime,
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype),
		sourceText,
		function.GetParameters(),
		function.GetBody(),
		false,
		env,
		privateEnv,
	)

	SetFunctionName(runtime, functionObject, name)
	MakeConstructor(runtime, functionObject)

	// Initialize the name binding with the function object.
	if function.GetName() != nil {
		completion := env.InitializeBinding(runtime, name.Value.(*String).Value, NewJavaScriptValue(TypeObject, functionObject))
		if completion.Type != Normal {
			panic("Assert failed: InitializeBinding threw an unexpected error in InstantiateOrdinaryFunctionExpression.")
		}
	}

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
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicFunctionPrototype),
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

	SetFunctionName(runtime, functionObject, name)
	return functionObject
}

func SetFunctionName(runtime *Runtime, function *FunctionObject, name *JavaScriptValue) {
	if !function.Extensible {
		panic("Assert failed: SetFunctionName called on a non-extensible function object.")
	}

	if function.Properties["name"] != nil {
		panic("Assert failed: SetFunctionName called on a function object with a 'name' property.")
	}

	switch name.Type {
	case TypeSymbol:
		symbol := name.Value.(*Symbol)
		if symbol.Description == "" {
			name = NewStringValue("")
		} else {
			name = NewStringValue(fmt.Sprintf("[%s]", symbol.Description))
		}
	case TypePrivateName:
		panic("TODO: Support setting function name to a private name.")
	}

	if function.IsNativeFunction {
		function.InitialName = name
	}

	// TODO: Support prefix.

	completion := DefinePropertyOrThrow(runtime, function, NewStringValue("name"), &DataPropertyDescriptor{
		Value:        name,
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: SetFunctionName threw an error when it should not have.")
	}
}

func SetFunctionNameWithPrefix(runtime *Runtime, function ObjectInterface, name *JavaScriptValue, prefix string) {
	if !function.GetExtensible() {
		panic("Assert failed: SetFunctionName called on a non-extensible function object.")
	}

	if function.GetProperties()["name"] != nil {
		panic("Assert failed: SetFunctionName called on a function object with a 'name' property.")
	}

	switch name.Type {
	case TypeSymbol:
		symbol := name.Value.(*Symbol)
		if symbol.Description == "" {
			name = NewStringValue("")
		} else {
			name = NewStringValue(fmt.Sprintf("[%s]", symbol.Description))
		}
	case TypePrivateName:
		panic("TODO: Support setting function name to a private name.")
	}

	if functionObj, ok := function.(*FunctionObject); ok && functionObj.IsNativeFunction {
		functionObj.InitialName = name
	}

	name = NewStringValue(prefix + " " + name.Value.(*String).Value)

	completion := DefinePropertyOrThrow(runtime, function, NewStringValue("name"), &DataPropertyDescriptor{
		Value:        name,
		Writable:     false,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: SetFunctionName threw an error when it should not have.")
	}
}

func MakeConstructor(runtime *Runtime, function *FunctionObject) {
	function.HasConstruct = true
	function.ConstructorKind = ConstructorKindBase

	prototype := OrdinaryObjectCreate(function.Realm.GetIntrinsic(IntrinsicObjectPrototype))
	completion := DefinePropertyOrThrow(runtime, prototype, NewStringValue("constructor"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, function),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: MakeConstructor threw an error when it should not have.")
	}

	completion = DefinePropertyOrThrow(runtime, function, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, prototype),
		Writable:     true,
		Enumerable:   false,
		Configurable: false,
	})
	if completion.Type != Normal {
		panic("Assert failed: MakeConstructor threw an error when it should not have.")
	}
}

func MakeConstructorWithPrototype(runtime *Runtime, function *FunctionObject, writablePrototype bool, prototype ObjectInterface) {
	function.HasConstruct = true
	function.ConstructorKind = ConstructorKindBase

	completion := DefinePropertyOrThrow(runtime, function, NewStringValue("prototype"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, prototype),
		Writable:     writablePrototype,
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
		errorObj := NewTypeError(runtime, "Cannot call a class constructor.")

		// Pop the callee context.
		runtime.PopExecutionContext()

		return NewThrowCompletion(errorObj)
	}

	OrdinaryCallBindThis(runtime, o, calleeContext, thisArg)

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

func (o *FunctionObject) Construct(
	runtime *Runtime,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if !o.HasConstruct {
		panic("Assert failed: Construct called on a function object that does not have the [[Construct]] internal method.")
	}

	if o.IsNativeFunction {
		completion := BuiltinCallOrConstruct(runtime, o, nil, arguments, newTarget)
		if completion.Type != Normal {
			return completion
		}

		result := completion.Value.(*JavaScriptValue)
		if result.Type != TypeObject {
			panic("Assert failed: BuiltinCallOrConstruct returned a non-object result in Construct.")
		}

		return completion
	}

	var thisArgument *JavaScriptValue = nil
	if o.ConstructorKind == ConstructorKindBase {
		completion := OrdinaryCreateFromConstructor(runtime, o, IntrinsicObjectPrototype)
		if completion.Type != Normal {
			return completion
		}

		thisArgument = completion.Value.(*JavaScriptValue)
	}

	calleeContext := PrepareForOrdinaryCall(runtime, o, newTarget)

	if o.ConstructorKind == ConstructorKindBase {
		OrdinaryCallBindThis(runtime, o, calleeContext, thisArgument)

		// TODO: Call InitializeInstanceElements (to initialize the private methods and fields)
	}

	// Store the constructor env before evaluating the body, this is important as the env changes when evaluating the body.
	constructorEnv := calleeContext.LexicalEnvironment

	completion := OrdinaryCallEvaluateBody(runtime, o, arguments)
	runtime.PopExecutionContext()

	if completion.Type == Throw {
		return completion
	}

	if completion.Type != Return {
		panic("Assert failed: function result completion is not a return or throw completion.")
	}

	result := completion.Value.(*JavaScriptValue)
	if result.Type == TypeObject {
		return NewNormalCompletion(result)
	}

	if result.Type != TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid 'return' type in constructor."))
	}

	completion = constructorEnv.GetThisBinding(runtime)
	if completion.Type != Normal {
		return completion
	}

	thisValue := completion.Value.(*JavaScriptValue)
	if thisValue.Type != TypeObject {
		panic("Assert failed: This value is not an object in Construct.")
	}

	return NewNormalCompletion(thisValue)
}

func InitializeInstanceElements(runtime *Runtime, object ObjectInterface, constructor *FunctionObject) *Completion {
	for _, method := range constructor.PrivateMethods {
		completion := PrivateMethodOrAccessorAdd(runtime, object, method)
		if completion.Type != Normal {
			return completion
		}
	}

	for _, field := range constructor.Fields {
		completion := DefineField(runtime, object, field)
		if completion.Type != Normal {
			return completion
		}
	}

	return NewUnusedCompletion()
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
	runtime *Runtime,
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
			toObjectCompletion := ToObject(runtime, thisArg)
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

	bindThisCompletion := BindThisValue(runtime, localEnv, thisValue)
	if bindThisCompletion.Type != Normal {
		panic("Assert failed: BindThisValue threw an error when it should not have.")
	}
}

func BindThisValue(runtime *Runtime, env *DeclarativeEnvironment, thisValue *JavaScriptValue) *Completion {
	if env.ThisBindingStatus == ThisBindingStatusLexical {
		panic("Assert failed: BindThisValue called on a lexical environment.")
	}

	if env.ThisBindingStatus == ThisBindingStatusInitialized {
		return NewThrowCompletion(NewReferenceError(runtime, "Cannot change the value of 'this'"))
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

func (o *FunctionObject) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(runtime, o, key)
}

func (o *FunctionObject) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *FunctionObject) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
}

func (o *FunctionObject) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *FunctionObject) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *FunctionObject) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryDelete(runtime, o, key)
}

func (o *FunctionObject) OwnPropertyKeys() *Completion {
	return NewNormalCompletion(OrdinaryOwnPropertyKeys(o))
}

func (o *FunctionObject) PreventExtensions() *Completion {
	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *FunctionObject) GetPrivateElements() []*PrivateElement {
	return o.PrivateElements
}

func (o *FunctionObject) SetPrivateElements(privateElements []*PrivateElement) {
	o.PrivateElements = privateElements
}

func (o *FunctionObject) HasConstructMethod() bool {
	return o.HasConstruct
}

func GetFunctionRealm(runtime *Runtime, function FunctionInterface) *Completion {
	if functionObj, ok := function.(*FunctionObject); ok {
		// TODO: The spec says to check if it has the [[Realm]] internal slot, not to check if it is nil.
		if functionObj.Realm != nil {
			return NewNormalCompletion(functionObj.Realm)
		}
	}

	if boundFunc, ok := function.(*BoundFunction); ok {
		boundTargetFunction := boundFunc.BoundTargetFunction.Value.(FunctionInterface)
		return GetFunctionRealm(runtime, boundTargetFunction)
	}

	// TODO: Handle proxy exotic objects according to the spec.

	return NewNormalCompletion(runtime.GetRunningRealm())
}

func Call(runtime *Runtime, function *JavaScriptValue, thisArg *JavaScriptValue, arguments []*JavaScriptValue) *Completion {
	if !IsCallable(function) {
		return NewThrowCompletion(NewTypeError(runtime, "Function is not callable."))
	}

	if functionObj, ok := function.Value.(FunctionInterface); ok {
		return functionObj.Call(runtime, thisArg, arguments)
	}

	if boundFunc, ok := function.Value.(*BoundFunction); ok {
		return boundFunc.Call(runtime, thisArg, arguments)
	}

	panic("Assert failed: Function is not a function or bound function.")
}

func IsCallable(value *JavaScriptValue) bool {
	if value.Type != TypeObject {
		return false
	}

	if _, ok := value.Value.(FunctionInterface); ok {
		return true
	}

	if _, ok := value.Value.(*BoundFunction); ok {
		return true
	}

	return false
}

func PrivateMethodOrAccessorAdd(runtime *Runtime, object ObjectInterface, method *PrivateElement) *Completion {
	if method.Kind != PrivateElementKindMethod && method.Kind != PrivateElementKindAccessor {
		panic("Assert failed: PrivateMethodOrAccessorAdd called on a non-method or accessor.")
	}

	entry := PrivateElementFind(object, method.Key)
	if entry != nil {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot add a private method or accessor that already exists."))
	}

	object.SetPrivateElements(append(object.GetPrivateElements(), method))
	return NewUnusedCompletion()
}

func PrivateElementFind(object ObjectInterface, key PrivateName) *PrivateElement {
	for _, privateMethod := range object.GetPrivateElements() {
		if privateMethod.Key.Description == key.Description {
			return privateMethod
		}
	}

	return nil
}

func PrivateFieldAdd(runtime *Runtime, receiver ObjectInterface, fieldName PrivateName, initValue *JavaScriptValue) *Completion {
	entry := PrivateElementFind(receiver, fieldName)
	if entry != nil {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot add a private field that already exists."))
	}

	receiver.SetPrivateElements(append(receiver.GetPrivateElements(), &PrivateElement{
		Key:   fieldName,
		Kind:  PrivateElementKindField,
		Value: initValue,
	}))

	return NewUnusedCompletion()
}

func DefineField(runtime *Runtime, receiver ObjectInterface, field *ClassFieldDefinition) *Completion {
	var initValue *JavaScriptValue = nil
	if field.Initializer != nil {
		completion := Call(
			runtime,
			NewJavaScriptValue(TypeObject, field.Initializer),
			NewJavaScriptValue(TypeObject, receiver),
			[]*JavaScriptValue{},
		)
		if completion.Type != Normal {
			return completion
		}

		initValue = completion.Value.(*JavaScriptValue)
	} else {
		initValue = NewUndefinedValue()
	}

	if field.Name.Type == TypePrivateName {
		completion := PrivateFieldAdd(runtime, receiver, field.Name.Value.(PrivateName), initValue)
		if completion.Type != Normal {
			return completion
		}
	} else {
		completion := CreateDataProperty(runtime, receiver, field.Name, initValue)
		if completion.Type != Normal {
			return completion
		}
		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError(runtime, "Failed to define field."))
		}
	}

	return NewUnusedCompletion()
}
