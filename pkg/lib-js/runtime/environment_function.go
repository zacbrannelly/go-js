package runtime

func NewFunctionEnvironment(function *FunctionObject, newTarget *JavaScriptValue) *DeclarativeEnvironment {
	functionEnv := &DeclarativeEnvironment{
		Bindings:              make(map[string]*DeclarativeBinding),
		OuterEnv:              function.Environment,
		IsFunctionEnvironment: true,
		FunctionObject:        function,
		NewTarget:             newTarget,
	}

	if function.ThisMode == ThisModeLexical {
		functionEnv.ThisBindingStatus = ThisBindingStatusLexical
	} else {
		functionEnv.ThisBindingStatus = ThisBindingStatusUninitialized
	}

	return functionEnv
}

func (e *DeclarativeEnvironment) HasThisBinding() bool {
	return e.IsFunctionEnvironment
}

func (e *DeclarativeEnvironment) GetThisBinding() *Completion {
	if !e.IsFunctionEnvironment {
		panic("Assert failed: GetThisBinding called on a non-function environment.")
	}

	if e.ThisBindingStatus == ThisBindingStatusUninitialized {
		return NewThrowCompletion(NewReferenceError("Cannot access 'this' before initialization"))
	}

	return NewNormalCompletion(e.ThisValue)
}
