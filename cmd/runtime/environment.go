package runtime

type Environment interface {
	GetOuterEnvironment() Environment
	HasBinding(name string) bool
	CreateMutableBinding(name string, value bool) *Completion
	CreateImmutableBinding(name string, value bool) *Completion
	GetBindingValue(name string, strict bool) *Completion
	InitializeBinding(name string, value *JavaScriptValue) *Completion
}
