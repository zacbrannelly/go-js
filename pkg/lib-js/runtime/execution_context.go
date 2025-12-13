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
