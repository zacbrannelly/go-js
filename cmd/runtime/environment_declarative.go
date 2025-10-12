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

func (e *DeclarativeEnvironment) SetMutableBinding(name string, value *JavaScriptValue, strict bool) *Completion {
	binding, ok := e.Bindings[name]

	if !ok && strict {
		return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Cannot assign to an unresolvable reference '%s'", name)))
	}

	// Non-strict mode, create binding for unresolvable reference.
	if !ok {
		completion := e.CreateMutableBinding(name, true)
		if completion.Type == Throw {
			return completion
		}
		completion = e.InitializeBinding(name, value)
		if completion.Type == Throw {
			return completion
		}

		return NewUnusedCompletion()
	}

	if binding.Strict {
		strict = true
	}

	if binding.Value == nil {
		return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Referencing variable '%s' before its initialization", name)))
	}

	if binding.Mutable {
		binding.Value = value
	} else if strict {
		return NewThrowCompletion(NewTypeError(fmt.Sprintf("Cannot assign to a read only variable '%s'", name)))
	}

	return NewUnusedCompletion()
}
