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

	return objectProto
}

func DefineObjectPrototypeProperties(runtime *Runtime, objectProto *ObjectPrototype) {
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

	// Object.prototype.valueOf
	DefineBuiltinFunction(runtime, objectProto, "valueOf", ObjectPrototypeValueOf, 0)

	// Object.prototype.__proto__
	DefineBuiltinAccessorFunction(
		runtime,
		objectProto,
		"__proto__",
		ObjectPrototypeProtoGetter,
		ObjectPrototypeProtoSetter,
		&AccessorPropertyDescriptor{
			Enumerable:   false,
			Configurable: true,
		},
	)

	// Object.prototype.__defineGetter__
	DefineBuiltinFunction(runtime, objectProto, "__defineGetter__", ObjectPrototypeDefineGetter, 2)

	// Object.prototype.__defineSetter__
	DefineBuiltinFunction(runtime, objectProto, "__defineSetter__", ObjectPrototypeDefineSetter, 2)

	// Object.prototype.__lookupGetter__
	DefineBuiltinFunction(runtime, objectProto, "__lookupGetter__", ObjectPrototypeLookupGetter, 1)

	// Object.prototype.__lookupSetter__
	DefineBuiltinFunction(runtime, objectProto, "__lookupSetter__", ObjectPrototypeLookupSetter, 1)
}

func ObjectPrototypeHasOwnProperty(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToPropertyKey(runtime, arguments[0])
	if completion.Type != Normal {
		return completion
	}

	propertyKey := completion.Value.(*JavaScriptValue)

	completion = ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	return HasOwnProperty(runtime, object, propertyKey)
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

	completion := ToObject(runtime, thisArg)
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
	completion := ToPropertyKey(runtime, arguments[0])
	if completion.Type != Normal {
		return completion
	}

	propertyKey := completion.Value.(*JavaScriptValue)

	completion = ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	completion = object.GetOwnProperty(runtime, propertyKey)
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
	completion := ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	completion = object.Get(runtime, NewStringValue("toString"), thisArg)
	if completion.Type != Normal {
		return completion
	}

	if functionObject, ok := completion.Value.(*JavaScriptValue).Value.(FunctionInterface); ok {
		return functionObject.Call(runtime, thisArg, []*JavaScriptValue{})
	}

	return NewThrowCompletion(NewTypeError(runtime, "'this' doesn't have a callable 'toString' method"))
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

	completion := ToObject(runtime, thisArg)
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
	if _, ok := object.(FunctionInterface); ok {
		tag = "Function"
	}

	// Error objects.
	if obj, ok := object.(*Object); ok && obj.IsError {
		tag = "Error"
	}

	// TODO: Detect "Arguments" object.
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

func ObjectPrototypeValueOf(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	return ToObject(runtime, thisArg)
}

func ObjectPrototypeProtoGetter(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	return object.GetPrototypeOf()
}

func ObjectPrototypeProtoSetter(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if thisArg.Type == TypeUndefined || thisArg.Type == TypeNull {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot set prototype to undefined or null"))
	}

	proto := arguments[0]

	if proto.Type != TypeObject {
		return NewNormalCompletion(NewUndefinedValue())
	}

	object := thisArg.Value.(ObjectInterface)

	completion := object.SetPrototypeOf(proto)
	if completion.Type != Normal {
		return completion
	}

	status := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !status {
		// TODO: Improve the error message.
		return NewThrowCompletion(NewTypeError(runtime, "Invalid prototype object"))
	}

	return NewNormalCompletion(NewUndefinedValue())
}

func ObjectPrototypeDefineGetter(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	propertyKey := arguments[0]
	getter := arguments[1]

	completion := ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	if _, ok := getter.Value.(FunctionInterface); !ok {
		return NewThrowCompletion(NewTypeError(runtime, "Getter must be a function"))
	}

	desc := &AccessorPropertyDescriptor{
		Get:          getter.Value.(FunctionInterface),
		Enumerable:   true,
		Configurable: true,
	}

	completion = ToPropertyKey(runtime, propertyKey)
	if completion.Type != Normal {
		return completion
	}

	propertyKey = completion.Value.(*JavaScriptValue)

	completion = DefinePropertyOrThrow(runtime, object, propertyKey, desc)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(NewUndefinedValue())
}

func ObjectPrototypeDefineSetter(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	propertyKey := arguments[0]
	setter := arguments[1]

	completion := ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	if _, ok := setter.Value.(FunctionInterface); !ok {
		return NewThrowCompletion(NewTypeError(runtime, "Getter must be a function"))
	}

	desc := &AccessorPropertyDescriptor{
		Get:          setter.Value.(FunctionInterface),
		Enumerable:   true,
		Configurable: true,
	}

	completion = ToPropertyKey(runtime, propertyKey)
	if completion.Type != Normal {
		return completion
	}

	propertyKey = completion.Value.(*JavaScriptValue)

	completion = DefinePropertyOrThrow(runtime, object, propertyKey, desc)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(NewUndefinedValue())
}

func ObjectPrototypeLookupGetter(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	propertyKey := arguments[0]

	completion := ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	completion = ToPropertyKey(runtime, propertyKey)
	if completion.Type != Normal {
		return completion
	}

	propertyKey = completion.Value.(*JavaScriptValue)

	for {
		completion = object.GetOwnProperty(runtime, propertyKey)
		if completion.Type != Normal {
			return completion
		}

		if descriptor, ok := completion.Value.(PropertyDescriptor); ok && descriptor != nil {
			if accessorDescriptor, ok := descriptor.(*AccessorPropertyDescriptor); ok {
				getter := accessorDescriptor.Get
				if getter == nil {
					return NewNormalCompletion(NewUndefinedValue())
				}
				return NewNormalCompletion(NewJavaScriptValue(TypeObject, accessorDescriptor.Get))
			}

			return NewNormalCompletion(NewUndefinedValue())
		}

		object = object.GetPrototype()
		if object == nil {
			return NewNormalCompletion(NewUndefinedValue())
		}
	}
}

func ObjectPrototypeLookupSetter(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	propertyKey := arguments[0]

	completion := ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

	completion = ToPropertyKey(runtime, propertyKey)
	if completion.Type != Normal {
		return completion
	}

	propertyKey = completion.Value.(*JavaScriptValue)

	for {
		completion = object.GetOwnProperty(runtime, propertyKey)
		if completion.Type != Normal {
			return completion
		}

		if descriptor, ok := completion.Value.(PropertyDescriptor); ok && descriptor != nil {
			if accessorDescriptor, ok := descriptor.(*AccessorPropertyDescriptor); ok {
				setter := accessorDescriptor.Set
				if setter == nil {
					return NewNormalCompletion(NewUndefinedValue())
				}
				return NewNormalCompletion(NewJavaScriptValue(TypeObject, setter))
			}

			return NewNormalCompletion(NewUndefinedValue())
		}

		object = object.GetPrototype()
		if object == nil {
			return NewNormalCompletion(NewUndefinedValue())
		}
	}
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

func (o *ObjectPrototype) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(runtime, o, key)
}

func (o *ObjectPrototype) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *ObjectPrototype) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
}

func (o *ObjectPrototype) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *ObjectPrototype) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *ObjectPrototype) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryDelete(runtime, o, key)
}

func (o *ObjectPrototype) OwnPropertyKeys() *Completion {
	return NewNormalCompletion(OrdinaryOwnPropertyKeys(o))
}

func (o *ObjectPrototype) PreventExtensions() *Completion {
	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}
