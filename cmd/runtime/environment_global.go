package runtime

import "fmt"

type GlobalEnvironment struct {
	DeclarativeRecord *DeclarativeEnvironment
	ObjectRecord      *ObjectEnvironment
	GlobalThisValue   *Object
}

func NewGlobalEnvironment(globalObject *Object, thisValue *Object) *GlobalEnvironment {
	return &GlobalEnvironment{
		DeclarativeRecord: NewDeclarativeEnvironment(nil),
		ObjectRecord:      NewObjectEnvironment(globalObject, false, nil),
		GlobalThisValue:   thisValue,
	}
}

func (e *GlobalEnvironment) GetOuterEnvironment() Environment {
	// Global environment are the outer environment.
	return nil
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

func (e *GlobalEnvironment) GetBindingValue(name string, strict bool) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return e.DeclarativeRecord.GetBindingValue(name, strict)
	}

	return e.ObjectRecord.GetBindingValue(name, strict)
}

func (e *GlobalEnvironment) InitializeBinding(name string, value *JavaScriptValue) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return e.DeclarativeRecord.InitializeBinding(name, value)
	}

	return e.ObjectRecord.InitializeBinding(name, value)
}

func (e *GlobalEnvironment) CreateGlobalFunctionBinding(functionName string, functionObject *Function, deletable bool) *Completion {
	panic("not implemented")
}

func (e *GlobalEnvironment) CreateGlobalVarBinding(varName string, deletable bool) *Completion {
	panic("not implemented")
}
