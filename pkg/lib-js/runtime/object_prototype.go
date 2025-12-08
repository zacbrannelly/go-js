package runtime

import "fmt"

type ObjectPrototype struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool
}

func NewObjectPrototype(runtime *Runtime) ObjectInterface {
	objectProto := &ObjectPrototype{
		Prototype:        nil,
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		Extensible:       true,
	}

	// Object.prototype.hasOwnProperty
	DefineBuiltinFunction(runtime, objectProto, "hasOwnProperty", ObjectPrototypeHasOwnProperty, 1)

	// Object.prototype.isPrototypeOf
	DefineBuiltinFunction(runtime, objectProto, "isPrototypeOf", ObjectPrototypeIsPrototypeOf, 1)

	// Object.prototype.propertyIsEnumerable
	DefineBuiltinFunction(runtime, objectProto, "propertyIsEnumerable", ObjectPrototypePropertyIsEnumerable, 1)

	// Object.prototype.toLocaleString
	DefineBuiltinFunction(runtime, objectProto, "toLocaleString", ObjectPrototypeToLocaleString, 0)

	// Object.prototype.toString
	DefineBuiltinFunction(runtime, objectProto, "toString", ObjectPrototypeToString, 0)

	return objectProto
}

func ObjectPrototypeHasOwnProperty(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToPropertyKey(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	propertyKey := completion.Value.(*JavaScriptValue)

	completion = ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	return HasOwnProperty(object, propertyKey)
}

func ObjectPrototypeIsPrototypeOf(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	value := arguments[0]
	if value.Type != TypeObject {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	completion := ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue)
	valueObj := value.Value.(ObjectInterface)

	for {
		completion := valueObj.GetPrototypeOf()
		if completion.Type != Normal {
			return completion
		}

		if completion.Value.(*JavaScriptValue).Type == TypeNull {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		value = completion.Value.(*JavaScriptValue)
		valueObj = value.Value.(ObjectInterface)

		completion = SameValue(object, value)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(NewBooleanValue(true))
		}
	}
}

func ObjectPrototypePropertyIsEnumerable(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToPropertyKey(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	propertyKey := completion.Value.(*JavaScriptValue)

	completion = ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	completion = object.GetOwnProperty(propertyKey)
	if completion.Type != Normal {
		return completion
	}

	if descriptor, ok := completion.Value.(PropertyDescriptor); ok && descriptor != nil {
		return NewNormalCompletion(NewBooleanValue(descriptor.GetEnumerable()))
	}

	return NewNormalCompletion(NewBooleanValue(false))
}

func ObjectPrototypeToLocaleString(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	completion = object.Get(runtime, NewStringValue("toString"), thisArg)
	if completion.Type != Normal {
		return completion
	}

	if functionObject, ok := completion.Value.(*JavaScriptValue).Value.(*FunctionObject); ok {
		return functionObject.Call(runtime, thisArg, []*JavaScriptValue{})
	}

	return NewThrowCompletion(NewTypeError("'this' doesn't have a callable 'toString' method"))
}

func ObjectPrototypeToString(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if thisArg.Type == TypeUndefined {
		return NewNormalCompletion(NewStringValue("[object Undefined]"))
	}

	if thisArg.Type == TypeNull {
		return NewNormalCompletion(NewStringValue("[object Null]"))
	}

	completion := ToObject(thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	completion = IsArray(thisArg)
	if completion.Type != Normal {
		return completion
	}

	tag := "Object"

	// Array objects.
	if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		tag = "Array"
	}

	// Function objects.
	if _, ok := object.GetPrototype().(*FunctionObject); ok {
		tag = "Function"
	}

	// TODO: Detect "Arguments" object.
	// TODO: Detect "Error" object.
	// TODO: Detect "Date" object.
	// TODO: Detect "RegExp" object.
	// TODO: Detect "String" object.
	// TODO: Detect "Number" object.
	// TODO: Detect "Boolean" object.
	// TODO: Detect "String" object.

	// Check if the object has a Symbol.toStringTag property.
	completion = object.Get(runtime, runtime.SymbolToStringTag, thisArg)
	if completion.Type != Normal {
		return completion
	}

	// If the object has a Symbol.toStringTag property, use it.
	maybeTag := completion.Value.(*JavaScriptValue)
	if maybeTag.Type == TypeString {
		tag = maybeTag.Value.(*String).Value
	}

	return NewNormalCompletion(NewStringValue(fmt.Sprintf("[object %s]", tag)))
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

func (o *ObjectPrototype) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *ObjectPrototype) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
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

func (o *ObjectPrototype) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *ObjectPrototype) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *ObjectPrototype) Delete(key *JavaScriptValue) *Completion {
	return OrdinaryDelete(o, key)
}
