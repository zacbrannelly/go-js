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

type ObjectInterface interface {
	GetPrototype() ObjectInterface
	SetPrototype(prototype ObjectInterface)

	GetProperties() map[string]PropertyDescriptor
	SetProperties(properties map[string]PropertyDescriptor)

	GetExtensible() bool
	SetExtensible(extensible bool)

	// Internal methods
	GetPrototypeOf() *Completion
	SetPrototypeOf(prototype *JavaScriptValue) *Completion
	GetOwnProperty(key *JavaScriptValue) *Completion
	DefineOwnProperty(key *JavaScriptValue, descriptor PropertyDescriptor) *Completion
	HasProperty(key *JavaScriptValue) *Completion
	Get(key *JavaScriptValue, receiver *JavaScriptValue) *Completion
	Set(key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion
	Delete(key *JavaScriptValue) *Completion
	// TODO: OwnPropertyKeys() *Completion
}

type Object struct {
	Prototype  ObjectInterface
	Properties map[string]PropertyDescriptor
	Extensible bool
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

func (o *Object) GetExtensible() bool {
	return o.Extensible
}

func (o *Object) SetExtensible(extensible bool) {
	o.Extensible = extensible
}

func NewEmptyObject() *Object {
	return &Object{
		Prototype:  nil,
		Properties: make(map[string]PropertyDescriptor),
		Extensible: true,
	}
}

func (o *Object) GetPrototypeOf() *Completion {
	return OrdinaryGetPrototypeOf(o)
}

func (o *Object) SetPrototypeOf(prototype *JavaScriptValue) *Completion {
	return OrdinarySetPrototypeOf(o, prototype)
}

func (o *Object) GetOwnProperty(key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(o, key)
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
	return OrdinaryHasProperty(o, key)
}

func (o *Object) DefineOwnProperty(key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	return OrdinaryDefineOwnProperty(o, key, descriptor)
}

func (o *Object) Set(key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinarySet(o, key, value, receiver)
}

func (o *Object) Get(key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	return OrdinaryGet(o, key, receiver)
}

func (o *Object) Delete(key *JavaScriptValue) *Completion {
	return OrdinaryDelete(o, key)
}
