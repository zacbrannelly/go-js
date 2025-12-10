package runtime

func NewObjectConstructor(runtime *Runtime) *FunctionObject {
	realm := runtime.GetRunningRealm()
	constructor := CreateBuiltinFunction(
		runtime,
		ObjectConstructor,
		1,
		NewStringValue("Object"),
		realm,
		realm.Intrinsics[IntrinsicFunctionPrototype],
	)
	MakeConstructor(constructor)

	// Object.assign
	DefineBuiltinFunction(runtime, constructor, "assign", ObjectAssign, 2)

	// Object.create
	DefineBuiltinFunction(runtime, constructor, "create", ObjectCreate, 2)

	// Object.defineProperties
	DefineBuiltinFunction(runtime, constructor, "defineProperties", ObjectDefinePropertiesFunc, 2)

	// Object.defineProperty
	DefineBuiltinFunction(runtime, constructor, "defineProperty", ObjectDefineProperty, 3)

	return constructor
}

func ObjectConstructor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	activeFunction := runtime.GetRunningExecutionContext().Function
	if newTarget != nil && newTarget.Type != TypeUndefined && newTarget.Value != activeFunction {
		panic("TODO: Support NewTarget in Object constructor.")
	}

	if len(arguments) == 0 || arguments[0].Type == TypeUndefined || arguments[0].Type == TypeNull {
		newObj := OrdinaryObjectCreate(function.Realm.Intrinsics[IntrinsicObjectPrototype])
		return NewNormalCompletion(NewJavaScriptValue(TypeObject, newObj))
	}

	completion := ToObject(arguments[0])
	if completion.Type != Normal {
		panic("Assert failed: ToObject threw an error when it should not have.")
	}

	return completion
}

func ObjectAssign(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	target := arguments[0]
	if len(arguments) == 1 {
		return NewNormalCompletion(target)
	}

	completion := ToObject(target)
	if completion.Type != Normal {
		return completion
	}

	target = completion.Value.(*JavaScriptValue)
	targetObj := target.Value.(ObjectInterface)

	for _, source := range arguments[1:] {
		if source.Type == TypeUndefined || source.Type == TypeNull {
			continue
		}

		completion = ToObject(source)
		if completion.Type != Normal {
			panic("Assert failed: ToObject threw an error when it should not have.")
		}

		fromObj := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)

		completion = fromObj.OwnPropertyKeys()
		if completion.Type != Normal {
			return completion
		}

		keys := completion.Value.([]*JavaScriptValue)
		for _, key := range keys {
			completion = fromObj.GetOwnProperty(key)
			if completion.Type != Normal {
				return completion
			}

			if desc, ok := completion.Value.(PropertyDescriptor); !ok || desc == nil || !desc.GetEnumerable() {
				continue
			}

			completion = fromObj.Get(runtime, key, source)
			if completion.Type != Normal {
				return completion
			}

			propValue := completion.Value.(*JavaScriptValue)
			completion = targetObj.Set(runtime, key, propValue, target)
			if completion.Type != Normal {
				return completion
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				// TODO: Improve the error message.
				return NewThrowCompletion(NewTypeError("Failed to set property."))
			}
		}
	}

	return NewNormalCompletion(target)
}

func ObjectCreate(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		arguments = append(arguments, NewUndefinedValue())
	}

	objectArg := arguments[0]

	if objectArg.Type != TypeObject && objectArg.Type != TypeNull {
		return NewThrowCompletion(NewTypeError("Object prototype may only be an Object or null"))
	}

	var prototype ObjectInterface = nil
	if objectArg.Type == TypeObject {
		prototype = objectArg.Value.(ObjectInterface)
	}

	resultObj := OrdinaryObjectCreate(prototype)

	if len(arguments) > 1 {
		properties := arguments[1]
		if properties.Type != TypeUndefined {
			return ObjectDefineProperties(runtime, resultObj, properties)
		}
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, resultObj))
}

func ObjectDefinePropertiesFunc(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	object := arguments[0]
	if object.Type != TypeObject {
		return NewThrowCompletion(NewTypeError("Object.defineProperties must be called with an object as the first argument"))
	}

	return ObjectDefineProperties(runtime, object.Value.(ObjectInterface), arguments[1])
}

func ObjectDefineProperty(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	object := arguments[0]
	if object.Type != TypeObject {
		return NewThrowCompletion(NewTypeError("Object.defineProperties must be called with an object as the first argument"))
	}

	completion := ToPropertyKey(arguments[1])
	if completion.Type != Normal {
		return completion
	}

	key := completion.Value.(*JavaScriptValue)

	completion = ToPropertyDescriptor(runtime, arguments[2])
	if completion.Type != Normal {
		return completion
	}

	descriptor := completion.Value.(*JavaScriptValue).Value.(PropertyDescriptor)

	completion = DefinePropertyOrThrow(object.Value.(ObjectInterface), key, descriptor)
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(object)
}

type DescriptorPair struct {
	Key        *JavaScriptValue
	Descriptor PropertyDescriptor
}

func ObjectDefineProperties(
	runtime *Runtime,
	object ObjectInterface,
	properties *JavaScriptValue,
) *Completion {
	completion := ToObject(properties)
	if completion.Type != Normal {
		return completion
	}

	propsValue := completion.Value.(*JavaScriptValue)
	props := propsValue.Value.(ObjectInterface)

	completion = props.OwnPropertyKeys()
	if completion.Type != Normal {
		return completion
	}

	keys := completion.Value.([]*JavaScriptValue)

	descriptors := make([]DescriptorPair, 0)

	for _, key := range keys {
		completion = props.GetOwnProperty(key)
		if completion.Type != Normal {
			return completion
		}

		desc, ok := completion.Value.(PropertyDescriptor)
		if !ok || desc == nil || !desc.GetEnumerable() {
			continue
		}

		completion = props.Get(runtime, key, propsValue)
		if completion.Type != Normal {
			return completion
		}

		descObj := completion.Value.(*JavaScriptValue)
		completion = ToPropertyDescriptor(runtime, descObj)
		if completion.Type != Normal {
			return completion
		}

		desc = completion.Value.(*JavaScriptValue).Value.(PropertyDescriptor)
		descriptors = append(descriptors, DescriptorPair{Key: key, Descriptor: desc})
	}

	for _, descriptor := range descriptors {
		completion = DefinePropertyOrThrow(object, descriptor.Key, descriptor.Descriptor)
		if completion.Type != Normal {
			return completion
		}
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, object))
}

var enumerableKey = NewStringValue("enumerable")
var configurableKey = NewStringValue("configurable")
var writableKey = NewStringValue("writable")
var valueKey = NewStringValue("value")
var getKey = NewStringValue("get")
var setKey = NewStringValue("set")

func ToPropertyDescriptor(runtime *Runtime, value *JavaScriptValue) *Completion {
	if value.Type != TypeObject {
		return NewThrowCompletion(NewTypeError("Invalid property descriptor"))
	}

	obj := value.Value.(ObjectInterface)

	completion := GetBoolPropertyFromObject(runtime, value, obj, enumerableKey)
	if completion.Type != Normal {
		return completion
	}
	enumerable := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	completion = GetBoolPropertyFromObject(runtime, value, obj, configurableKey)
	if completion.Type != Normal {
		return completion
	}
	configurable := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	completion = GetValuePropertyFromObject(runtime, value, obj, valueKey)
	if completion.Type != Normal {
		return completion
	}

	valueSlot, ok := completion.Value.(*JavaScriptValue)
	if !ok {
		valueSlot = NewUndefinedValue()
	}

	completion = GetBoolPropertyFromObject(runtime, value, obj, writableKey)
	if completion.Type != Normal {
		return completion
	}
	writable := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	completion = GetValuePropertyFromObject(runtime, value, obj, getKey)
	if completion.Type != Normal {
		return completion
	}

	var getSlot *FunctionObject = nil
	if get, ok := completion.Value.(*JavaScriptValue); ok && get != nil {
		if getSlot, ok = get.Value.(*FunctionObject); !ok {
			return NewThrowCompletion(NewTypeError("get property must be a function"))
		}
	}

	completion = GetValuePropertyFromObject(runtime, value, obj, setKey)
	if completion.Type != Normal {
		return completion
	}

	var setSlot *FunctionObject = nil
	if set, ok := completion.Value.(*JavaScriptValue); ok && set != nil {
		if setSlot, ok = set.Value.(*FunctionObject); !ok {
			return NewThrowCompletion(NewTypeError("set property must be a function"))
		}
	}

	var desc PropertyDescriptor = nil
	if getSlot != nil || setSlot != nil {
		desc = &AccessorPropertyDescriptor{
			Get:          getSlot,
			Set:          setSlot,
			Enumerable:   enumerable,
			Configurable: configurable,
		}
	} else {
		desc = &DataPropertyDescriptor{
			Value:        valueSlot,
			Writable:     writable,
			Enumerable:   enumerable,
			Configurable: configurable,
		}
	}

	return NewNormalCompletion(NewJavaScriptValue(TypePropertyDescriptor, desc))
}

func GetBoolPropertyFromObject(
	runtime *Runtime,
	objValue *JavaScriptValue,
	obj ObjectInterface,
	key *JavaScriptValue,
) *Completion {
	completion := obj.HasProperty(key)
	if completion.Type != Normal {
		return completion
	}

	hasKey := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !hasKey {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	completion = obj.Get(runtime, key, objValue)
	if completion.Type != Normal {
		return completion
	}

	completion = ToBoolean(completion.Value.(*JavaScriptValue))
	if completion.Type != Normal {
		return completion
	}

	return completion
}

func GetValuePropertyFromObject(
	runtime *Runtime,
	objValue *JavaScriptValue,
	obj ObjectInterface,
	key *JavaScriptValue,
) *Completion {
	completion := obj.HasProperty(key)
	if completion.Type != Normal {
		return completion
	}

	hasKey := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !hasKey {
		// Nil to signal not found.
		return NewNormalCompletion(nil)
	}

	return obj.Get(runtime, key, objValue)
}
