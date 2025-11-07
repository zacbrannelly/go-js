package runtime

import "fmt"

type GlobalEnvironment struct {
	DeclarativeRecord *DeclarativeEnvironment
	ObjectRecord      *ObjectEnvironment
	GlobalThisValue   ObjectInterface
}

func NewGlobalEnvironment(globalObject ObjectInterface, thisValue ObjectInterface) *GlobalEnvironment {
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
	if e.DeclarativeRecord.HasBinding(name) {
		return true
	}

	return e.ObjectRecord.HasBinding(name)
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

func (e *GlobalEnvironment) SetMutableBinding(name string, value *JavaScriptValue, strict bool) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return e.DeclarativeRecord.SetMutableBinding(name, value, strict)
	}

	return e.ObjectRecord.SetMutableBinding(name, value, strict)
}

func (e *GlobalEnvironment) DeleteBinding(name string) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return e.DeclarativeRecord.DeleteBinding(name)
	}

	globalObject := e.ObjectRecord.BindingObject
	existingPropCompletion := HasOwnProperty(globalObject, NewStringValue(name))
	if existingPropCompletion.Type != Normal {
		return existingPropCompletion
	}

	existingPropVal := existingPropCompletion.Value.(*JavaScriptValue)
	if existingPropVal.Type != TypeBoolean {
		panic("Assert failed: Expected a boolean value for HasOwnProperty.")
	}

	if existingPropVal.Value.(*Boolean).Value {
		return e.ObjectRecord.DeleteBinding(name)
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func CanDeclareGlobalFunction(env *GlobalEnvironment, functionName string) *Completion {
	panic("not implemented")
}

func CanDeclareGlobalVar(env *GlobalEnvironment, varName string) *Completion {
	globalObject := env.ObjectRecord.BindingObject
	hasOwnCompletion := HasOwnProperty(globalObject, NewStringValue(varName))
	if hasOwnCompletion.Type != Normal {
		return hasOwnCompletion
	}

	hasOwnVal := hasOwnCompletion.Value.(*JavaScriptValue)
	if hasOwnVal.Type != TypeBoolean {
		panic("Assert failed: Expected a boolean value for HasOwnProperty.")
	}

	if hasOwnVal.Value.(*Boolean).Value {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	return NewNormalCompletion(NewBooleanValue(globalObject.GetExtensible()))
}

func (e *GlobalEnvironment) CreateGlobalVarBinding(varName string, deletable bool) *Completion {
	globalObject := e.ObjectRecord.BindingObject
	hasOwnCompletion := HasOwnProperty(globalObject, NewStringValue(varName))
	if hasOwnCompletion.Type != Normal {
		return hasOwnCompletion
	}

	hasOwnVal := hasOwnCompletion.Value.(*JavaScriptValue)
	if hasOwnVal.Type != TypeBoolean {
		panic("Assert failed: Expected a boolean value for HasOwnProperty.")
	}

	if !hasOwnVal.Value.(*Boolean).Value && globalObject.GetExtensible() {
		completion := e.ObjectRecord.CreateMutableBinding(varName, deletable)
		if completion.Type != Normal {
			return completion
		}
		completion = e.ObjectRecord.InitializeBinding(varName, NewUndefinedValue())
		if completion.Type != Normal {
			return completion
		}
	}

	return NewUnusedCompletion()
}

func (e *GlobalEnvironment) CreateGlobalFunctionBinding(functionName string, functionObject *Function, deletable bool) *Completion {
	panic("not implemented")
}
