package runtime

type Environment interface {
	GetOuterEnvironment() Environment
	HasBinding(name string) bool
	CreateMutableBinding(name string, value bool) *Completion
	CreateImmutableBinding(name string, value bool) *Completion
	GetBindingValue(name string, strict bool) *Completion
	InitializeBinding(name string, value *JavaScriptValue) *Completion
	SetMutableBinding(name string, value *JavaScriptValue, strict bool) *Completion
	DeleteBinding(name string) *Completion
	WithBaseObject() *JavaScriptValue
}

func InitializeBoundName(
	runtime *Runtime,
	name string,
	value *JavaScriptValue,
	env Environment,
	isStrict bool,
) *Completion {
	if env != nil {
		completion := env.InitializeBinding(name, value)
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
