package runtime

import "fmt"

type DeclarativeBinding struct {
	Mutable   bool
	Deletable bool
	Strict    bool
	Value     *JavaScriptValue
}

type DeclarativeEnvironment struct {
	Bindings map[string]*DeclarativeBinding
	OuterEnv Environment
}

func NewDeclarativeEnvironment(outerEnv Environment) *DeclarativeEnvironment {
	return &DeclarativeEnvironment{
		Bindings: make(map[string]*DeclarativeBinding),
		OuterEnv: outerEnv,
	}
}

func (e *DeclarativeEnvironment) GetOuterEnvironment() Environment {
	return e.OuterEnv
}

func (e *DeclarativeEnvironment) HasBinding(name string) bool {
	_, ok := e.Bindings[name]
	return ok
}

func (e *DeclarativeEnvironment) CreateMutableBinding(name string, deletable bool) *Completion {
	// Assert that name is not already bound in an environment record.
	if _, ok := e.Bindings[name]; ok {
		panic("Assert failed: CreateMutableBinding called with a name that is already bound in an environment record.")
	}

	e.Bindings[name] = &DeclarativeBinding{
		Mutable:   true,
		Deletable: deletable,
		Strict:    false,
		Value:     nil,
	}
	return NewUnusedCompletion()
}

func (e *DeclarativeEnvironment) CreateImmutableBinding(name string, strict bool) *Completion {
	// Assert that name is not already bound in an environment record.
	if _, ok := e.Bindings[name]; ok {
		panic("Assert failed: CreateMutableBinding called with a name that is already bound in an environment record.")
	}

	e.Bindings[name] = &DeclarativeBinding{
		Mutable:   false,
		Deletable: false,
		Strict:    strict,
		Value:     nil,
	}
	return NewUnusedCompletion()
}

func (e *DeclarativeEnvironment) GetBindingValue(name string, strict bool) *Completion {
	binding, ok := e.Bindings[name]
	if !ok {
		panic(fmt.Sprintf("Assert failed: GetBindingValue called with a name that is not bound: %s", name))
	}

	if binding.Value == nil {
		return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Cannot access '%s' before initialization", name)))
	}

	return NewNormalCompletion(binding.Value)
}

func (e *DeclarativeEnvironment) InitializeBinding(name string, value *JavaScriptValue) *Completion {
	binding, ok := e.Bindings[name]
	if !ok {
		panic(fmt.Sprintf("Assert failed: InitializeBinding called with a name that is not bound: %s", name))
	}

	binding.Value = value
	return NewUnusedCompletion()
}
