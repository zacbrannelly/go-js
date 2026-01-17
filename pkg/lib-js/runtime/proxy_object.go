package runtime

import "slices"

type ProxyObject struct {
	ProxyHandler *JavaScriptValue
	ProxyTarget  *JavaScriptValue

	PrivateElements []*PrivateElement

	HasConstruct bool
	HasCall      bool
}

var (
	getPrototypeOfKey    = NewStringValue("getPrototypeOf")
	setPrototypeOfKey    = NewStringValue("setPrototypeOf")
	getOwnPropertyKey    = NewStringValue("getOwnPropertyDescriptor")
	isExtensibleKey      = NewStringValue("isExtensible")
	hasKey               = NewStringValue("has")
	definePropertyKey    = NewStringValue("defineProperty")
	deleteKey            = NewStringValue("delete")
	ownKeysStr           = NewStringValue("ownKeys")
	preventExtensionsKey = NewStringValue("preventExtensions")
	applyStr             = NewStringValue("apply")
	constructStr         = NewStringValue("construct")
)

func ProxyCreate(runtime *Runtime, target *JavaScriptValue, handler *JavaScriptValue) *Completion {
	if target.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Target must be an object."))
	}

	if handler.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Handler must be an object."))
	}

	isCallable := IsCallable(target)
	isConstructor := false

	if isCallable {
		if constructor, ok := target.Value.(FunctionInterface); ok {
			isConstructor = constructor.HasConstructMethod()
		}
	}

	proxy := &ProxyObject{
		ProxyHandler:    handler,
		ProxyTarget:     target,
		PrivateElements: make([]*PrivateElement, 0),
		HasCall:         isCallable,
		HasConstruct:    isConstructor,
	}
	return NewNormalCompletion(NewJavaScriptValue(TypeObject, proxy))
}

func (o *ProxyObject) GetPrototype() ObjectInterface {
	panic("Assert failed: ProxyObject.GetPrototype called.")
}

func (o *ProxyObject) SetPrototype(prototype ObjectInterface) {
	panic("Assert failed: ProxyObject.SetPrototype called.")
}

func (o *ProxyObject) GetProperties() map[string]PropertyDescriptor {
	panic("Assert failed: ProxyObject.GetProperties called.")
}

func (o *ProxyObject) SetProperties(properties map[string]PropertyDescriptor) {
	panic("Assert failed: ProxyObject.SetProperties called.")
}

func (o *ProxyObject) GetSymbolProperties() map[*Symbol]PropertyDescriptor {
	panic("Assert failed: ProxyObject.GetSymbolProperties called.")
}

func (o *ProxyObject) SetSymbolProperties(symbolProperties map[*Symbol]PropertyDescriptor) {
	panic("Assert failed: ProxyObject.SetSymbolProperties called.")
}

func (o *ProxyObject) IsExtensible(runtime *Runtime) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	if handler.Type != TypeObject {
		panic("Assert failed: ProxyHandler is not an Object.")
	}

	completion = GetMethod(runtime, handler, isExtensibleKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).IsExtensible(runtime)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = ToBoolean(trapResult)
	if completion.Type != Normal {
		return completion
	}

	booleanTrapResult := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	completion = target.Value.(ObjectInterface).IsExtensible(runtime)
	if completion.Type != Normal {
		return completion
	}

	targetResult := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if booleanTrapResult != targetResult {
		return NewThrowCompletion(NewTypeError(runtime, "Proxy isExtensible trap returned inconsistent result."))
	}

	return NewNormalCompletion(NewBooleanValue(booleanTrapResult))
}

func (o *ProxyObject) GetPrototypeOf(runtime *Runtime) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, getPrototypeOfKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).GetPrototypeOf(runtime)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target})
	if completion.Type != Normal {
		return completion
	}

	handlerProtoVal := completion.Value.(*JavaScriptValue)
	if handlerProtoVal.Type != TypeObject && handlerProtoVal.Type != TypeNull {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid handler prototype value."))
	}

	targetObj := target.Value.(ObjectInterface)

	completion = targetObj.IsExtensible(runtime)
	if completion.Type != Normal {
		return completion
	}

	isExtensible := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if isExtensible {
		// Return the handler prototype value.
		return NewNormalCompletion(handlerProtoVal)
	}

	completion = targetObj.GetPrototypeOf(runtime)
	if completion.Type != Normal {
		return completion
	}

	targetProtoVal := completion.Value.(*JavaScriptValue)

	completion = SameValue(handlerProtoVal, targetProtoVal)
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid prototype object."))
	}

	return NewNormalCompletion(handlerProtoVal)
}

func (o *ProxyObject) SetPrototypeOf(runtime *Runtime, prototype *JavaScriptValue) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, setPrototypeOfKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).SetPrototypeOf(runtime, prototype)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, prototype})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = ToBoolean(trapResult)
	if completion.Type != Normal {
		return completion
	}

	booleanTrapResult := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value
	if !booleanTrapResult {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	targetObj := target.Value.(ObjectInterface)

	completion = targetObj.IsExtensible(runtime)
	if completion.Type != Normal {
		return completion
	}

	isExtensible := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if isExtensible {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	targetProto := targetObj.GetPrototypeOf(runtime)
	if targetProto.Type != Normal {
		return targetProto
	}

	targetProtoVal := targetProto.Value.(*JavaScriptValue)

	completion = SameValue(prototype, targetProtoVal)
	if completion.Type != Normal {
		return completion
	}

	if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid prototype object."))
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *ProxyObject) GetOwnProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, getOwnPropertyKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).GetOwnProperty(runtime, key)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, key})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	if trapResult.Type != TypeObject && trapResult.Type != TypeUndefined {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid trap result."))
	}

	completion = target.Value.(ObjectInterface).GetOwnProperty(runtime, key)
	if completion.Type != Normal {
		return completion
	}

	targetDesc, _ := completion.Value.(PropertyDescriptor)

	if trapResult.Type == TypeUndefined {
		if completion.Value == nil {
			return NewNormalCompletion(NewUndefinedValue())
		}

		if !targetDesc.GetConfigurable() {
			return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
		}

		completion = target.Value.(ObjectInterface).IsExtensible(runtime)
		if completion.Type != Normal {
			return completion
		}

		if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return NewThrowCompletion(NewTypeError(runtime, "Target is not extensible."))
		}

		return NewNormalCompletion(nil)
	}

	completion = target.Value.(ObjectInterface).IsExtensible(runtime)
	if completion.Type != Normal {
		return completion
	}

	extensibleTarget := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	completion = ToPropertyDescriptor(runtime, trapResult)
	if completion.Type != Normal {
		return completion
	}

	resultDesc := completion.Value.(*JavaScriptValue).Value.(PropertyDescriptor)

	// TODO: Implement CompletePropertyDescriptor function.

	compatibleVal := IsCompatiblePropertyDescriptor(extensibleTarget, resultDesc, targetDesc)
	compatible := compatibleVal.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if !compatible {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
	}

	if !resultDesc.GetConfigurable() {
		if targetDesc == nil || targetDesc.GetConfigurable() {
			return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
		}

		if dataDesc, ok := resultDesc.(*DataPropertyDescriptor); ok && !dataDesc.Writable {
			if targetDataDesc, ok := targetDesc.(*DataPropertyDescriptor); ok && targetDataDesc.Writable {
				return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
			}
		}
	}

	return NewNormalCompletion(resultDesc)
}

func (o *ProxyObject) HasProperty(runtime *Runtime, key *JavaScriptValue) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, hasKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).HasProperty(runtime, key)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, key})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = ToBoolean(trapResult)
	if completion.Type != Normal {
		return completion
	}

	booleanTrapResultVal := completion.Value.(*JavaScriptValue)
	booleanTrapResult := booleanTrapResultVal.Value.(*Boolean).Value

	if !booleanTrapResult {
		targetObj := target.Value.(ObjectInterface)
		completion = targetObj.GetOwnProperty(runtime, key)
		if completion.Type != Normal {
			return completion
		}

		if completion.Value != nil {
			targetDesc := completion.Value.(PropertyDescriptor)
			if !targetDesc.GetConfigurable() {
				return NewThrowCompletion(NewTypeError(runtime, "Target property is not configurable."))
			}

			completion = targetObj.IsExtensible(runtime)
			if completion.Type != Normal {
				return completion
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError(runtime, "Target is not extensible."))
			}
		}
	}

	return NewNormalCompletion(booleanTrapResultVal)
}

func (o *ProxyObject) DefineOwnProperty(runtime *Runtime, key *JavaScriptValue, descriptor PropertyDescriptor) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, definePropertyKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).DefineOwnProperty(runtime, key, descriptor)
	}

	descriptorVal := FromPropertyDescriptor(runtime, descriptor)
	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, key, descriptorVal})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = ToBoolean(trapResult)
	if completion.Type != Normal {
		return completion
	}

	booleanTrapResult := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if !booleanTrapResult {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	targetObj := target.Value.(ObjectInterface)
	completion = targetObj.GetOwnProperty(runtime, key)

	if completion.Type != Normal {
		return completion
	}

	targetDesc, _ := completion.Value.(PropertyDescriptor)

	completion = targetObj.IsExtensible(runtime)
	if completion.Type != Normal {
		return completion
	}

	extensibleTarget := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	settingConfigFalse := false

	if !descriptor.GetConfigurable() {
		settingConfigFalse = true
	}

	if targetDesc == nil {
		if !extensibleTarget {
			return NewThrowCompletion(NewTypeError(runtime, "Target is not extensible."))
		}

		if settingConfigFalse {
			return NewThrowCompletion(NewTypeError(runtime, "Descriptor is not configurable."))
		}
	} else {
		compatibleVal := IsCompatiblePropertyDescriptor(extensibleTarget, descriptor, targetDesc)
		compatible := compatibleVal.Value.(*JavaScriptValue).Value.(*Boolean).Value

		if !compatible {
			return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
		}

		if settingConfigFalse && targetDesc.GetConfigurable() {
			return NewThrowCompletion(NewTypeError(runtime, "Descriptor is not configurable."))
		}

		if dataDesc, ok := targetDesc.(*DataPropertyDescriptor); ok && !dataDesc.Configurable && dataDesc.Writable {
			if descriptorDataDesc, ok := descriptor.(*DataPropertyDescriptor); ok && !descriptorDataDesc.Writable {
				return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
			}
		}
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *ProxyObject) Set(runtime *Runtime, key *JavaScriptValue, value *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	if handler.Type != TypeObject {
		panic("Assert failed: ProxyHandler is not an Object.")
	}

	completion = GetMethod(runtime, handler, setKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).Set(runtime, key, value, receiver)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, key, value, receiver})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = ToBoolean(trapResult)
	if completion.Type != Normal {
		return completion
	}

	booleanTrapResult := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if !booleanTrapResult {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	completion = target.Value.(ObjectInterface).GetOwnProperty(runtime, key)
	if completion.Type != Normal {
		return completion
	}

	targetDesc, _ := completion.Value.(PropertyDescriptor)

	if targetDesc != nil && !targetDesc.GetConfigurable() {
		if dataDesc, ok := targetDesc.(*DataPropertyDescriptor); ok && !dataDesc.Writable {
			completion = SameValue(value, dataDesc.Value)
			if completion.Type != Normal {
				return completion
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
			}
		}

		if accessorDesc, ok := targetDesc.(*AccessorPropertyDescriptor); ok && accessorDesc.Set == nil {
			return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
		}
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *ProxyObject) Get(runtime *Runtime, key *JavaScriptValue, receiver *JavaScriptValue) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, getKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).Get(runtime, key, receiver)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, key, receiver})

	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = target.Value.(ObjectInterface).GetOwnProperty(runtime, key)
	if completion.Type != Normal {
		return completion
	}

	targetDesc, _ := completion.Value.(PropertyDescriptor)

	if targetDesc != nil && !targetDesc.GetConfigurable() {
		if dataDesc, ok := targetDesc.(*DataPropertyDescriptor); ok && !dataDesc.Writable {
			completion = SameValue(trapResult, dataDesc.Value)
			if completion.Type != Normal {
				return completion
			}

			if !completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
				return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
			}
		}

		if accessorDesc, ok := targetDesc.(*AccessorPropertyDescriptor); ok && accessorDesc.Get == nil {
			if trapResult.Type != TypeUndefined {
				return NewThrowCompletion(NewTypeError(runtime, "Invalid target property descriptor."))
			}
		}
	}

	return NewNormalCompletion(trapResult)
}

func (o *ProxyObject) Delete(runtime *Runtime, key *JavaScriptValue) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	if handler.Type != TypeObject {
		panic("Assert failed: ProxyHandler is not an Object.")
	}

	completion = GetMethod(runtime, handler, deleteKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).Delete(runtime, key)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, key})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = ToBoolean(trapResult)
	if completion.Type != Normal {
		return completion
	}

	booleanTrapResult := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if !booleanTrapResult {
		return NewNormalCompletion(NewBooleanValue(false))
	}

	targetObj := target.Value.(ObjectInterface)
	completion = targetObj.GetOwnProperty(runtime, key)
	if completion.Type != Normal {
		return completion
	}

	if completion.Value == nil {
		return NewNormalCompletion(NewBooleanValue(true))
	}

	targetDesc := completion.Value.(PropertyDescriptor)

	if !targetDesc.GetConfigurable() {
		return NewThrowCompletion(NewTypeError(runtime, "Target property is not configurable."))
	}

	completion = targetObj.IsExtensible(runtime)
	if completion.Type != Normal {
		return completion
	}

	extensibleTarget := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if !extensibleTarget {
		return NewThrowCompletion(NewTypeError(runtime, "Target is not extensible."))
	}

	return NewNormalCompletion(NewBooleanValue(true))
}

func (o *ProxyObject) OwnPropertyKeys(runtime *Runtime) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, ownKeysStr)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).OwnPropertyKeys(runtime)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target})
	if completion.Type != Normal {
		return completion
	}

	trapResultArray := completion.Value.(*JavaScriptValue)

	if trapResultArray.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid trap result."))
	}

	completion = LengthOfArrayLike(runtime, trapResultArray.Value.(ObjectInterface))
	if completion.Type != Normal {
		return completion
	}

	// The following code is the semantics of CreateListFromArrayLike with the PROPERTY-KEY element type.
	length := int(completion.Value.(*JavaScriptValue).Value.(*Number).Value)
	trapResult := make([]*JavaScriptValue, 0)

	for idx := range length {
		completion = ToString(runtime, NewNumberValue(float64(idx), false))
		if completion.Type != Normal {
			return completion
		}

		key := completion.Value.(*JavaScriptValue)

		completion = trapResultArray.Value.(ObjectInterface).Get(runtime, key, trapResultArray)
		if completion.Type != Normal {
			return completion
		}

		value := completion.Value.(*JavaScriptValue)

		if value.Type != TypeString && value.Type != TypeSymbol {
			return NewThrowCompletion(NewTypeError(runtime, "Invalid trap result."))
		}

		if KeyListContains(trapResult, value) {
			return NewThrowCompletion(NewTypeError(runtime, "Duplicate key in trap result."))
		}

		trapResult = append(trapResult, value)
	}

	targetObj := target.Value.(ObjectInterface)
	completion = targetObj.IsExtensible(runtime)

	if completion.Type != Normal {
		return completion
	}

	extensibleTarget := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	completion = targetObj.OwnPropertyKeys(runtime)
	if completion.Type != Normal {
		return completion
	}

	targetKeys := completion.Value.([]*JavaScriptValue)

	targetConfigurableKeys := make([]*JavaScriptValue, 0)
	targetNonConfigurableKeys := make([]*JavaScriptValue, 0)

	for _, key := range targetKeys {
		completion = targetObj.GetOwnProperty(runtime, key)
		if completion.Type != Normal {
			return completion
		}

		targetDesc := completion.Value.(PropertyDescriptor)

		if targetDesc != nil && !targetDesc.GetConfigurable() {
			targetNonConfigurableKeys = append(targetNonConfigurableKeys, key)
		} else {
			targetConfigurableKeys = append(targetConfigurableKeys, key)
		}
	}

	if extensibleTarget && len(targetNonConfigurableKeys) == 0 {
		return NewNormalCompletion(trapResult)
	}

	uncheckedResultKeys := make([]*JavaScriptValue, 0)
	copy(uncheckedResultKeys, trapResult)

	for _, key := range targetNonConfigurableKeys {
		if !KeyListContains(uncheckedResultKeys, key) {
			// TODO: Improve error message.
			return NewThrowCompletion(NewTypeError(runtime, "Invalid trap result."))
		}

		uncheckedResultKeys = slices.Delete(uncheckedResultKeys, slices.Index(uncheckedResultKeys, key), 1)
	}

	if extensibleTarget {
		return NewNormalCompletion(trapResult)
	}

	for _, key := range targetConfigurableKeys {
		if !KeyListContains(uncheckedResultKeys, key) {
			// TODO: Improve error message.
			return NewThrowCompletion(NewTypeError(runtime, "Invalid trap result."))
		}

		uncheckedResultKeys = slices.Delete(uncheckedResultKeys, slices.Index(uncheckedResultKeys, key), 1)
	}

	if len(uncheckedResultKeys) > 0 {
		// TODO: Improve error message.
		return NewThrowCompletion(NewTypeError(runtime, "Invalid trap result."))
	}

	return NewNormalCompletion(trapResult)
}

func (o *ProxyObject) PreventExtensions(runtime *Runtime) *Completion {
	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	if handler.Type != TypeObject {
		panic("Assert failed: ProxyHandler is not an Object.")
	}

	completion = GetMethod(runtime, handler, preventExtensionsKey)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return target.Value.(ObjectInterface).PreventExtensions(runtime)
	}

	completion = Call(runtime, trap, handler, []*JavaScriptValue{target})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	completion = ToBoolean(trapResult)
	if completion.Type != Normal {
		return completion
	}

	booleanTrapResult := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if booleanTrapResult {
		completion = target.Value.(ObjectInterface).IsExtensible(runtime)
		if completion.Type != Normal {
			return completion
		}

		extensibleTarget := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

		if extensibleTarget {
			return NewThrowCompletion(NewTypeError(runtime, "Proxy preventExtensions trap returned true but target is extensible."))
		}
	}

	return NewNormalCompletion(NewBooleanValue(booleanTrapResult))
}

func (o *ProxyObject) GetPrivateElements() []*PrivateElement {
	return o.PrivateElements
}

func (o *ProxyObject) SetPrivateElements(privateElements []*PrivateElement) {
	o.PrivateElements = privateElements
}

func (o *ProxyObject) Call(runtime *Runtime, thisArg *JavaScriptValue, arguments []*JavaScriptValue) *Completion {
	if !o.HasCall {
		panic("Assert failed: ProxyObject.Call called on a proxy that does not have the [[Call]] internal method.")
	}

	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, applyStr)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return Call(runtime, target, thisArg, arguments)
	}

	argList := CreateArrayFromList(runtime, arguments)
	argListValue := NewJavaScriptValue(TypeObject, argList)
	return Call(runtime, trap, handler, []*JavaScriptValue{target, thisArg, argListValue})
}

func (o *ProxyObject) Construct(runtime *Runtime, arguments []*JavaScriptValue, newTarget *JavaScriptValue) *Completion {
	if !o.HasConstruct {
		panic("Assert failed: ProxyObject.Construct called on a proxy that does not have the [[Construct]] internal method.")
	}

	completion := ValidateNonRevokedProxy(runtime, o)
	if completion.Type != Normal {
		return completion
	}

	target := o.ProxyTarget
	handler := o.ProxyHandler

	completion = GetMethod(runtime, handler, constructStr)
	if completion.Type != Normal {
		return completion
	}

	trap := completion.Value.(*JavaScriptValue)

	if trap.Type == TypeUndefined {
		return Construct(runtime, target.Value.(FunctionInterface), arguments, newTarget)
	}

	argList := CreateArrayFromList(runtime, arguments)
	argListValue := NewJavaScriptValue(TypeObject, argList)
	completion = Call(runtime, trap, handler, []*JavaScriptValue{target, argListValue, newTarget})
	if completion.Type != Normal {
		return completion
	}

	trapResult := completion.Value.(*JavaScriptValue)

	if trapResult.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Invalid trap result."))
	}

	return NewNormalCompletion(trapResult)
}

func (o *ProxyObject) HasConstructMethod() bool {
	return o.HasConstruct
}

func ValidateNonRevokedProxy(runtime *Runtime, proxy *ProxyObject) *Completion {
	if proxy.ProxyTarget.Type == TypeNull {
		return NewThrowCompletion(NewTypeError(runtime, "Proxy has been revoked."))
	}

	return NewUnusedCompletion()
}

func KeyListContains(keyList []*JavaScriptValue, key *JavaScriptValue) bool {
	for _, k := range keyList {
		completion := SameValue(k, key)
		if completion.Type != Normal {
			panic("Assert failed: KeyListContains SameValue threw an unexpected error.")
		}

		if completion.Value.(*JavaScriptValue).Value.(*Boolean).Value {
			return true
		}
	}

	return false
}
