package runtime

type ObjectBinding struct {
}

type ObjectEnvironment struct {
	Bindings          map[string]ObjectBinding
	OuterEnv          Environment
	BindingObject     *Object
	IsWithEnvironment bool
}

func NewObjectEnvironment(bindingObject *Object, isWithEnvironment bool, outerEnv Environment) *ObjectEnvironment {
	return &ObjectEnvironment{
		Bindings:          make(map[string]ObjectBinding),
		OuterEnv:          outerEnv,
		BindingObject:     bindingObject,
		IsWithEnvironment: isWithEnvironment,
	}
}

func (e *ObjectEnvironment) GetOuterEnvironment() Environment {
	return e.OuterEnv
}

func (e *ObjectEnvironment) HasBinding(name string) bool {
	_, ok := e.Bindings[name]
	return ok
}

func (e *ObjectEnvironment) CreateMutableBinding(name string, deletable bool) *Completion {
	panic("not implemented")
}

func (e *ObjectEnvironment) CreateImmutableBinding(name string, strict bool) *Completion {
	panic("not implemented")
}

func (e *ObjectEnvironment) GetBindingValue(name string, strict bool) *Completion {
	panic("not implemented")
}

func (e *ObjectEnvironment) InitializeBinding(name string, value *JavaScriptValue) *Completion {
	panic("not implemented")
}
