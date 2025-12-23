package runtime

type Runtime struct {
	ExecutionContextStack []*ExecutionContext

	// Well-known symbols.
	SymbolToStringTag *JavaScriptValue
	SymbolIterator    *JavaScriptValue
	SymbolSpecies     *JavaScriptValue
	SymbolUnscopables *JavaScriptValue
}

func NewRuntime() *Runtime {
	return &Runtime{
		ExecutionContextStack: []*ExecutionContext{},
		SymbolToStringTag:     NewSymbolValue("Symbol.toStringTag"),
		SymbolIterator:        NewSymbolValue("Symbol.iterator"),
		SymbolSpecies:         NewSymbolValue("Symbol.species"),
		SymbolUnscopables:     NewSymbolValue("Symbol.unscopables"),
		// TODO: Add other well-known symbols.
	}
}

func (r *Runtime) PushExecutionContext(executionContext *ExecutionContext) {
	r.ExecutionContextStack = append(r.ExecutionContextStack, executionContext)
}

func (r *Runtime) PopExecutionContext() *ExecutionContext {
	// Assert that the execution context stack is not empty.
	if len(r.ExecutionContextStack) == 0 {
		panic("Assert failed: Execution context stack is empty.")
	}

	executionContext := r.ExecutionContextStack[len(r.ExecutionContextStack)-1]
	r.ExecutionContextStack = r.ExecutionContextStack[:len(r.ExecutionContextStack)-1]
	return executionContext
}

func (r *Runtime) GetRunningExecutionContext() *ExecutionContext {
	if len(r.ExecutionContextStack) == 0 {
		panic("Assert failed: Execution context stack is empty.")
	}

	return r.ExecutionContextStack[len(r.ExecutionContextStack)-1]
}

func (r *Runtime) GetRunningRealm() *Realm {
	return r.GetRunningExecutionContext().Realm
}

func (r *Runtime) PushLabel(label string) {
	executionContext := r.GetRunningExecutionContext()
	executionContext.Labels = append(executionContext.Labels, label)
}

func (r *Runtime) PopLabel() string {
	executionContext := r.GetRunningExecutionContext()
	label := executionContext.Labels[len(executionContext.Labels)-1]
	executionContext.Labels = executionContext.Labels[:len(executionContext.Labels)-1]
	return label
}

func (r *Runtime) GetRunningLabels() []string {
	executionContext := r.GetRunningExecutionContext()
	return executionContext.Labels
}

func (r *Runtime) GetRunningScript() *Script {
	if len(r.ExecutionContextStack) == 0 {
		return nil
	}

	// Loop backwards from the top of the execution context stack to find the first script.
	for i := len(r.ExecutionContextStack) - 1; i >= 0; i-- {
		executionContext := r.ExecutionContextStack[i]
		if executionContext.Script != nil {
			return executionContext.Script
		}
	}

	return nil
}
