package runtime

type BoundFunction struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool
	PrivateElements  []*PrivateElement

	BoundTargetFunction *JavaScriptValue
	BoundThis           *JavaScriptValue
	BoundArguments      []*JavaScriptValue

	HasConstruct bool
}

func BoundFunctionCreate(
	runtime *Runtime,
	targetFunction ObjectInterface,
	boundThis *JavaScriptValue,
	boundArgs []*JavaScriptValue,
) *Completion {
	completion := targetFunction.GetPrototypeOf(runtime)
	if completion.Type != Normal {
		return completion
	}

	targetFunctionObj, ok := targetFunction.(FunctionInterface)
	if !ok {
		panic("Assert failed: Target function is not a function.")
	}

	prototype := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	boundFunc := &BoundFunction{
		Prototype:           prototype,
		Properties:          make(map[string]PropertyDescriptor),
		SymbolProperties:    make(map[*Symbol]PropertyDescriptor),
		Extensible:          true,
		PrivateElements:     make([]*PrivateElement, 0),
		BoundTargetFunction: NewJavaScriptValue(TypeObject, targetFunction),
		BoundThis:           boundThis,
		BoundArguments:      boundArgs,
		HasConstruct:        targetFunctionObj.HasConstructMethod(),
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, boundFunc))
}

func (o *BoundFunction) Call(runtime *Runtime, thisArg *JavaScriptValue, arguments []*JavaScriptValue) *Completion {
	finalArgs := make([]*JavaScriptValue, 0)
	finalArgs = append(finalArgs, o.BoundArguments...)
	finalArgs = append(finalArgs, arguments...)

	return Call(runtime, o.BoundTargetFunction, o.BoundThis, finalArgs)
}

func (o *BoundFunction) Construct(runtime *Runtime, arguments []*JavaScriptValue, newTarget *JavaScriptValue) *Completion {
	boundTargetFunction, ok := o.BoundTargetFunction.Value.(FunctionInterface)
	if !ok {
		return NewThrowCompletion(NewTypeError(runtime, "Bound target function is not callable."))
	}

	completion := SameValue(NewJavaScriptValue(TypeObject, o), newTarget)
	if completion.Type != Normal {
		return completion
	}

	if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		newTarget = o.BoundTargetFunction
	}

	finalArgs := make([]*JavaScriptValue, 0)
	finalArgs = append(finalArgs, o.BoundArguments...)
	finalArgs = append(finalArgs, arguments...)

	return Construct(runtime, boundTargetFunction, finalArgs, newTarget)
}

func (o *BoundFunction) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *BoundFunction) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *BoundFunction) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *BoundFunction) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *BoundFunction) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *BoundFunction) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
}

func (o *BoundFunction) IsExtensible(runtime *Runtime) *Completion {
	return NewNormalCompletion(NewBooleanValue(o.Extensible))
}

func (o *BoundFunction) GetPrototypeOf(runtime *Runtime) *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *BoundFunction) SetPrototypeOf(runtime *Runtime, prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(runtime, o, prototype)
}

func (o *BoundFunction) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(runtime, o, key)
}

func (o *BoundFunction) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *BoundFunction) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
}

func (o *BoundFunction) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *BoundFunction) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *BoundFunction) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryDelete(runtime, o, key)
}

func (o *BoundFunction) OwnPropertyKeys(runtime *Runtime) *Completion {
	return NewNormalCompletion(OrdinaryOwnPropertyKeys(o))
}

func (o *BoundFunction) PreventExtensions(runtime *Runtime) *Completion {
	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *BoundFunction) GetPrivateElements() []*PrivateElement {
	return o.PrivateElements
}

func (o *BoundFunction) SetPrivateElements(privateElements []*PrivateElement) {
	o.PrivateElements = privateElements
}

func (o *BoundFunction) HasConstructMethod() bool {
	return o.HasConstruct
}
