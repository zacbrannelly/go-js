package runtime

type Runtime struct {
	ExecutionContextStack []*ExecutionContext
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
