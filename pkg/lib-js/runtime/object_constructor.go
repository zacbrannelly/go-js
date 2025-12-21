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

	// Object.entries
	DefineBuiltinFunction(runtime, constructor, "entries", ObjectEntries, 1)

	// Object.freeze
	DefineBuiltinFunction(runtime, constructor, "freeze", ObjectFreeze, 1)

	// Object.fromEntries
	DefineBuiltinFunction(runtime, constructor, "fromEntries", ObjectFromEntries, 1)

	// Object.getOwnPropertyDescriptor
	DefineBuiltinFunction(runtime, constructor, "getOwnPropertyDescriptor", ObjectGetOwnPropertyDescriptor, 2)

	// Object.getOwnPropertyDescriptors
	DefineBuiltinFunction(runtime, constructor, "getOwnPropertyDescriptors", ObjectGetOwnPropertyDescriptors, 1)

	// Object.getOwnPropertyNames
	DefineBuiltinFunction(runtime, constructor, "getOwnPropertyNames", ObjectGetOwnPropertyNames, 1)

	// Object.getOwnPropertySymbols
	DefineBuiltinFunction(runtime, constructor, "getOwnPropertySymbols", ObjectGetOwnPropertySymbols, 1)

	// Object.getPrototypeOf
	DefineBuiltinFunction(runtime, constructor, "getPrototypeOf", ObjectGetPrototypeOf, 1)

	// Object.groupBy
	DefineBuiltinFunction(runtime, constructor, "groupBy", ObjectGroupBy, 2)

	// Object.hasOwn
	DefineBuiltinFunction(runtime, constructor, "hasOwn", ObjectHasOwn, 2)

	// Object.is
	DefineBuiltinFunction(runtime, constructor, "is", ObjectIs, 2)

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

func ObjectEntries(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToObject(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	objectValue := completion.Value.(*JavaScriptValue)
	object := objectValue.Value.(ObjectInterface)

	completion = object.OwnPropertyKeys()
	if completion.Type != Normal {
		return completion
	}

	keys := completion.Value.([]*JavaScriptValue)

	entries := make([]*JavaScriptValue, 0)

	for _, key := range keys {
		if key.Type != TypeString {
			continue
		}

		completion = object.Get(runtime, key, objectValue)
		if completion.Type != Normal {
			return completion
		}

		entry := CreateArrayFromList(runtime, []*JavaScriptValue{key, completion.Value.(*JavaScriptValue)})
		entries = append(entries, NewJavaScriptValue(TypeObject, entry))
	}

	entriesArray := CreateArrayFromList(runtime, entries)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, entriesArray))
}

func ObjectFreeze(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		return NewNormalCompletion(NewUndefinedValue())
	}

	object := arguments[0]
	if object.Type != TypeObject {
		return NewNormalCompletion(object)
	}

	completion := SetIntegrityLevel(object.Value.(ObjectInterface), IntegrityLevelFrozen)
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError("Failed to freeze object"))
	}

	return NewNormalCompletion(object)
}

func ObjectFromEntries(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	iterable := arguments[0]
	completion := RequireObjectCoercible(iterable)
	if completion.Type != Normal {
		return completion
	}

	realm := runtime.GetRunningRealm()
	obj := OrdinaryObjectCreate(realm.Intrinsics[IntrinsicObjectPrototype])

	closure := func(
		runtime *Runtime,
		function *FunctionObject,
		thisArg *JavaScriptValue,
		arguments []*JavaScriptValue,
		newTarget *JavaScriptValue,
	) *Completion {
		key := arguments[0]
		value := arguments[1]

		completion := ToPropertyKey(key)
		if completion.Type != Normal {
			return completion
		}

		propertyKey := completion.Value.(*JavaScriptValue)
		completion = CreateDataProperty(obj, propertyKey, value)
		if completion.Type != Normal {
			panic("Assert failed: CreateDataProperty threw an unexpected error in Object.fromEntries closure.")
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			panic("Assert failed: CreateDataProperty returned false when it shouldn't have in Object.fromEntries closure.")
		}

		return NewNormalCompletion(NewUndefinedValue())
	}

	adder := CreateBuiltinFunction(runtime, closure, 2, NewStringValue(""), nil, nil)
	return AddEntriesFromIterable(runtime, obj, iterable, adder)
}

func ObjectGetOwnPropertyDescriptor(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 2 {
		undef := NewUndefinedValue()
		arguments = []*JavaScriptValue{undef, undef}
	}

	completion := ToObject(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(ObjectInterface)

	completion = ToPropertyKey(arguments[1])
	if completion.Type != Normal {
		return completion
	}

	key := completion.Value.(*JavaScriptValue)

	completion = object.GetOwnProperty(key)
	if completion.Type != Normal {
		return completion
	}

	if propertyDesc, ok := completion.Value.(PropertyDescriptor); ok && propertyDesc != nil {
		return NewNormalCompletion(FromPropertyDescriptor(runtime, propertyDesc))
	}

	return NewNormalCompletion(NewUndefinedValue())
}

func ObjectGetOwnPropertyDescriptors(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = []*JavaScriptValue{NewUndefinedValue()}
	}

	completion := ToObject(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(ObjectInterface)

	completion = object.OwnPropertyKeys()
	if completion.Type != Normal {
		return completion
	}

	keys := completion.Value.([]*JavaScriptValue)
	resultObj := OrdinaryObjectCreate(runtime.GetRunningRealm().Intrinsics[IntrinsicObjectPrototype])

	for _, key := range keys {
		completion = object.GetOwnProperty(key)
		if completion.Type != Normal {
			return completion
		}

		if descriptor, ok := completion.Value.(PropertyDescriptor); ok && descriptor != nil {
			obj := FromPropertyDescriptor(runtime, descriptor)
			completion = CreateDataProperty(resultObj, key, obj)
			if completion.Type != Normal {
				panic("Assert failed: CreateDataProperty threw an unexpected error in Object.getOwnPropertyDescriptors.")
			}
		}
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, resultObj))
}

func ObjectGetOwnPropertyNames(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = []*JavaScriptValue{NewUndefinedValue()}
	}

	completion := GetOwnPropertyKeys(runtime, arguments[0], false)
	if completion.Type != Normal {
		return completion
	}

	nameList := completion.Value.([]*JavaScriptValue)
	nameArray := CreateArrayFromList(runtime, nameList)

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, nameArray))
}

func ObjectGetOwnPropertySymbols(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = []*JavaScriptValue{NewUndefinedValue()}
	}

	completion := GetOwnPropertyKeys(runtime, arguments[0], true)
	if completion.Type != Normal {
		return completion
	}

	symbolList := completion.Value.([]*JavaScriptValue)
	symbolArray := CreateArrayFromList(runtime, symbolList)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, symbolArray))
}

func ObjectGetPrototypeOf(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) < 1 {
		arguments = []*JavaScriptValue{NewUndefinedValue()}
	}

	completion := ToObject(arguments[0])
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(ObjectInterface)

	return object.GetPrototypeOf()
}

func ObjectGroupBy(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	for idx := range 2 {
		if len(arguments) <= idx {
			arguments = append(arguments, NewUndefinedValue())
		}
	}

	items := arguments[0]
	callback := arguments[1]

	completion := GroupBy(runtime, items, callback, GroupByKeyCoercionProperty)
	if completion.Type != Normal {
		return completion
	}

	groups := completion.Value.(*GroupByResult)

	obj := OrdinaryObjectCreate(nil)
	for key, group := range groups.GroupsByString {
		elements := CreateArrayFromList(runtime, group)
		completion = CreateDataProperty(obj, NewStringValue(key), NewJavaScriptValue(TypeObject, elements))
		if completion.Type != Normal {
			panic("Assert failed: CreateDataProperty threw an unexpected error in Object.groupBy.")
		}
	}

	for key, group := range groups.GroupsBySymbol {
		elements := CreateArrayFromList(runtime, group)
		completion = CreateDataProperty(obj, NewJavaScriptValue(TypeSymbol, key), NewJavaScriptValue(TypeObject, elements))
		if completion.Type != Normal {
			panic("Assert failed: CreateDataProperty threw an unexpected error in Object.groupBy.")
		}
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, obj))
}

func ObjectHasOwn(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	for idx := range 2 {
		if len(arguments) <= idx {
			arguments = append(arguments, NewUndefinedValue())
		}
	}

	object := arguments[0]
	property := arguments[1]

	completion := ToObject(object)
	if completion.Type != Normal {
		return completion
	}

	objVal := completion.Value.(*JavaScriptValue)
	obj := objVal.Value.(ObjectInterface)

	completion = ToPropertyKey(property)
	if completion.Type != Normal {
		return completion
	}

	key := completion.Value.(*JavaScriptValue)

	return HasOwnProperty(obj, key)
}

func ObjectIs(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	for idx := range 2 {
		if len(arguments) <= idx {
			arguments = append(arguments, NewUndefinedValue())
		}
	}

	x := arguments[0]
	y := arguments[1]

	return SameValue(x, y)
}

func GetOwnPropertyKeys(
	runtime *Runtime,
	objectVal *JavaScriptValue,
	symbolsOnly bool,
) *Completion {
	completion := ToObject(objectVal)
	if completion.Type != Normal {
		return completion
	}

	objectVal = completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(ObjectInterface)

	completion = object.OwnPropertyKeys()
	if completion.Type != Normal {
		return completion
	}

	keys := completion.Value.([]*JavaScriptValue)

	nameList := make([]*JavaScriptValue, 0)
	for _, key := range keys {
		if symbolsOnly {
			if key.Type == TypeSymbol {
				nameList = append(nameList, key)
			}
		} else {
			if key.Type != TypeSymbol {
				nameList = append(nameList, key)
			}
		}
	}

	return NewNormalCompletion(nameList)
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

var (
	enumerableKey   = NewStringValue("enumerable")
	configurableKey = NewStringValue("configurable")
	writableKey     = NewStringValue("writable")
	valueKey        = NewStringValue("value")
	getKey          = NewStringValue("get")
	setKey          = NewStringValue("set")
)

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

func FromPropertyDescriptor(runtime *Runtime, descriptor PropertyDescriptor) *JavaScriptValue {
	resultObj := OrdinaryObjectCreate(runtime.GetRunningRealm().Intrinsics[IntrinsicObjectPrototype])

	if dataDescriptor, ok := descriptor.(*DataPropertyDescriptor); ok {
		CreateDataProperty(resultObj, valueKey, dataDescriptor.Value)
		CreateDataProperty(resultObj, writableKey, NewBooleanValue(dataDescriptor.Writable))
	} else {
		accessor := descriptor.(*AccessorPropertyDescriptor)
		if accessor.Get != nil {
			CreateDataProperty(resultObj, getKey, NewJavaScriptValue(TypeObject, accessor.Get))
		}
		if accessor.Set != nil {
			CreateDataProperty(resultObj, setKey, NewJavaScriptValue(TypeObject, accessor.Set))
		}
	}

	CreateDataProperty(resultObj, enumerableKey, NewBooleanValue(descriptor.GetEnumerable()))
	CreateDataProperty(resultObj, configurableKey, NewBooleanValue(descriptor.GetConfigurable()))

	return NewJavaScriptValue(TypeObject, resultObj)
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
