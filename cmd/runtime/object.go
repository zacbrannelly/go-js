package runtime

type PropertyDescriptor interface {
	GetEnumerable() bool
	GetConfigurable() bool
	Copy() PropertyDescriptor
}

type DataPropertyDescriptor struct {
	Value        any
	Writable     bool
	Enumerable   bool
	Configurable bool
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
	Get          any
	Set          any
	Enumerable   bool
	Configurable bool
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
}

func NewEmptyObject() *Object {
	return &Object{
		Prototype:  nil,
		Properties: make(map[string]PropertyDescriptor),
	}
}

func (o *Object) GetPrototypeOf() *Completion {
	return NewNormalCompletion(o.Prototype)
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
