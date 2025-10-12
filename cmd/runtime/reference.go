package runtime

import (
	"fmt"
	"strings"
)

type Reference struct {
	BaseEnv       Environment
	BaseObject    *Object
	ReferenceName *JavaScriptValue
	Strict        bool
	ThisValue     *JavaScriptValue
}

func NewReferenceValueForEnvironment(base Environment, referenceName string, strict bool, thisValue *JavaScriptValue) *JavaScriptValue {
	return NewJavaScriptValue(TypeReference, &Reference{
		BaseEnv:       base,
		BaseObject:    nil,
		ReferenceName: NewStringValue(referenceName),
		Strict:        strict,
		ThisValue:     thisValue,
	})
}

func NewReferenceValueForObject(base *Object, referenceName string, strict bool, thisValue *JavaScriptValue) *JavaScriptValue {
	return NewJavaScriptValue(TypeReference, &Reference{
		BaseEnv:       nil,
		BaseObject:    base,
		ReferenceName: NewStringValue(referenceName),
		Strict:        strict,
		ThisValue:     thisValue,
	})
}

func (r *Reference) InitializeReferencedBinding(value *JavaScriptValue) *Completion {
	if r.BaseEnv == nil && r.BaseObject == nil {
		panic("Assert failed: InitializeReferencedBinding called on an unresolvable reference.")
	}

	if r.BaseEnv != nil {
		if r.ReferenceName.Type != TypeString {
			panic("Assert failed: Reference name is not a string.")
		}

		return r.BaseEnv.InitializeBinding(r.ReferenceName.Value.(*String).Value, value)
	}

	panic("TODO: Property reference not implemented in InitializeReferencedBinding.")
}

func (r *Reference) IsSuperReference() bool {
	return r.ThisValue != nil
}

func (r *Reference) GetThisValue() *JavaScriptValue {
	if r.BaseEnv != nil || (r.BaseObject == nil && r.BaseEnv == nil) {
		panic("Assert failed: GetThisValue called on an unresolvable reference or a reference to an environment record.")
	}

	if r.IsSuperReference() {
		return r.ThisValue
	}

	return NewJavaScriptValue(TypeObject, r.BaseObject)
}

func GetValue(maybeRef *JavaScriptValue) *Completion {
	if maybeRef.Type != TypeReference {
		return NewNormalCompletion(maybeRef)
	}

	ref := maybeRef.Value.(*Reference)
	if ref.BaseEnv == nil && ref.BaseObject == nil {
		// Unresolvable reference.
		refNameString := PropertyKeyToString(ref.ReferenceName)
		return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Unresolvable reference '%s'", refNameString)))
	}

	if ref.BaseObject != nil {
		panic("TODO: Property reference not implemented in GetValue.")
	}

	if ref.ReferenceName.Type != TypeString {
		panic("Assert failed: Reference name is not a string.")
	}

	return ref.BaseEnv.GetBindingValue(ref.ReferenceName.Value.(*String).Value, ref.Strict)
}

func PutValue(runtime *Runtime, maybeRef *JavaScriptValue, value *JavaScriptValue) *Completion {
	if maybeRef.Type != TypeReference {
		return NewThrowCompletion(NewTypeError("Cannot assign to a non-reference value."))
	}

	ref := maybeRef.Value.(*Reference)
	if ref.BaseEnv == nil && ref.BaseObject == nil {
		// Unresolvable reference.
		if ref.Strict {
			refNameString := PropertyKeyToString(ref.ReferenceName)
			return NewThrowCompletion(NewReferenceError(fmt.Sprintf("Cannot assign to an unresolvable reference '%s'", refNameString)))
		}

		runningContext := runtime.GetRunningExecutionContext()
		if runningContext == nil {
			panic("Assert failed: Running execution context is nil.")
		}
		if runningContext.Realm == nil {
			panic("Assert failed: Running execution context has no realm.")
		}

		globalObject := runningContext.Realm.GlobalObject
		if globalObject == nil {
			panic("Assert failed: Running execution context has no global object.")
		}

		completion := globalObject.Set(ref.ReferenceName, value, NewJavaScriptValue(TypeObject, globalObject))
		if completion.Type == Throw {
			return completion
		}

		return NewUnusedCompletion()
	}

	if ref.BaseObject != nil {
		if ref.ReferenceName.Type == TypeString && strings.HasPrefix(ref.ReferenceName.Value.(*String).Value, "#") {
			panic("TODO: Support setting private object properties.")
		}

		refNamePrimitive := ToPrimitive(runtime, ref.ReferenceName)
		if refNamePrimitive.Type == Throw {
			return refNamePrimitive
		}

		refNameVal := refNamePrimitive.Value.(*JavaScriptValue)
		propertyKeyCompletion := ToPropertyKey(runtime, refNameVal)

		if propertyKeyCompletion.Type == Throw {
			return propertyKeyCompletion
		}

		ref.ReferenceName = propertyKeyCompletion.Value.(*JavaScriptValue)

		// Set the property on the object.
		successCompletion := ref.BaseObject.Set(ref.ReferenceName, value, ref.GetThisValue())
		if successCompletion.Type == Throw {
			return successCompletion
		}

		successVal := successCompletion.Value.(*JavaScriptValue)
		if successVal.Type != TypeBoolean {
			panic("Assert failed: Expected a boolean value for Set.")
		}

		// If failed to set property and in strict mode, throw a TypeError.
		if !successVal.Value.(*Boolean).Value && ref.Strict {
			refNameString := PropertyKeyToString(ref.ReferenceName)
			return NewThrowCompletion(NewTypeError(fmt.Sprintf("Cannot set '%s' property of this object.", refNameString)))
		}

		return NewUnusedCompletion()
	}

	// TODO: Can it be a symbol in this path?
	if ref.ReferenceName.Type != TypeString {
		panic("Assert failed: Reference name is not a string.")
	}

	return ref.BaseEnv.SetMutableBinding(ref.ReferenceName.Value.(*String).Value, value, ref.Strict)
}

func PropertyKeyToString(value *JavaScriptValue) string {
	switch value.Type {
	case TypeSymbol:
		return value.Value.(*Symbol).Name
	case TypeString:
		return value.Value.(*String).Value
	}

	panic("Assert failed: Reference name is not a string or symbol.")
}

func ToPropertyKey(runtime *Runtime, value *JavaScriptValue) *Completion {
	key := ToPrimitive(runtime, value)
	if key.Type == Throw {
		return key
	}

	keyVal := key.Value.(*JavaScriptValue)

	if keyVal.Type == TypeSymbol {
		return key
	}

	return ToString(runtime, keyVal)
}
