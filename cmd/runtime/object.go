package runtime

import "fmt"

type PropertyDescriptorType int

const (
	DataPropertyDescriptorType PropertyDescriptorType = iota
	AccessorPropertyDescriptorType
)

type PropertyDescriptor interface {
	GetType() PropertyDescriptorType
	GetEnumerable() bool
	GetConfigurable() bool
	Copy() PropertyDescriptor
}

type DataPropertyDescriptor struct {
	Value        *JavaScriptValue
	Writable     bool
	Enumerable   bool
	Configurable bool
}

func (d *DataPropertyDescriptor) GetType() PropertyDescriptorType {
	return DataPropertyDescriptorType
}

func (d *DataPropertyDescriptor) GetEnumerable() bool {
	return d.Enumerable
}

func (d *DataPropertyDescriptor) GetConfigurable() bool {
	return d.Configurable
}

func (d *DataPropertyDescriptor) GetValue() any {
	return d.Value
}

func (d *DataPropertyDescriptor) GetWritable() bool {
	return d.Writable
}

func (d *DataPropertyDescriptor) Copy() PropertyDescriptor {
	return &DataPropertyDescriptor{
		Value:        d.Value,
		Writable:     d.Writable,
		Enumerable:   d.Enumerable,
		Configurable: d.Configurable,
	}
}

type AccessorPropertyDescriptor struct {
	Get          *JavaScriptValue
	Set          *JavaScriptValue
	Enumerable   bool
	Configurable bool
}

func (d *AccessorPropertyDescriptor) GetType() PropertyDescriptorType {
	return AccessorPropertyDescriptorType
}

func (d *AccessorPropertyDescriptor) GetEnumerable() bool {
	return d.Enumerable
}

func (d *AccessorPropertyDescriptor) GetConfigurable() bool {
	return d.Configurable
}

func (d *AccessorPropertyDescriptor) GetGet() any {
	return d.Get
}

func (d *AccessorPropertyDescriptor) GetSet() any {
	return d.Set
}

func (d *AccessorPropertyDescriptor) Copy() PropertyDescriptor {
	return &AccessorPropertyDescriptor{
		Get:          d.Get,
		Set:          d.Set,
		Enumerable:   d.Enumerable,
		Configurable: d.Configurable,
	}
}

type Object struct {
	Prototype  *Object
	Properties map[string]PropertyDescriptor
	Extensible bool
}

func NewEmptyObject() *Object {
	return &Object{
		Prototype:  nil,
		Properties: make(map[string]PropertyDescriptor),
		Extensible: true,
	}
}

func (o *Object) GetPrototypeOf() *Completion {
	return NewNormalCompletion(o.Prototype)
}

func (o *Object) HasOwnProperty(key *JavaScriptValue) *Completion {
	ownProperty := o.GetOwnProperty(key)
	if ownProperty.Type == Throw {
		return ownProperty
	}

	return NewNormalCompletion(NewBooleanValue(ownProperty.Value != nil))
}

func (o *Object) GetOwnProperty(key *JavaScriptValue) *Completion {
	if key.Type != TypeString && key.Type != TypeSymbol {
		return NewThrowCompletion(NewTypeError("Invalid key type"))
	}

	var propertyName string
	switch key.Type {
	case TypeString:
		propertyName = key.Value.(*String).Value
	case TypeSymbol:
		propertyName = key.Value.(*Symbol).Name
	}

	propertyDesc, ok := o.Properties[propertyName]
	if !ok {
		// Nil to signal undefined.
		return NewNormalCompletion(nil)
	}

	// Return a copy of the property descriptor.
	return NewNormalCompletion(propertyDesc.Copy())
}

func (o *Object) GetOwnPropertyViaString(key string) PropertyDescriptor {
	propertyDesc, ok := o.Properties[key]
	if !ok {
		return nil
	}

	// Return a copy of the property descriptor.
	return propertyDesc.Copy()
}

func (o *Object) HasProperty(key *JavaScriptValue) *Completion {
	ownPropertyCompletion := o.GetOwnProperty(key)
	if ownPropertyCompletion.Type == Throw {
		return ownPropertyCompletion
	}

	if val, ok := ownPropertyCompletion.Value.(PropertyDescriptor); ok && val != nil {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	prototypeCompletion := o.GetPrototypeOf()
	if prototypeCompletion.Type == Throw {
		return prototypeCompletion
	}

	if prototypeVal, ok := prototypeCompletion.Value.(*Object); ok && prototypeVal != nil {
		return prototypeVal.HasProperty(key)
	}

	return NewNormalCompletion(NewBooleanValue(false))
}

func (o *Object) DefineOwnProperty(key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	currentCompletion := o.GetOwnProperty(key)
	if currentCompletion.Type == Throw {
		return currentCompletion
	}

	var currentDescriptor PropertyDescriptor = nil
	if val, ok := currentCompletion.Value.(PropertyDescriptor); ok && val != nil {
		currentDescriptor = val
	}

	return ValidateAndApplyPropertyDescriptor(
		NewJavaScriptValue(TypeObject, o),
		key,
		o.Extensible,
		descriptor,
		currentDescriptor,
	)
}

func (o *Object) Set(key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	ownDescriptor := o.GetOwnProperty(key)
	if ownDescriptor.Type == Throw {
		return ownDescriptor
	}

	var ownDescriptorVal PropertyDescriptor
	if ownDescriptor.Value != nil {
		ownDescriptorVal = ownDescriptor.Value.(PropertyDescriptor)
	}

	// property descriptor is undefined.
	if ownDescriptorVal == nil {
		parent := o.GetPrototypeOf()
		if parent.Type == Throw {
			return parent
		}

		parentVal := parent.Value

		// NOTE: Nil checks from `any` types require a type assertion check, otherwise it will be a false positive.
		if parentObj, ok := parentVal.(*Object); ok && parentObj != nil {
			return parentObj.Set(key, value, receiver)
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

		receiverObj := receiver.Value.(*Object)
		if receiverObj == nil {
			panic("Assert failed: Receiver is nil when it should be an object.")
		}

		existingDescCompletion := receiverObj.GetOwnProperty(key)
		if existingDescCompletion.Type == Throw {
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
			return receiverObj.DefineOwnProperty(key, valueDesc)
		} else {
			return receiverObj.CreateDataProperty(key, value)
		}
	}

	if ownDescriptorVal.GetType() != AccessorPropertyDescriptorType {
		panic("Assert failed: Descriptor must be a data or accessor property descriptor.")
	}

	// setter := ownDescriptorVal.(*AccessorPropertyDescriptor).GetSet()
	// if setter == nil {
	// 	return NewNormalCompletion(NewBooleanValue(false))
	// }

	panic("TODO: Support setting accessor property descriptors.")
}

func (o *Object) Get(key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	ownDescriptorCompletion := o.GetOwnProperty(key)
	if ownDescriptorCompletion.Type == Throw {
		return ownDescriptorCompletion
	}

	if ownDescriptor, ok := ownDescriptorCompletion.Value.(PropertyDescriptor); ok && ownDescriptor == nil {
		parent := o.GetPrototypeOf()
		if parent.Type == Throw {
			return parent
		}

		parentVal := parent.Value
		if parentObj, ok := parentVal.(*Object); ok && parentObj != nil {
			return parentObj.Get(key, receiver)
		}

		return NewNormalCompletion(NewUndefinedValue())
	}

	ownDescriptor := ownDescriptorCompletion.Value.(PropertyDescriptor)
	if ownDescriptor.GetType() == DataPropertyDescriptorType {
		dataDescriptor := ownDescriptor.(*DataPropertyDescriptor)
		return NewNormalCompletion(dataDescriptor.Value)
	}

	panic("TODO: Support accessor property descriptors.")
}

func (o *Object) CreateDataProperty(key *JavaScriptValue, value *JavaScriptValue) *Completion {
	return o.DefineOwnProperty(key, &DataPropertyDescriptor{
		Value:        value,
		Writable:     true,
		Enumerable:   true,
		Configurable: true,
	})
}

func (o *Object) DefinePropertyOrThrow(key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	completion := o.DefineOwnProperty(key, descriptor)
	if completion.Type == Throw {
		return completion
	}

	if success, ok := completion.Value.(*Boolean); ok && !success.Value {
		keyString := PropertyKeyToString(key)
		return NewThrowCompletion(NewTypeError(fmt.Sprintf("Cannot define property '%s', object is not extensible", keyString)))
	}

	return NewUnusedCompletion()
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

		if objectVal, ok := object.Value.(*Object); ok && objectVal != nil {
			objectVal.Properties[PropertyKeyToString(key)] = descriptor
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

	objectVal, ok := object.Value.(*Object)
	if !ok {
		panic("Assert failed: Object is not an object.")
	}

	// TODO: Merge the existing descriptor with the new descriptor based on which fields are set in the new descriptor.
	objectVal.Properties[PropertyKeyToString(key)] = descriptor
	return NewNormalCompletion(NewBooleanValue(true))
}
