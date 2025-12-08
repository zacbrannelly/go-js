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
	GetSymbolProperties() map[*Symbol]PropertyDescriptor
	SetProperties(properties map[string]PropertyDescriptor)
	SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor)

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

type Object struct {
	Prototype        ObjectInterface
	Properties       map[string]PropertyDescriptor
	SymbolProperties map[*Symbol]PropertyDescriptor
	Extensible       bool
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

func (o *Object) GetOwnProperty(key *JavaScriptValue) *Completion {
	return OrdinaryGetOwnProperty(o, key)
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

func CopyDataProperties(
	target ObjectInterface,
	source *JavaScriptValue,
	excludedItems []*JavaScriptValue,
) *Completion {
	if source.Type == TypeUndefined || source.Type == TypeNull {
		return NewUnusedCompletion()
	}

	fromObjCompletion := ToObject(source)
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
			valueCompletion := fromObj.Get(key, fromObjVal)
			if valueCompletion.Type != Normal {
				return valueCompletion
			}

			value := valueCompletion.Value.(*JavaScriptValue)

			completion := CreateDataProperty(target, key, value)
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
