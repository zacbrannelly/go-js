package runtime

type ExecutionContext struct {
	Realm    *Realm
	Function *FunctionObject
	Script   *Script
	// TODO: Store module record.

	// Points to the environments that can resolve identifier references.
	LexicalEnvironment  Environment
	VariableEnvironment Environment
	PrivateEnvironment  Environment

	// Labels.
	Labels []string
}

func ResolveBindingFromCurrentContext(name string, runtime *Runtime, strict bool) *Completion {
	executionContext := runtime.GetRunningExecutionContext()
	env := executionContext.LexicalEnvironment
	return ResolveBinding(name, env, strict)
}

func ResolveBinding(name string, environment Environment, strict bool) *Completion {
	return GetIdentifierReference(environment, name, strict)
}

func GetIdentifierReference(env Environment, name string, strict bool) *Completion {
	if env == nil {
		// Unresolvable reference.
		return NewNormalCompletion(NewReferenceValueForEnvironment(nil, name, strict, nil))
	}

	exists := env.HasBinding(name)

	if exists {
		return NewNormalCompletion(NewReferenceValueForEnvironment(env, name, strict, nil))
	}

	// Recursively resolve the reference in the outer environments.
	return GetIdentifierReference(env.GetOuterEnvironment(), name, strict)
}
