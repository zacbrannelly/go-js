package runtime

import (
	"slices"

	"zbrannelly.dev/go-js/pkg/lib-js/parser/ast"
)

type ArgumentsObject struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool

	ParameterMap ObjectInterface
}

func CreateUnmappedArgumentsObject(runtime *Runtime, arguments []*JavaScriptValue) ObjectInterface {
	obj := &ArgumentsObject{
		Prototype:        runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype),
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		Extensible:       true,
		ParameterMap:     nil,
	}

	// "length" property.
	completion := DefinePropertyOrThrow(runtime, obj, lengthStr, &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(len(arguments)), false),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	if completion.Type != Normal {
		panic("Assert failed: DefinePropertyOrThrow threw an unexpected error.")
	}

	for idx, arg := range arguments {
		completion = ToString(NewNumberValue(float64(idx), false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}

		key := completion.Value.(*JavaScriptValue)

		completion = CreateDataProperty(runtime, obj, key, arg)

		if completion.Type != Normal {
			panic("Assert failed: CreateDataProperty threw an unexpected error.")
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: CreateDataProperty did not create a data property.")
		}
	}

	// %Symbol.iterator% property.
	DefineBuiltinSymbolFunction(runtime, obj, runtime.SymbolIterator, ArrayPrototypeValues, 0)

	// "callee" property.
	DefineBuiltinAccessorFunction(
		runtime,
		obj,
		"callee",
		// TODO: Replace with %ThrowTypeError% intrinsic.
		func(runtime *Runtime, function *FunctionObject, thisArg *JavaScriptValue, arguments []*JavaScriptValue, newTarget *JavaScriptValue) *Completion {
			return NewThrowCompletion(NewTypeError(runtime, "Cannot access callee of an arguments object."))
		},
		// TODO: Replace with %ThrowTypeError% intrinsic.
		func(runtime *Runtime, function *FunctionObject, thisArg *JavaScriptValue, arguments []*JavaScriptValue, newTarget *JavaScriptValue) *Completion {
			return NewThrowCompletion(NewTypeError(runtime, "Cannot set callee of an arguments object."))
		},
		&AccessorPropertyDescriptor{
			Enumerable:   false,
			Configurable: false,
		},
	)

	return obj
}

func CreateMappedArgumentsObject(
	runtime *Runtime,
	function ObjectInterface,
	formals []ast.Node,
	arguments []*JavaScriptValue,
	env Environment,
) ObjectInterface {
	parameterMap := OrdinaryObjectCreate(nil)

	obj := &ArgumentsObject{
		Prototype:        runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype),
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		Extensible:       true,
		ParameterMap:     parameterMap,
	}

	parameterNames := make([]string, 0)
	for _, formal := range formals {
		parameterNames = append(parameterNames, BoundNames(formal)...)
	}

	for idx := range len(arguments) {
		completion := ToString(NewNumberValue(float64(idx), false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}

		key := completion.Value.(*JavaScriptValue)

		completion = CreateDataProperty(runtime, parameterMap, key, arguments[idx])
		if completion.Type != Normal {
			panic("Assert failed: CreateDataProperty threw an unexpected error.")
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: CreateDataProperty did not create a data property.")
		}
	}

	// "length" property.
	completion := DefinePropertyOrThrow(runtime, obj, lengthStr, &DataPropertyDescriptor{
		Value:        NewNumberValue(float64(len(arguments)), false),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: DefinePropertyOrThrow threw an unexpected error.")
	}

	mappedNames := make([]string, 0)

	for i := len(parameterNames) - 1; i >= 0; i-- {
		name := parameterNames[i]
		if slices.Contains(mappedNames, name) {
			continue
		}

		mappedNames = append(mappedNames, name)

		if i >= len(arguments) {
			continue
		}

		completion = ToString(NewNumberValue(float64(i), false))
		if completion.Type != Normal {
			panic("Assert failed: ToString threw an unexpected error.")
		}

		key := completion.Value.(*JavaScriptValue)

		completion = parameterMap.DefineOwnProperty(runtime, key, &AccessorPropertyDescriptor{
			Get:          MakeArgGetter(runtime, name, env),
			Set:          MakeArgSetter(runtime, name, env),
			Enumerable:   false,
			Configurable: true,
		})
		if completion.Type != Normal {
			panic("Assert failed: parameterMap.DefineOwnProperty threw an unexpected error.")
		}
	}

	// %Symbol.iterator% property.
	DefineBuiltinSymbolFunction(runtime, obj, runtime.SymbolIterator, ArrayPrototypeValues, 0)

	// "callee" property.
	completion = DefinePropertyOrThrow(runtime, obj, NewStringValue("callee"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, function),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
	if completion.Type != Normal {
		panic("Assert failed: DefinePropertyOrThrow threw an unexpected error.")
	}

	return obj
}

func MakeArgGetter(runtime *Runtime, name string, env Environment) FunctionInterface {
	getterClosure := func(
		runtime *Runtime,
		function *FunctionObject,
		thisArg *JavaScriptValue,
		arguments []*JavaScriptValue,
		newTarget *JavaScriptValue,
	) *Completion {
		return env.GetBindingValue(runtime, name, false)
	}
	getter := CreateBuiltinFunction(runtime, getterClosure, 0, NewStringValue(""), nil, nil)
	return getter
}

func MakeArgSetter(runtime *Runtime, name string, env Environment) FunctionInterface {
	setterClosure := func(
		runtime *Runtime,
		function *FunctionObject,
		thisArg *JavaScriptValue,
		arguments []*JavaScriptValue,
		newTarget *JavaScriptValue,
	) *Completion {
		if len(arguments) < 1 {
			arguments = append(arguments, NewUndefinedValue())
		}

		return env.SetMutableBinding(runtime, name, arguments[0], false)
	}
	setter := CreateBuiltinFunction(runtime, setterClosure, 1, NewStringValue(""), nil, nil)
	return setter
}

func (o *ArgumentsObject) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *ArgumentsObject) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *ArgumentsObject) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *ArgumentsObject) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *ArgumentsObject) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *ArgumentsObject) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
}

func (o *ArgumentsObject) GetExtensible() bool {
	return o.Extensible
}

func (o *ArgumentsObject) SetExtensible(extensible bool) {
	o.Extensible = extensible
}

func (o *ArgumentsObject) GetPrototypeOf() *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *ArgumentsObject) SetPrototypeOf(prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(o, prototype)
}

func (o *ArgumentsObject) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	objVal := NewJavaScriptValue(TypeObject, o)

	completion := OrdinaryGetOwnProperty(runtime, o, key)
	if o.ParameterMap == nil {
		return completion
	}

	if completion.Value == nil {
		return NewNormalCompletion(NewUndefinedValue())
	}

	propertyDesc := completion.Value.(PropertyDescriptor).Copy()

	completion = HasOwnProperty(runtime, o.ParameterMap, key)
	if completion.Type != Normal {
		panic("Assert failed: HasOwnProperty threw an error.")
	}

	isMapped := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if isMapped {
		completion = o.ParameterMap.Get(runtime, key, objVal)
		if completion.Type != Normal {
			panic("Assert failed: Get threw an error.")
		}

		if dataDesc, ok := propertyDesc.(*DataPropertyDescriptor); ok {
			dataDesc.Value = completion.Value.(*JavaScriptValue)
		} else if _, ok := propertyDesc.(*AccessorPropertyDescriptor); ok {
			// TODO: Currently accessor descriptors don't have value slots.
			// TODO: Should this be possible?
			panic("Assert failed: Accessor property descriptor is not supported for mapped arguments objects.")
		}
	}

	return NewNormalCompletion(propertyDesc.Copy())
}

func (o *ArgumentsObject) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *ArgumentsObject) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	if o.ParameterMap == nil {
		return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
	}

	parameterMapVal := NewJavaScriptValue(TypeObject, o.ParameterMap)

	completion := HasOwnProperty(runtime, o.ParameterMap, key)
	if completion.Type != Normal {
		panic("Assert failed: HasOwnProperty threw an error.")
	}

	isMapped := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	newArgDesc := descriptor

	if isMapped {
		if dataDesc, ok := descriptor.(*DataPropertyDescriptor); ok {
			if dataDesc.Value == nil && !dataDesc.Writable {

				completion = o.ParameterMap.Get(runtime, key, parameterMapVal)
				if completion.Type != Normal {
					panic("Assert failed: Get threw an error.")
				}

				newArgDesc = descriptor.Copy()
				newArgDesc.(*DataPropertyDescriptor).Value = completion.Value.(*JavaScriptValue)
			}
		}
	}

	completion = OrdinaryDefineOwnProperty(runtime, o, key, newArgDesc)
	if completion.Type != Normal {
		panic("Assert failed: OrdinaryDefineOwnProperty threw an error.")
	}

	allowed := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !allowed {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	if isMapped {
		if _, ok := descriptor.(*AccessorPropertyDescriptor); ok {
			completion = o.ParameterMap.Delete(runtime, key)
			if completion.Type != Normal {
				panic("Assert failed: Delete threw an error.")
			}
		} else {
			dataDesc := descriptor.(*DataPropertyDescriptor)
			if dataDesc.Value != nil {
				completion = o.ParameterMap.Set(runtime, key, dataDesc.Value, parameterMapVal)
				if completion.Type != Normal {
					panic("Assert failed: Set threw an error.")
				}
			}

			if !dataDesc.Writable {
				completion = o.ParameterMap.Delete(runtime, key)
				if completion.Type != Normal {
					panic("Assert failed: Delete threw an error.")
				}
			}
		}
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *ArgumentsObject) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	if o.ParameterMap == nil {
		return OrdinarySet(runtime, o, key, value, receiver)
	}

	objectVal := NewJavaScriptValue(TypeObject, o)

	completion := SameValue(value, objectVal)
	if completion.Type != Normal {
		panic("Assert failed: SameValue threw an error.")
	}

	isSame := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	isMapped := false
	if !isSame {
		completion = HasOwnProperty(runtime, o.ParameterMap, key)
		if completion.Type != Normal {
			panic("Assert failed: HasOwnProperty threw an error.")
		}

		isMapped = completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	}

	if isMapped {
		completion = o.ParameterMap.Set(runtime, key, value, objectVal)
		if completion.Type != Normal {
			panic("Assert failed: Set threw an error.")
		}
	}

	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *ArgumentsObject) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	if o.ParameterMap == nil {
		return OrdinaryGet(runtime, o, key, receiver)
	}

	completion := HasOwnProperty(runtime, o.ParameterMap, key)
	if completion.Type != Normal {
		panic("Assert failed: HasOwnProperty threw an error.")
	}

	isMapped := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if isMapped {
		return o.ParameterMap.Get(runtime, key, NewJavaScriptValue(TypeObject, o.ParameterMap))
	}

	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *ArgumentsObject) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	if o.ParameterMap == nil {
		return OrdinaryDelete(runtime, o, key)
	}

	completion := HasOwnProperty(runtime, o.ParameterMap, key)
	if completion.Type != Normal {
		panic("Assert failed: HasOwnProperty threw an error.")
	}

	isMapped := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	result := OrdinaryDelete(runtime, o, key)
	if completion.Type != Normal {
		panic("Assert failed: OrdinaryDelete threw an error.")
	}

	allowed := result.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if allowed && isMapped {
		completion = o.ParameterMap.Delete(runtime, key)
		if completion.Type != Normal {
			panic("Assert failed: Delete threw an error.")
		}
	}

	return result
}

func (o *ArgumentsObject) OwnPropertyKeys() *Completion {
	return NewNormalCompletion(OrdinaryOwnPropertyKeys(o))
}

func (o *ArgumentsObject) PreventExtensions() *Completion {
	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}
