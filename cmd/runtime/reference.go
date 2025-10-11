package runtime

import "fmt"

type Reference struct {
	BaseEnv       Environment
	BaseObject    *Object
	ReferenceName string
	Strict        bool
	ThisValue     *JavaScriptValue
}

func NewReferenceValueForEnvironment(base Environment, referenceName string, strict bool, thisValue *JavaScriptValue) *JavaScriptValue {
	return NewJavaScriptValue(TypeReference, &Reference{
		BaseEnv:       base,
		BaseObject:    nil,
		ReferenceName: referenceName,
		Strict:        strict,
		ThisValue:     thisValue,
	})
}

func NewReferenceValueForObject(base *Object, referenceName string, strict bool, thisValue *JavaScriptValue) *JavaScriptValue {
	return NewJavaScriptValue(TypeReference, &Reference{
		BaseEnv:       nil,
		BaseObject:    base,
		ReferenceName: referenceName,
		Strict:        strict,
		ThisValue:     thisValue,
	})
}

func GetValue(maybeRef *JavaScriptValue) *Completion {
	if maybeRef.Type != TypeReference {
		return NewNormalCompletion(maybeRef)
	}

	ref := maybeRef.Value.(*Reference)
	if ref.BaseEnv == nil && ref.BaseObject == nil {
		// Unresolvable reference.
		return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Unresolvable reference '%s'", ref.ReferenceName)))
	}

	if ref.BaseObject != nil {
		panic("TODO: Property reference not implemented in GetValue.")
	}

	return ref.BaseEnv.GetBindingValue(ref.ReferenceName, ref.Strict)
}

func (r *Reference) InitializeReferencedBinding(value *JavaScriptValue) *Completion {
	if r.BaseEnv == nil && r.BaseObject == nil {
		panic("Assert failed: InitializeReferencedBinding called on an unresolvable reference.")
	}

	if r.BaseEnv != nil {
		return r.BaseEnv.InitializeBinding(r.ReferenceName, value)
	}

	panic("TODO: Property reference not implemented in InitializeReferencedBinding.")
}
