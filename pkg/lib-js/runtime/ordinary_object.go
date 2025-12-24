package runtime

import (
	"fmt"
	"sort"
	"strconv"
)

func OrdinaryObjectCreate(proto ObjectInterface) ObjectInterface {
	object := &Object{
		Prototype:        proto,
		Properties:       make(map[string]PropertyDescriptor),
		SymbolProperties: make(map[*Symbol]PropertyDescriptor),
		Extensible:       true,
	}

	return object
}

func OrdinaryGetPrototypeOf(object ObjectInterface) *Completion {
	prototype := object.GetPrototype()
	if prototype == nil {
		return NewNormalCompletion(NewNullValue())
	}

	return NewNormalCompletion(NewJavaScriptValue(TypeObject, prototype))
}

func OrdinarySetPrototypeOf(object ObjectInterface, prototype *JavaScriptValue) *Completion {
	var currentVal *JavaScriptValue = nil
	if currentProto := object.GetPrototype(); currentProto != nil {
		currentVal = NewJavaScriptValue(TypeObject, currentProto)
	} else {
		currentVal = NewNullValue()
	}

	sameValCompletion := SameValue(
		currentVal,
		prototype,
	)
	if sameValCompletion.Type != Normal {
		return sameValCompletion
	}

	if sameValCompletion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	if !object.GetExtensible() {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	objectVal := NewJavaScriptValue(TypeObject, object)

	p := prototype
	for {
		if p.Type == TypeNull {
			break
		}

		sameValCompletion := SameValue(
			p,
			objectVal,
		)
		if sameValCompletion.Type != Normal {
			return sameValCompletion
		}

		if sameValCompletion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		pObj := p.Value.(ObjectInterface)

		// If the prototype is not an ordinary object, break the loop.
		if !HasOrdinaryGetPrototypeOf(pObj) {
			break
		}

		maybeObj := pObj.GetPrototype()
		if maybeObj == nil {
			break
		}

		p = NewJavaScriptValue(TypeObject, maybeObj)
	}

	if p.Type == TypeNull {
		object.SetPrototype(nil)
	} else {
		object.SetPrototype(p.Value.(ObjectInterface))
	}
	return NewNormalCompletion(NewBooleanValue(true))
}

func HasOrdinaryGetPrototypeOf(object ObjectInterface) bool {
	if _, ok := object.GetPrototype().(*Object); ok {
		return true
	}

	if _, ok := object.GetPrototype().(*ObjectPrototype); ok {
		return true
	}

	if _, ok := object.GetPrototype().(*ArrayObject); ok {
		return true
	}

	if _, ok := object.GetPrototype().(*FunctionObject); ok {
		return true
	}

	return false
}

func OrdinaryGetOwnProperty(runtime *Runtime, object ObjectInterface, key *JavaScriptValue) *Completion {
	if key.Type != TypeString && key.Type != TypeSymbol {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid key type"))
	}

	propertyDesc, ok := GetPropertyFromObject(object, key)
	if !ok {
		// Nil to signal undefined.
		return NewNormalCompletion(nil)
	}
	return NewNormalCompletion(propertyDesc.Copy())
}

func OrdinaryHasProperty(runtime *Runtime, object ObjectInterface, key *JavaScriptValue) *Completion {
	ownPropertyCompletion := object.GetOwnProperty(runtime, key)
	if ownPropertyCompletion.Type != Normal {
		return ownPropertyCompletion
	}

	if val, ok := ownPropertyCompletion.Value.(PropertyDescriptor); ok && val != nil {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	prototypeCompletion := object.GetPrototypeOf()
	if prototypeCompletion.Type != Normal {
		return prototypeCompletion
	}

	if prototypeVal, ok := prototypeCompletion.Value.(*JavaScriptValue).Value.(ObjectInterface); ok && prototypeVal != nil {
		return prototypeVal.HasProperty(runtime, key)
	}

	return NewNormalCompletion(NewBooleanValue(false))
}

func OrdinaryDefineOwnProperty(runtime *Runtime, object ObjectInterface, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	currentCompletion := object.GetOwnProperty(runtime, key)
	if currentCompletion.Type != Normal {
		return currentCompletion
	}

	var currentDescriptor PropertyDescriptor = nil
	if val, ok := currentCompletion.Value.(PropertyDescriptor); ok && val != nil {
		currentDescriptor = val
	}

	return ValidateAndApplyPropertyDescriptor(
		NewJavaScriptValue(TypeObject, object),
		key,
		object.GetExtensible(),
		descriptor,
		currentDescriptor,
	)
}

func ValidateAndApplyPropertyDescriptor(
	object *JavaScriptValue,
	key *JavaScriptValue,
	extensible bool,
	descriptor PropertyDescriptor,
	currentDescriptor PropertyDescriptor,
) *Completion {
	if currentDescriptor == nil {
		if !extensible {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		if object.Type == TypeUndefined {
			return NewNormalCompletion(NewBooleanValue(true))
		}

		if objectVal, ok := object.Value.(ObjectInterface); ok && objectVal != nil {
			SetPropertyToObject(objectVal, key, descriptor)
			return NewNormalCompletion(NewBooleanValue(true))
		}
		panic("Assert failed: Object is not an object.")
	}

	if !currentDescriptor.GetConfigurable() {
		if descriptor.GetConfigurable() {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		if descriptor.GetEnumerable() != currentDescriptor.GetEnumerable() {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		if descriptor.GetType() != currentDescriptor.GetType() {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		if currentDescriptor.GetType() == AccessorPropertyDescriptorType {
			panic("TODO: Support setting accessor property descriptors.")
		} else if !currentDescriptor.(*DataPropertyDescriptor).Writable {
			if descriptor.(*DataPropertyDescriptor).Writable {
				return NewNormalCompletion(NewBooleanValue(false))
			}

			// TODO: Return true if the value in the existing and new descriptor are the same using SameValue function.
			panic("TODO: Implement SameValue function.")
		}
	}

	if object.Type == TypeUndefined {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	objectVal, ok := object.Value.(ObjectInterface)
	if !ok || objectVal == nil {
		panic("Assert failed: Object is not an object.")
	}

	// TODO: Merge the existing descriptor with the new descriptor based on which fields are set in the new descriptor.
	SetPropertyToObject(objectVal, key, descriptor)
	return NewNormalCompletion(NewBooleanValue(true))
}

func OrdinarySet(runtime *Runtime, object ObjectInterface, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	ownDescriptor := object.GetOwnProperty(runtime, key)
	if ownDescriptor.Type != Normal {
		return ownDescriptor
	}

	var ownDescriptorVal PropertyDescriptor
	if ownDescriptor.Value != nil {
		ownDescriptorVal = ownDescriptor.Value.(PropertyDescriptor)
	}

	// property descriptor is undefined.
	if ownDescriptorVal == nil {
		parent := object.GetPrototypeOf()
		if parent.Type != Normal {
			return parent
		}

		parentVal := parent.Value

		// NOTE: Nil checks from `any` types require a type assertion check, otherwise it will be a false positive.
		if parentObj, ok := parentVal.(*JavaScriptValue).Value.(ObjectInterface); ok && parentObj != nil {
			return parentObj.Set(runtime, key, value, receiver)
		}

		ownDescriptorVal = &DataPropertyDescriptor{
			Value:        nil,
			Writable:     true,
			Enumerable:   true,
			Configurable: true,
		}
	}

	if ownDescriptorVal.GetType() == DataPropertyDescriptorType {
		dataDescriptor := ownDescriptorVal.(*DataPropertyDescriptor)
		if !dataDescriptor.Writable {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		if receiver.Type != TypeObject {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		receiverObj := receiver.Value.(ObjectInterface)
		if receiverObj == nil {
			panic("Assert failed: Receiver is nil when it should be an object.")
		}

		existingDescCompletion := receiverObj.GetOwnProperty(runtime, key)
		if existingDescCompletion.Type != Normal {
			return existingDescCompletion
		}

		if existingDesc, ok := existingDescCompletion.Value.(PropertyDescriptor); ok && existingDesc != nil {
			if existingDesc.GetType() == AccessorPropertyDescriptorType {
				return NewNormalCompletion(NewBooleanValue(false))
			}

			dataDesc := existingDesc.(*DataPropertyDescriptor)
			if !dataDesc.Writable {
				return NewNormalCompletion(NewBooleanValue(false))
			}

			// TODO: This deviates from the spec, should just be the new value, all other fields are unset.
			// TODO: Then the ValidateAndApplyPropertyDescriptor function should merge the new value with the existing descriptor.
			// TODO: To support this, we need to modify PropertyDescriptor to keep track of which fields are set.
			// TODO: Potentially just make all the fields JavaScriptValue types, then `nil` can signal unset.
			valueDesc := &DataPropertyDescriptor{
				Writable:     dataDesc.Writable,
				Enumerable:   dataDesc.Enumerable,
				Configurable: dataDesc.Configurable,
				Value:        value,
			}
			return receiverObj.DefineOwnProperty(runtime, key, valueDesc)
		} else {
			return CreateDataProperty(runtime, receiverObj, key, value)
		}
	}

	if ownDescriptorVal.GetType() != AccessorPropertyDescriptorType {
		panic("Assert failed: Descriptor must be a data or accessor property descriptor.")
	}

	setter := ownDescriptorVal.(*AccessorPropertyDescriptor).GetSet()
	if setter == nil {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	completion := setter.Call(runtime, receiver, []*JavaScriptValue{value})
	if completion.Type != Normal {
		return completion
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func OrdinaryGet(runtime *Runtime, object ObjectInterface, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	ownDescriptorCompletion := object.GetOwnProperty(runtime, key)
	if ownDescriptorCompletion.Type != Normal {
		return ownDescriptorCompletion
	}

	if ownDescriptor, _ := ownDescriptorCompletion.Value.(PropertyDescriptor); ownDescriptor == nil {
		parent := object.GetPrototypeOf()
		if parent.Type != Normal {
			return parent
		}

		parentVal := parent.Value
		if parentObj, ok := parentVal.(*JavaScriptValue).Value.(ObjectInterface); ok && parentObj != nil {
			return parentObj.Get(runtime, key, receiver)
		}

		return NewNormalCompletion(NewUndefinedValue())
	}

	ownDescriptor := ownDescriptorCompletion.Value.(PropertyDescriptor)
	if dataDescriptor, ok := ownDescriptor.(*DataPropertyDescriptor); ok {
		return NewNormalCompletion(dataDescriptor.Value)
	}

	if accessorDescriptor, ok := ownDescriptor.(*AccessorPropertyDescriptor); ok {
		return accessorDescriptor.Get.Call(runtime, receiver, []*JavaScriptValue{})
	}

	panic("Assert failed: Descriptor must be a data or accessor property descriptor.")
}

func OrdinaryDelete(runtime *Runtime, object ObjectInterface, key *JavaScriptValue) *Completion {
	descCompletion := object.GetOwnProperty(runtime, key)
	if descCompletion.Type != Normal {
		return descCompletion
	}

	if descCompletion.Value == nil {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	desc := descCompletion.Value.(PropertyDescriptor)
	if !desc.GetConfigurable() {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	DeletePropertyFromObject(object, key)
	return NewNormalCompletion(NewBooleanValue(true))
}

func CreateDataProperty(runtime *Runtime, object ObjectInterface, key *JavaScriptValue, value *JavaScriptValue) *Completion {
	return object.DefineOwnProperty(runtime, key, &DataPropertyDescriptor{
		Value:        value,
		Writable:     true,
		Enumerable:   true,
		Configurable: true,
	})
}

func DefinePropertyOrThrow(runtime *Runtime, object ObjectInterface, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	completion := object.DefineOwnProperty(runtime, key, descriptor)
	if completion.Type != Normal {
		return completion
	}

	if success, ok := completion.Value.(*Boolean); ok && !success.Value {
		keyString := PropertyKeyToString(key)
		return NewThrowCompletion(NewTypeError(runtime, fmt.Sprintf("Cannot define property '%s', object is not extensible", keyString)))
	}

	return NewUnusedCompletion()
}

func HasOwnProperty(runtime *Runtime, object ObjectInterface, key *JavaScriptValue) *Completion {
	ownProperty := object.GetOwnProperty(runtime, key)
	if ownProperty.Type != Normal {
		return ownProperty
	}

	return NewNormalCompletion(NewBooleanValue(ownProperty.Value != nil))
}

func OrdinaryOwnPropertyKeys(object ObjectInterface) []*JavaScriptValue {
	keys := make([]*JavaScriptValue, 0)
	arrayKeys := make([]int, 0)

	seen := make(map[string]bool)

	for key, _ := range object.GetProperties() {
		arrayKey, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			continue
		}

		keys = append(keys, NewStringValue(key))
		arrayKeys = append(arrayKeys, int(arrayKey))

		seen[key] = true
	}

	// Sort the keys in ascending order.
	sort.Slice(keys, func(i, j int) bool {
		return arrayKeys[i] < arrayKeys[j]
	})

	// TODO: This needs to be insertion order.
	for key, _ := range object.GetProperties() {
		if seen[key] {
			continue
		}

		keys = append(keys, NewStringValue(key))
	}

	// TODO: This needs to be insertion order.
	for key, _ := range object.GetSymbolProperties() {
		keys = append(keys, NewJavaScriptValue(TypeSymbol, key))
	}

	return keys
}

func OrdinaryCreateFromConstructor(
	runtime *Runtime,
	constructor *FunctionObject,
	defaultProto Intrinsic,
) *Completion {
	completion := GetPrototypeFromConstructor(runtime, constructor, defaultProto)
	if completion.Type != Normal {
		return completion
	}

	prototype := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, OrdinaryObjectCreate(prototype)))
}

func GetPrototypeFromConstructor(
	runtime *Runtime,
	constructor *FunctionObject,
	defaultProto Intrinsic,
) *Completion {
	completion := constructor.Get(
		runtime,
		NewStringValue("prototype"),
		NewJavaScriptValue(TypeObject, defaultProto),
	)
	if completion.Type != Normal {
		return completion
	}

	if prototype, ok := completion.Value.(*JavaScriptValue).Value.(ObjectInterface); ok && prototype != nil {
		return completion
	}

	completion = GetFunctionRealm(runtime, constructor)
	if completion.Type != Normal {
		return completion
	}

	realm := completion.Value.(*Realm)
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, realm.GetIntrinsic(defaultProto)))
}

func OrdinaryHasInstance(runtime *Runtime, constructorVal *JavaScriptValue, objectVal *JavaScriptValue) *Completion {
	if constructorVal.Type != TypeObject {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	constructorFuncObj, ok := constructorVal.Value.(*FunctionObject)
	if !ok {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	// TODO: Check [[BoundTargetFunction]] when supported.

	if objectVal.Type != TypeObject {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	completion := constructorFuncObj.Get(runtime, NewStringValue("prototype"), objectVal)
	if completion.Type != Normal {
		return completion
	}

	prototypeVal := completion.Value.(*JavaScriptValue)

	if prototypeVal.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Prototype value is not an object."))
	}

	for {
		completion = objectVal.Value.(ObjectInterface).GetPrototypeOf()
		if completion.Type != Normal {
			return completion
		}

		objectVal = completion.Value.(*JavaScriptValue)

		if objectVal.Type == TypeNull {
			return NewNormalCompletion(NewBooleanValue(false))
		}

		completion = SameValue(prototypeVal, objectVal)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewNormalCompletion(NewBooleanValue(true))
		}
	}
}
