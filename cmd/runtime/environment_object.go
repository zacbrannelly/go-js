package runtime

import "fmt"

type ObjectEnvironment struct {
	OuterEnv          Environment
	BindingObject     ObjectInterface
	IsWithEnvironment bool
}

func NewObjectEnvironment(bindingObject ObjectInterface, isWithEnvironment bool, outerEnv Environment) *ObjectEnvironment {
	return &ObjectEnvironment{
		OuterEnv:          outerEnv,
		BindingObject:     bindingObject,
		IsWithEnvironment: isWithEnvironment,
	}
}

func (e *ObjectEnvironment) GetOuterEnvironment() Environment {
	return e.OuterEnv
}

func (e *ObjectEnvironment) HasBinding(name string) bool {
	bindingObj := e.BindingObject

	hasPropertyCompletion := bindingObj.HasProperty(NewStringValue(name))
	if hasPropertyCompletion.Type != Normal {
		// TODO: If this happens, we need to change this function to return a completion.
		panic("Assert failed: bindingObject.HasProperty threw an error.")
	}

	hasPropertyVal := hasPropertyCompletion.Value.(*JavaScriptValue)
	if !hasPropertyVal.Value.(*Boolean).Value {
		return false
	}

	if !e.IsWithEnvironment {
		return true
	}

	// TODO: Return false if the name is in the %unscopables% symbol object.

	return true
}

func (e *ObjectEnvironment) CreateMutableBinding(name string, deletable bool) *Completion {
	completion := DefinePropertyOrThrow(e.BindingObject, NewStringValue(name), &DataPropertyDescriptor{
		Value:        NewUndefinedValue(),
		Writable:     true,
		Enumerable:   true,
		Configurable: deletable,
	})
	if completion.Type != Normal {
		return completion
	}

	return NewUnusedCompletion()
}

func (e *ObjectEnvironment) CreateImmutableBinding(name string, strict bool) *Completion {
	panic("Assert failed: This should never be called.")
}

func (e *ObjectEnvironment) GetBindingValue(name string, strict bool) *Completion {
	bindingObj := e.BindingObject
	nameValue := NewStringValue(name)

	existsCompletion := bindingObj.HasProperty(nameValue)
	if existsCompletion.Type != Normal {
		return existsCompletion
	}

	if existsVal, ok := existsCompletion.Value.(*Boolean); ok && !existsVal.Value {
		if strict {
			return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Unresolvable reference '%s'", name)))
		}
		return NewNormalCompletion(NewUndefinedValue())
	}

	return bindingObj.Get(nameValue, NewJavaScriptValue(TypeObject, bindingObj))
}

func (e *ObjectEnvironment) InitializeBinding(name string, value *JavaScriptValue) *Completion {
	completion := e.SetMutableBinding(name, value, false)
	if completion.Type != Normal {
		return completion
	}

	return NewUnusedCompletion()
}

func (e *ObjectEnvironment) SetMutableBinding(name string, value *JavaScriptValue, strict bool) *Completion {
	bindingObj := e.BindingObject
	nameValue := NewStringValue(name)

	existsCompletion := bindingObj.HasProperty(nameValue)
	if existsCompletion.Type != Normal {
		return existsCompletion
	}

	if existsVal, ok := existsCompletion.Value.(*Boolean); ok && !existsVal.Value && strict {
		return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Unresolvable reference '%s'", name)))
	}

	// The following steps are based on: 7.3.4 Set ( O, P, V, Throw )
	successCompletion := bindingObj.Set(nameValue, value, NewJavaScriptValue(TypeObject, bindingObj))
	if successCompletion.Type != Normal {
		return successCompletion
	}

	if successVal, ok := successCompletion.Value.(*Boolean); ok && !successVal.Value && strict {
		return NewThrowCompletion(NewTypeError(fmt.Sprintf("Cannot assign to read only property '%s'", name)))
	}

	return NewUnusedCompletion()
}

func (e *ObjectEnvironment) DeleteBinding(name string) *Completion {
	return e.BindingObject.Delete(NewStringValue(name))
}
