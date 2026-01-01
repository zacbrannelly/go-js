package runtime

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
	Get          FunctionInterface
	Set          FunctionInterface
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

func (d *AccessorPropertyDescriptor) GetGet() FunctionInterface {
	return d.Get
}

func (d *AccessorPropertyDescriptor) GetSet() FunctionInterface {
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

func IsCompatiblePropertyDescriptor(
	extensible bool,
	newDesc PropertyDescriptor,
	currentDesc PropertyDescriptor,
) *JavaScriptValue {
	return ValidateAndApplyPropertyDescriptor(
		NewUndefinedValue(),
		NewStringValue(""),
		extensible,
		newDesc,
		currentDesc,
	).Value.(*JavaScriptValue)
}

// TODO: Stop using Completion and just return the boolean value directly.
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
