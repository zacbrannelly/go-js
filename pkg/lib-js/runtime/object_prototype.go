package runtime

type ObjectPrototype struct {
	Prototype  ObjectInterface
	Properties map[string]PropertyDescriptor
	Extensible bool
}

func NewObjectPrototype(runtime *Runtime) ObjectInterface {
	objectProto := &ObjectPrototype{
		Prototype:  nil,
		Properties: make(map[string]PropertyDescriptor),
		Extensible: true,
	}

	// Object.prototype.hasOwnProperty
	DefineBuiltinFunction(runtime, objectProto, "hasOwnProperty", ObjectPrototypeHasOwnProperty, 1)

	return objectProto
}

func ObjectPrototypeHasOwnProperty(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	panic("TODO: Implement ObjectPrototypeHasOwnProperty")
}

func (o *ObjectPrototype) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *ObjectPrototype) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *ObjectPrototype) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *ObjectPrototype) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *ObjectPrototype) GetExtensible() bool {
	return o.Extensible
}

func (o *ObjectPrototype) SetExtensible(extensible bool) {
	o.Extensible = extensible
}

func (o *ObjectPrototype) GetPrototypeOf() *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *ObjectPrototype) SetPrototypeOf(prototype *JavaScriptValue) *Completion {
	// Should be the same semantics as SetImmutablePrototype in the spec.
	getPrototypeOfCompletion := o.GetPrototypeOf()
	if getPrototypeOfCompletion.Type != Normal {
		return getPrototypeOfCompletion
	}

	current := getPrototypeOfCompletion.Value.(*JavaScriptValue)
	return SameValue(prototype, current)
}

func (o *ObjectPrototype) GetOwnProperty(key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(o, key)
}

func (o *ObjectPrototype) HasProperty(key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(o, key)
}

func (o *ObjectPrototype) DefineOwnProperty(key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	return OrdinaryDefineOwnProperty(o, key, descriptor)
}

func (o *ObjectPrototype) Set(key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(o, key, value, receiver)
}

func (o *ObjectPrototype) Get(key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(o, key, receiver)
}

func (o *ObjectPrototype) Delete(key *JavaScriptValue) *Completion {
	return OrdinaryDelete(o, key)
}
