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

func (r *Runtime) GetRunningExecutionContext() *ExecutionContext {
	if len(r.ExecutionContextStack) == 0 {
		panic("Assert failed: Execution context stack is empty.")
	}

	return r.ExecutionContextStack[len(r.ExecutionContextStack)-1]
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
