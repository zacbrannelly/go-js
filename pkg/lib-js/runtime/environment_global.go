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

func (e *GlobalEnvironment) GetBindingValue(runtime *Runtime, name string, strict bool) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return e.DeclarativeRecord.GetBindingValue(runtime, name, strict)
	}

	return e.ObjectRecord.GetBindingValue(runtime, name, strict)
}

func (e *GlobalEnvironment) InitializeBinding(runtime *Runtime, name string, value *JavaScriptValue) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return e.DeclarativeRecord.InitializeBinding(runtime, name, value)
	}

	return e.ObjectRecord.InitializeBinding(runtime, name, value)
}

func (e *GlobalEnvironment) SetMutableBinding(runtime *Runtime, name string, value *JavaScriptValue, strict bool) *Completion {
	if e.DeclarativeRecord.HasBinding(name) {
		return e.DeclarativeRecord.SetMutableBinding(runtime, name, value, strict)
	}

	return e.ObjectRecord.SetMutableBinding(runtime, name, value, strict)
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
	globalObject := env.ObjectRecord.BindingObject
	functionNameValue := NewStringValue(functionName)
	existingPropCompletion := globalObject.GetOwnProperty(functionNameValue)
	if existingPropCompletion.Type != Normal {
		return existingPropCompletion
	}

	existingDescriptor, hasOwnProperty := existingPropCompletion.Value.(PropertyDescriptor)
	if !hasOwnProperty || existingDescriptor == nil {
		return NewNormalCompletion(NewBooleanValue(globalObject.GetExtensible()))
	}

	if existingDescriptor.GetConfigurable() {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	if dataDescriptor, ok := existingDescriptor.(*DataPropertyDescriptor); ok {
		if dataDescriptor.Writable && dataDescriptor.Enumerable {
			return NewNormalCompletion(NewBooleanValue(true))
		}
	}

	return NewNormalCompletion(NewBooleanValue(false))
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

func (e *GlobalEnvironment) CreateGlobalVarBinding(runtime *Runtime, varName string, deletable bool) *Completion {
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
		completion = e.ObjectRecord.InitializeBinding(runtime, varName, NewUndefinedValue())
		if completion.Type != Normal {
			return completion
		}
	}

	return NewUnusedCompletion()
}

func (e *GlobalEnvironment) CreateGlobalFunctionBinding(
	runtime *Runtime,
	functionName string,
	functionObject *FunctionObject,
	deletable bool,
) *Completion {
	functionNameValue := NewStringValue(functionName)
	globalObject := e.ObjectRecord.BindingObject

	existingPropCompletion := globalObject.GetOwnProperty(functionNameValue)
	if existingPropCompletion.Type != Normal {
		return existingPropCompletion
	}

	existingDescriptor, hasOwnProperty := existingPropCompletion.Value.(PropertyDescriptor)

	value := NewJavaScriptValue(TypeObject, functionObject)
	globalObjectValue := NewJavaScriptValue(TypeObject, globalObject)

	var descriptor PropertyDescriptor
	if !hasOwnProperty || existingDescriptor.GetConfigurable() {
		descriptor = &DataPropertyDescriptor{
			Value:        value,
			Writable:     true,
			Enumerable:   true,
			Configurable: deletable,
		}
	} else {
		descriptor = existingDescriptor.Copy()
		descriptor.(*DataPropertyDescriptor).Value = value
	}

	completion := DefinePropertyOrThrow(globalObject, functionNameValue, descriptor)
	if completion.Type != Normal {
		return completion
	}

	completion = globalObject.Set(runtime, functionNameValue, value, globalObjectValue)
	if completion.Type != Normal {
		return completion
	}

	return NewUnusedCompletion()
}

func (e *GlobalEnvironment) WithBaseObject() *JavaScriptValue {
	return NewUndefinedValue()
}

func (e *GlobalEnvironment) GetThisBinding() *Completion {
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, e.GlobalThisValue))
}
