package runtime

type Environment interface {
	GetOuterEnvironment() Environment
	HasBinding(runtime *Runtime, name string) bool
	CreateMutableBinding(runtime *Runtime, name string, value bool) *Completion
	CreateImmutableBinding(runtime *Runtime, name string, value bool) *Completion
	GetBindingValue(runtime *Runtime, name string, strict bool) *Completion
	InitializeBinding(runtime *Runtime, name string, value *JavaScriptValue) *Completion
	SetMutableBinding(runtime *Runtime, name string, value *JavaScriptValue, strict bool) *Completion
	DeleteBinding(runtime *Runtime, name string) *Completion
	WithBaseObject() *JavaScriptValue
	HasThisBinding() bool
	GetThisBinding(runtime *Runtime) *Completion
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

func ResolveBinding(runtime *Runtime, name string, environment Environment, strict bool) *Completion {
	return GetIdentifierReference(runtime, environment, name, strict)
}

func GetIdentifierReference(runtime *Runtime, env Environment, name string, strict bool) *Completion {
	if env == nil {
		// Unresolvable reference.
		return NewNormalCompletion(NewReferenceValueForEnvironment(nil, name, strict, nil))
	}

	exists := env.HasBinding(runtime, name)

	if exists {
		return NewNormalCompletion(NewReferenceValueForEnvironment(env, name, strict, nil))
	}

	// Recursively resolve the reference in the outer environments.
	return GetIdentifierReference(runtime, env.GetOuterEnvironment(), name, strict)
}
