package runtime

func NewArrayBufferPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
	return prototype
}

func DefineArrayBufferPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// ArrayBuffer.prototype.resize
	DefineBuiltinFunction(runtime, prototype, "resize", ArrayBufferPrototypeResize, 1)

	// TODO: Define other properties.
}

func ArrayBufferPrototypeResize(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	if len(arguments) == 0 {
		arguments = append(arguments, NewUndefinedValue())
	}

	if thisArg.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot call method resize on a non-object."))
	}

	obj, ok := thisArg.Value.(*Object)
	if !ok {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot call method resize on a non-ArrayBuffer object."))
	}

	if !obj.ArrayBufferHasMaxByteLength {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot call method resize on a non-resizable ArrayBuffer object."))
	}

	if IsSharedArrayBuffer(obj) {
		return NewThrowCompletion(NewTypeError(runtime, "Cannot call method resize on a SharedArrayBuffer object."))
	}

	length := arguments[0]
	completion := ToIndex(runtime, length)
	if completion.Type != Normal {
		return completion
	}

	newByteLength := uint(completion.Value.(*JavaScriptValue).Value.(*Number).Value)

	if IsDetachedArrayBuffer(obj) {
		return NewThrowCompletion(NewTypeError(runtime, "ArrayBuffer is detached."))
	}

	if newByteLength > obj.ArrayBufferMaxByteLength {
		return NewThrowCompletion(NewRangeError(runtime, "New length exceeds the maximum byte length."))
	}

	// Resize the array buffer.
	obj.ArrayBufferByteLength = newByteLength
	newData := make([]byte, newByteLength)
	copy(newData, obj.ArrayBufferData)
	obj.ArrayBufferData = newData

	return NewNormalCompletion(NewUndefinedValue())
}
