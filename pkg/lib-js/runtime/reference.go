package runtime

import (
	"fmt"
	"strings"
)

type Reference struct {
	BaseEnv       Environment
	BaseObject    *JavaScriptValue
	ReferenceName *JavaScriptValue
	Strict        bool
	ThisValue     *JavaScriptValue
}

func NewReferenceValueForEnvironment(
	base Environment,
	referenceName string,
	strict bool,
	thisValue *JavaScriptValue,
) *JavaScriptValue {
	return NewJavaScriptValue(TypeReference, &Reference{
		BaseEnv:       base,
		BaseObject:    nil,
		ReferenceName: NewStringValue(referenceName),
		Strict:        strict,
		ThisValue:     thisValue,
	})
}

func NewReferenceValueForObject(
	base *JavaScriptValue,
	referenceName string,
	strict bool,
	thisValue *JavaScriptValue,
) *JavaScriptValue {
	return NewReferenceValueForObjectProperty(base, NewStringValue(referenceName), strict, thisValue)
}

func NewReferenceValueForObjectProperty(
	base *JavaScriptValue,
	propertyKey *JavaScriptValue,
	strict bool,
	thisValue *JavaScriptValue,
) *JavaScriptValue {
	return NewJavaScriptValue(TypeReference, &Reference{
		BaseEnv:       nil,
		BaseObject:    base,
		ReferenceName: propertyKey,
		Strict:        strict,
		ThisValue:     thisValue,
	})
}

func (r *Reference) InitializeReferencedBinding(runtime *Runtime, value *JavaScriptValue) *Completion {
	if r.BaseEnv == nil && r.BaseObject == nil {
		panic("Assert failed: InitializeReferencedBinding called on an unresolvable reference.")
	}

	if r.BaseEnv != nil {
		if r.ReferenceName.Type != TypeString {
			panic("Assert failed: Reference name is not a string.")
		}

		return r.BaseEnv.InitializeBinding(runtime, r.ReferenceName.Value.(*String).Value, value)
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

	return r.BaseObject
}

func GetValue(runtime *Runtime, maybeRef *JavaScriptValue) *Completion {
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
		baseObjectCompletion := ToObject(ref.BaseObject)
		if baseObjectCompletion.Type != Normal {
			return baseObjectCompletion
		}

		baseObject := baseObjectCompletion.Value.(*JavaScriptValue).Value.(ObjectInterface)

		propertyKeyCompletion := ToPropertyKey(ref.ReferenceName)
		if propertyKeyCompletion.Type != Normal {
			return propertyKeyCompletion
		}

		propertyKey := propertyKeyCompletion.Value.(*JavaScriptValue)

		if propertyKey.Type == TypeSymbol {
			panic("TODO: Support getting symbol properties.")
		}

		if propertyKey.Type != TypeString {
			panic("Assert failed: Property key is not a string.")
		}

		propertyKeyString := propertyKey.Value.(*String).Value
		if strings.HasPrefix(propertyKeyString, "#") {
			panic("TODO: Support getting private object properties.")
		}

		ref.ReferenceName = propertyKey
		return baseObject.Get(runtime, propertyKey, ref.GetThisValue())
	}

	if ref.ReferenceName.Type != TypeString {
		// TODO: This assertion is not in the spec, unsure if it's needed.
		panic("Assert failed: Reference name is not a string.")
	}

	return ref.BaseEnv.GetBindingValue(runtime, ref.ReferenceName.Value.(*String).Value, ref.Strict)
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

		completion := globalObject.Set(runtime, ref.ReferenceName, value, NewJavaScriptValue(TypeObject, globalObject))
		if completion.Type != Normal {
			return completion
		}

		return NewUnusedCompletion()
	}

	if ref.BaseObject != nil {
		baseObjectCompletion := ToObject(ref.BaseObject)
		if baseObjectCompletion.Type != Normal {
			return baseObjectCompletion
		}

		baseObject := baseObjectCompletion.Value.(*JavaScriptValue).Value.(ObjectInterface)

		if ref.ReferenceName.Type == TypeString && strings.HasPrefix(ref.ReferenceName.Value.(*String).Value, "#") {
			panic("TODO: Support setting private object properties.")
		}

		refNamePrimitive := ToPrimitive(ref.ReferenceName)
		if refNamePrimitive.Type != Normal {
			return refNamePrimitive
		}

		refNameVal := refNamePrimitive.Value.(*JavaScriptValue)
		propertyKeyCompletion := ToPropertyKey(refNameVal)

		if propertyKeyCompletion.Type != Normal {
			return propertyKeyCompletion
		}

		ref.ReferenceName = propertyKeyCompletion.Value.(*JavaScriptValue)

		// Set the property on the object.
		successCompletion := baseObject.Set(runtime, ref.ReferenceName, value, ref.GetThisValue())
		if successCompletion.Type != Normal {
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

	return ref.BaseEnv.SetMutableBinding(runtime, ref.ReferenceName.Value.(*String).Value, value, ref.Strict)
}

func PropertyKeyToString(value *JavaScriptValue) string {
	switch value.Type {
	case TypeSymbol:
		return value.Value.(*Symbol).Description
	case TypeString:
		return value.Value.(*String).Value
	}

	panic("Assert failed: Reference name is not a string or symbol.")
}

func ToPropertyKey(value *JavaScriptValue) *Completion {
	key := ToPrimitive(value)
	if key.Type != Normal {
		return key
	}

	keyVal := key.Value.(*JavaScriptValue)

	if keyVal.Type == TypeSymbol {
		return key
	}

	return ToString(keyVal)
}
