package runtime

type ExecutionContext struct {
	Realm     *Realm
	Function  *FunctionObject
	Script    *Script
	Generator *Object
	// TODO: Store module record.

	// Points to the environments that can resolve identifier references.
	LexicalEnvironment  Environment
	VariableEnvironment Environment
	PrivateEnvironment  Environment

	// Labels.
	Labels []string

	// Execution state (Generator / Async).
	VM *ExecutionVM
}

func ResolveBindingFromCurrentContext(name string, runtime *Runtime, strict bool) *Completion {
	executionContext := runtime.GetRunningExecutionContext()
	env := executionContext.LexicalEnvironment
	return ResolveBinding(runtime, name, env, strict)
}
