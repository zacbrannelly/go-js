package runtime

import "fmt"

type Environment interface {
	HasBinding(name string) bool
	CreateMutableBinding(name string, value bool) *Completion
	CreateImmutableBinding(name string, value bool) *Completion
}

type DeclarativeEnvironment struct {
	Bindings          map[string]bool
	MutableBindings   map[string]bool
	DeletableBindings map[string]bool
	StrictBindings    map[string]bool
	OuterEnv          Environment
}

func (e *DeclarativeEnvironment) HasBinding(name string) bool {
	value, ok := e.Bindings[name]
	return ok && value
}

func (e *DeclarativeEnvironment) CreateMutableBinding(name string, deletable bool) *Completion {
	// Assert that name is not already bound in an environment record.
	if _, ok := e.Bindings[name]; ok {
		panic("Assert failed: CreateMutableBinding called with a name that is already bound in an environment record.")
	}

	e.Bindings[name] = true
	e.MutableBindings[name] = true
	e.DeletableBindings[name] = deletable
	return NewUnusedCompletion()
}

func (e *DeclarativeEnvironment) CreateImmutableBinding(name string, strict bool) *Completion {
	// Assert that name is not already bound in an environment record.
	if _, ok := e.Bindings[name]; ok {
		panic("Assert failed: CreateMutableBinding called with a name that is already bound in an environment record.")
	}

	e.Bindings[name] = true
	e.StrictBindings[name] = strict
	return NewUnusedCompletion()
}

type GlobalEnvironment struct {
	DeclarativeRecord *DeclarativeEnvironment
	ObjectRecord      *ObjectEnvironment
}

func (e *GlobalEnvironment) HasBinding(name string) bool {
	// TODO: Confirm this is correct to the spec.
	return e.DeclarativeRecord.HasBinding(name)
}

func (e *GlobalEnvironment) CreateMutableBinding(name string, deletable bool) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return NewThrowCompletion(NewTypeError(fmt.Sprintf("Identifier '%s' has already been declared", name)))
	}

	return e.DeclarativeRecord.CreateMutableBinding(name, deletable)
}

func (e *GlobalEnvironment) CreateImmutableBinding(name string, strict bool) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return NewThrowCompletion(NewTypeError(fmt.Sprintf("Identifier '%s' has already been declared", name)))
	}

	return e.DeclarativeRecord.CreateImmutableBinding(name, strict)
}

func (e *GlobalEnvironment) CreateGlobalFunctionBinding(functionName string, functionObject *Function, deletable bool) *Completion {
	panic("not implemented")
}

func (e *GlobalEnvironment) CreateGlobalVarBinding(varName string, deletable bool) *Completion {
	panic("not implemented")
}
