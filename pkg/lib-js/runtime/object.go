package runtime

import "math"

type ObjectInterface interface {
	GetPrototype() ObjectInterface
	SetPrototype(prototype ObjectInterface)

	GetProperties() map[string]PropertyDescriptor
	GetSymbolProperties() map[*Symbol]PropertyDescriptor
	SetProperties(properties map[string]PropertyDescriptor)
	SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor)

	GetExtensible() bool
	SetExtensible(extensible bool)

	// Internal methods
	GetPrototypeOf() *Completion
	SetPrototypeOf(prototype *JavaScriptValue) *Completion
	GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion
	DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion
	HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion
	Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion
	Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion
	Delete(runtime *Runtime, key *JavaScriptValue) *Completion
	OwnPropertyKeys() *Completion
	PreventExtensions() *Completion
}

func GetPropertyFromObject(object ObjectInterface, key *JavaScriptValue) (PropertyDescriptor, bool) {
	if key.Type == TypeSymbol {
		propertyDesc, ok := object.GetSymbolProperties()[key.Value.(*Symbol)]
		if !ok {
			return nil, false
		}
		return propertyDesc, true
	}

	if key.Type != TypeString {
		panic("Assert failed: GetPropertyFromObject key is not a string.")
	}

	propertyName := key.Value.(*String).Value
	propertyDesc, ok := object.GetProperties()[propertyName]
	if !ok {
		return nil, false
	}
	return propertyDesc, true
}

func SetPropertyToObject(object ObjectInterface, key *JavaScriptValue, descriptor PropertyDescriptor) {
	if key.Type == TypeSymbol {
		object.GetSymbolProperties()[key.Value.(*Symbol)] = descriptor
		return
	}

	if key.Type != TypeString {
		panic("Assert failed: SetPropertyToObject key is not a string.")
	}

	propertyName := key.Value.(*String).Value
	object.GetProperties()[propertyName] = descriptor
}

func DeletePropertyFromObject(object ObjectInterface, key *JavaScriptValue) {
	if key.Type == TypeSymbol {
		delete(object.GetSymbolProperties(), key.Value.(*Symbol))
		return
	}

	if key.Type != TypeString {
		panic("Assert failed: DeletePropertyFromObject key is not a string.")
	}

	propertyName := key.Value.(*String).Value
	delete(object.GetProperties(), propertyName)
}

type GeneratorState int

const (
	GeneratorStateSuspendedStart GeneratorState = iota
	GeneratorStateSuspendedYield
	GeneratorStateExecuting
	GeneratorStateCompleted
)

type Object struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool

	IsGenerator      bool
	GeneratorState   GeneratorState
	GeneratorContext *ExecutionContext
	GeneratorBrand   string

	// This corresponds to [[ErrorData]] in the spec.
	IsError bool
}

func NewEmptyObject() *Object {
	return &Object{
		Prototype:        nil,
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		Extensible:       true,
	}
}

func (o *Object) GetPrototype() ObjectInterface {
	return o.Prototype
}

func (o *Object) SetPrototype(prototype ObjectInterface) {
	o.Prototype = prototype
}

func (o *Object) GetProperties() map[string]PropertyDescriptor {
	return o.Properties
}

func (o *Object) SetProperties(properties map[string]PropertyDescriptor) {
	o.Properties = properties
}

func (o *Object) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	return o.SymbolProperties
}

func (o *Object) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	o.SymbolProperties = symbolProperties
}

func (o *Object) GetExtensible() bool {
	return o.Extensible
}

func (o *Object) SetExtensible(extensible bool) {
	o.Extensible = extensible
}

func (o *Object) GetPrototypeOf() *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *Object) SetPrototypeOf(prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(o, prototype)
}

func (o *Object) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(runtime, o, key)
}

func (o *Object) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryHasProperty(runtime, o, key)
}

func (o *Object) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	return OrdinaryDefineOwnProperty(runtime, o, key, descriptor)
}

func (o *Object) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(runtime, o, key, value, receiver)
}

func (o *Object) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(runtime, o, key, receiver)
}

func (o *Object) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	return OrdinaryDelete(runtime, o, key)
}

func (o *Object) OwnPropertyKeys() *Completion {
	return NewNormalCompletion(OrdinaryOwnPropertyKeys(o))
}

func (o *Object) PreventExtensions() *Completion {
	o.Extensible = false
	return NewNormalCompletion(NewBooleanValue(true))
}

func CopyDataProperties(
	runtime *Runtime,
	target ObjectInterface,
	source *JavaScriptValue,
	excludedItems []*JavaScriptValue,
) *Completion {
	if source.Type == TypeUndefined || source.Type == TypeNull {
		return NewUnusedCompletion()
	}

	fromObjCompletion := ToObject(runtime, source)
	if fromObjCompletion.Type != Normal {
		panic("Assert failed: CopyDataProperties ToObject threw an unexpected error.")
	}

	fromObjVal := fromObjCompletion.Value.(*JavaScriptValue)
	fromObj := fromObjVal.Value.(ObjectInterface)

	copyProperty := func(key *JavaScriptValue, value PropertyDescriptor) *Completion {
		excluded := false
		for _, excludedItem := range excludedItems {
			sameValCompletion := SameValue(key, excludedItem)
			if sameValCompletion.Type != Normal {
				panic("Assert failed: CopyDataProperties SameValue threw an unexpected error.")
			}
			if sameValCompletion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				excluded = true
				break
			}
		}

		if excluded {
			return NewUnusedCompletion()
		}

		if desc, ok := value.(*DataPropertyDescriptor); ok && desc != nil && desc.Enumerable {
			valueCompletion := fromObj.Get(runtime, key, fromObjVal)
			if valueCompletion.Type != Normal {
				return valueCompletion
			}

			value := valueCompletion.Value.(*JavaScriptValue)

			completion := CreateDataProperty(runtime, target, key, value)
			if completion.Type != Normal {
				panic("Assert failed: CreateDataProperty threw an unexpected error in CopyDataProperties.")
			}
		}

		return NewUnusedCompletion()
	}

	for key, value := range fromObj.GetProperties() {
		keyString := NewStringValue(key)
		completion := copyProperty(keyString, value)
		if completion.Type != Normal {
			return completion
		}
	}

	for key, value := range fromObj.GetSymbolProperties() {
		keyString := NewJavaScriptValue(TypeSymbol, key)
		completion := copyProperty(keyString, value)
		if completion.Type != Normal {
			return completion
		}
	}

	return NewUnusedCompletion()
}

type IntegrityLevel int

const (
	IntegrityLevelSealed IntegrityLevel = iota
	IntegrityLevelFrozen
)

func SetIntegrityLevel(runtime *Runtime, object ObjectInterface, integrityLevel IntegrityLevel) *Completion {
	completion := object.PreventExtensions()
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return completion
	}

	completion = object.OwnPropertyKeys()
	if completion.Type != Normal {
		return completion
	}

	keys := completion.Value.([]*JavaScriptValue)

	for _, key := range keys {
		completion = object.GetOwnProperty(runtime, key)
		if completion.Type != Normal {
			return completion
		}

		desc := completion.Value.(PropertyDescriptor)
		if dataDesc, ok := desc.(*DataPropertyDescriptor); ok && dataDesc != nil {
			dataDesc.Configurable = false
			if integrityLevel == IntegrityLevelFrozen {
				dataDesc.Writable = false
			}
		} else if accessorDesc, ok := desc.(*AccessorPropertyDescriptor); ok && accessorDesc != nil {
			accessorDesc.Configurable = false
		} else {
			panic("Assert failed: Descriptor must be a data or accessor property descriptor.")
		}

		completion = DefinePropertyOrThrow(runtime, object, key, desc)
		if completion.Type != Normal {
			return completion
		}
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func TestIntegrityLevel(runtime *Runtime, object ObjectInterface, integrityLevel IntegrityLevel) *Completion {
	if object.GetExtensible() {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	completion := object.OwnPropertyKeys()
	if completion.Type != Normal {
		return completion
	}

	keys := completion.Value.([]*JavaScriptValue)

	for _, key := range keys {
		completion = object.GetOwnProperty(runtime, key)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value == nil {
			continue
		}

		desc := completion.Value.(PropertyDescriptor)
		if desc.GetConfigurable() {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		if integrityLevel == IntegrityLevelFrozen {
			if dataDesc, ok := desc.(*DataPropertyDescriptor); ok && dataDesc != nil && dataDesc.Writable {
				return NewNormalCompletion(NewBooleanValue(false))
			}
		}
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

type GroupByKeyCoercion int

const (
	GroupByKeyCoercionProperty GroupByKeyCoercion = iota
	GroupByKeyCoercionCollection
)

type GroupByResult struct {
	GroupsByString map[string][]*JavaScriptValue
	GroupsBySymbol map[*Symbol][]*JavaScriptValue
}

func GroupBy(
	runtime *Runtime,
	items *JavaScriptValue,
	callback *JavaScriptValue,
	keyCoercion GroupByKeyCoercion,
) *Completion {
	completion := RequireObjectCoercible(runtime, items)
	if completion.Type != Normal {
		return completion
	}

	if callback.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Callback is not callable."))
	}

	callbackFunc, ok := callback.Value.(FunctionInterface)
	if !ok {
		return NewThrowCompletion(NewTypeError(runtime, "Callback is not a function."))
	}

	groupsByString := make(map[string][]*JavaScriptValue)
	groupsBySymbol := make(map[*Symbol][]*JavaScriptValue)

	completion = GetIterator(runtime, items, IteratorKindSync)
	if completion.Type != Normal {
		return completion
	}

	iterator := completion.Value.(*Iterator)

	k := uint64(0)

	for {
		if float64(k) >= math.Pow(2, 53)-1 {
			completion = NewThrowCompletion(NewTypeError(runtime, "Too many iterations in GroupBy."))
			return IteratorClose(runtime, iterator, completion)
		}

		completion = IteratorStepValue(runtime, iterator)
		if completion.Type != Normal {
			return completion
		}

		if stepResult, ok := completion.Value.(*IteratorStepResult); ok && stepResult.Done {
			return NewNormalCompletion(&GroupByResult{
				GroupsByString: groupsByString,
				GroupsBySymbol: groupsBySymbol,
			})
		}

		value, ok := completion.Value.(*JavaScriptValue)
		if !ok {
			panic("Assert failed: GroupBy received an invalid value when iterating.")
		}

		completion = callbackFunc.Call(
			runtime,
			NewUndefinedValue(),
			[]*JavaScriptValue{value, NewNumberValue(float64(k), false)},
		)
		IfAbruptCloseIterator(runtime, completion, iterator)

		key := completion
		keyVal, ok := key.Value.(*JavaScriptValue)

		if keyCoercion == GroupByKeyCoercionProperty {
			if ok {
				key = ToPropertyKey(keyVal)
				IfAbruptCloseIterator(runtime, key, iterator)
			}
		} else if ok {
			panic("TODO: GroupByKeyCoercionCollection is not implemented.")
		}

		if key.Type == Normal {
			keyVal = key.Value.(*JavaScriptValue)
			if stringVal, ok := keyVal.Value.(*String); ok {
				groupsByString[stringVal.Value] = append(groupsByString[stringVal.Value], value)
			} else if symbolVal, ok := keyVal.Value.(*Symbol); ok {
				groupsBySymbol[symbolVal] = append(groupsBySymbol[symbolVal], value)
			} else {
				panic("Assert failed: GroupBy received an invalid key when iterating.")
			}
		}

		k++
	}
}

type EnumerableOwnPropertiesKind int

const (
	EnumerableOwnPropertiesKindKey EnumerableOwnPropertiesKind = iota
	EnumerableOwnPropertiesKindValue
	EnumerableOwnPropertiesKindKeyAndValue
)

func EnumerableOwnProperties(runtime *Runtime, object ObjectInterface, kind EnumerableOwnPropertiesKind) *Completion {
	completion := object.OwnPropertyKeys()
	if completion.Type != Normal {
		return completion
	}

	keys := completion.Value.([]*JavaScriptValue)
	results := make([]*JavaScriptValue, 0)

	for _, key := range keys {
		if key.Type != TypeString {
			continue
		}

		completion = object.GetOwnProperty(runtime, key)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value == nil {
			continue
		}

		desc := completion.Value.(PropertyDescriptor)
		if !desc.GetEnumerable() {
			continue
		}

		if kind == EnumerableOwnPropertiesKindKey {
			results = append(results, key)
		} else {
			completion = object.Get(runtime, key, NewJavaScriptValue(TypeObject, object))
			if completion.Type != Normal {
				return completion
			}

			if kind == EnumerableOwnPropertiesKindValue {
				results = append(results, completion.Value.(*JavaScriptValue))
			} else {
				entry := CreateArrayFromList(runtime, []*JavaScriptValue{key, completion.Value.(*JavaScriptValue)})
				results = append(results, NewJavaScriptValue(TypeObject, entry))
			}
		}
	}

	return NewNormalCompletion(results)
}

func SetConstructor(runtime *Runtime, object ObjectInterface, constructor FunctionInterface) {
	object.DefineOwnProperty(runtime, NewStringValue("constructor"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, constructor),
		Writable:     false,
		Enumerable:   false,
		Configurable: false,
	})
}

func Invoke(
	runtime *Runtime,
	value *JavaScriptValue,
	propertyKey *JavaScriptValue,
	argumentList []*JavaScriptValue,
) *Completion {
	if argumentList == nil {
		argumentList = make([]*JavaScriptValue, 0)
	}

	completion := ToObject(runtime, value)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)
	object := objectVal.Value.(ObjectInterface)

	completion = object.Get(runtime, propertyKey, objectVal)
	if completion.Type != Normal {
		return completion
	}

	functionVal := completion.Value.(*JavaScriptValue)

	functionObj, ok := functionVal.Value.(FunctionInterface)
	if !ok {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot invoke a non-callable object."))
	}

	return functionObj.Call(runtime, value, argumentList)
}
