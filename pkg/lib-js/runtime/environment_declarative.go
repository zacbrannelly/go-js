package runtime

import "fmt"

type DeclarativeBinding struct {
	Mutable   bool
	Deletable bool
	Strict    bool
	Value     *JavaScriptValue
}

type ThisBindingStatus int

const (
	ThisBindingStatusUninitialized ThisBindingStatus = iota
	ThisBindingStatusInitialized
	ThisBindingStatusLexical
)

type DeclarativeEnvironment struct {
	Bindings map[string]*DeclarativeBinding
	OuterEnv Environment

	// Extra Function Environment fields.
	IsFunctionEnvironment bool
	ThisValue             *JavaScriptValue
	ThisBindingStatus     ThisBindingStatus
	FunctionObject        *FunctionObject
	NewTarget             *JavaScriptValue
}

func NewDeclarativeEnvironment(outerEnv Environment) *DeclarativeEnvironment {
	return &DeclarativeEnvironment{
		Bindings:              make(map[string]*DeclarativeBinding),
		OuterEnv:              outerEnv,
		IsFunctionEnvironment: false,
	}
}

func (e *DeclarativeEnvironment) GetOuterEnvironment() Environment {
	return e.OuterEnv
}

func (e *DeclarativeEnvironment) HasBinding(runtime *Runtime, name string) bool {
	_, ok := e.Bindings[name]
	return ok
}

func (e *DeclarativeEnvironment) CreateMutableBinding(runtime *Runtime, name string, deletable bool) *Completion {
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

func (e *DeclarativeEnvironment) CreateImmutableBinding(runtime *Runtime, name string, strict bool) *Completion {
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

func (e *DeclarativeEnvironment) GetBindingValue(runtime *Runtime, name string, strict bool) *Completion {
	binding, ok := e.Bindings[name]
	if !ok {
		panic(fmt.Sprintf("Assert failed: GetBindingValue called with a name that is not bound: %s", name))
	}

	if binding.Value == nil {
		return NewThrowCompletion(NewReferenceError(runtime, fmt.Sprintf("Cannot access '%s' before initialization", name)))
	}

	return NewNormalCompletion(binding.Value)
}

func (e *DeclarativeEnvironment) InitializeBinding(runtime *Runtime, name string, value *JavaScriptValue) *Completion {
	binding, ok := e.Bindings[name]
	if !ok {
		panic(fmt.Sprintf("Assert failed: InitializeBinding called with a name that is not bound: %s", name))
	}

	binding.Value = value
	return NewUnusedCompletion()
}

func (e *DeclarativeEnvironment) SetMutableBinding(runtime *Runtime, name string, value *JavaScriptValue, strict bool) *Completion {
	binding, ok := e.Bindings[name]

	if !ok && strict {
		return NewThrowCompletion(NewReferenceError(runtime, fmt.Sprintf("Cannot assign to an unresolvable reference '%s'", name)))
	}

	// Non-strict mode, create binding for unresolvable reference.
	if !ok {
		completion := e.CreateMutableBinding(runtime, name, true)
		if completion.Type != Normal {
			return completion
		}
		completion = e.InitializeBinding(runtime, name, value)
		if completion.Type != Normal {
			return completion
		}

		return NewUnusedCompletion()
	}

	if binding.Strict {
		strict = true
	}

	if binding.Value == nil {
		return NewThrowCompletion(NewReferenceError(runtime, fmt.Sprintf("Referencing variable '%s' before its initialization", name)))
	}

	if binding.Mutable {
		binding.Value = value
	} else if strict {
		return NewThrowCompletion(NewTypeError(runtime, fmt.Sprintf("Cannot assign to a read only variable '%s'", name)))
	}

	return NewUnusedCompletion()
}

func (e *DeclarativeEnvironment) DeleteBinding(runtime *Runtime, name string) *Completion {
	binding, ok := e.Bindings[name]
	if !ok {
		panic(fmt.Sprintf("Assert failed: DeleteBinding called with a name '%s' that is not bound", name))
	}

	if !binding.Deletable {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	delete(e.Bindings, name)
	return NewNormalCompletion(NewBooleanValue(true))
}

func (e *DeclarativeEnvironment) WithBaseObject() *JavaScriptValue {
	return NewUndefinedValue()
}
