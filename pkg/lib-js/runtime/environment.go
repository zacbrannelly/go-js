package runtime

type Environment interface {
	GetOuterEnvironment() Environment
	HasBinding(name string) bool
	CreateMutableBinding(name string, value bool) *Completion
	CreateImmutableBinding(name string, value bool) *Completion
	GetBindingValue(runtime *Runtime, name string, strict bool) *Completion
	InitializeBinding(runtime *Runtime, name string, value *JavaScriptValue) *Completion
	SetMutableBinding(runtime *Runtime, name string, value *JavaScriptValue, strict bool) *Completion
	DeleteBinding(name string) *Completion
	WithBaseObject() *JavaScriptValue
	GetThisBinding() *Completion
}

func InitializeBoundName(
	runtime *Runtime,
	name string,
	value *JavaScriptValue,
	env Environment,
	isStrict bool,
) *Completion {
	if env != nil {
		completion := env.InitializeBinding(runtime, name, value)
		if completion.Type != Normal {
			panic("Assert failed: InitializeBinding threw an unexpected error in InitializeBoundName.")
		}
		return NewUnusedCompletion()
	}

	lhsCompletion := ResolveBindingFromCurrentContext(name, runtime, isStrict)
	if lhsCompletion.Type != Normal {
		return lhsCompletion
	}

	lhs := lhsCompletion.Value.(*JavaScriptValue)
	return PutValue(runtime, lhs, value)
}
