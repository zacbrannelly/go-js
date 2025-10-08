package runtime

type ObjectEnvironment struct {
	Bindings      map[string]any
	OuterEnv      Environment
	BindingObject *Object
}

func (e *ObjectEnvironment) HasBinding(name string) bool {
	_, ok := e.Bindings[name]
	return ok
}
