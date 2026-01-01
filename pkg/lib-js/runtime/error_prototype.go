package runtime

import "fmt"

var (
	messageStr = NewStringValue("message")
	nameStr    = NewStringValue("name")
)

func NewErrorPrototype(runtime *Runtime) ObjectInterface {
	return OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
}

func DefineErrorPrototypeProperties(runtime *Runtime, errorProto ObjectInterface) {
	// Error.prototype.message
	errorProto.DefineOwnProperty(runtime, NewStringValue("message"), &DataPropertyDescriptor{
		Value:        NewStringValue(""),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	// Error.prototype.name
	errorProto.DefineOwnProperty(runtime, NewStringValue("name"), &DataPropertyDescriptor{
		Value:        NewStringValue("Error"),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	// Error.prototype.toString
	DefineBuiltinFunction(runtime, errorProto, "toString", ErrorPrototypeToString, 0)
}

func ErrorPrototypeToString(
	runtime *Runtime,
	function *FunctionObject,
	thisArg *JavaScriptValue,
	arguments []*JavaScriptValue,
	newTarget *JavaScriptValue,
) *Completion {
	completion := ToObject(runtime, thisArg)
	if completion.Type != Normal {
		return completion
	}

	objectVal := completion.Value.(*JavaScriptValue)

	if objectVal.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Error.prototype.toString called on non-object"))
	}

	object := objectVal.Value.(ObjectInterface)

	completion = object.Get(runtime, nameStr, objectVal)
	if completion.Type != Normal {
		return completion
	}

	nameVal := completion.Value.(*JavaScriptValue)

	if nameVal.Type == TypeUndefined {
		nameVal = NewStringValue("Error")
	} else {
		completion = ToString(runtime, nameVal)
		if completion.Type != Normal {
			return completion
		}
		nameVal = completion.Value.(*JavaScriptValue)
	}

	completion = object.Get(runtime, messageStr, objectVal)
	if completion.Type != Normal {
		return completion
	}

	messageVal := completion.Value.(*JavaScriptValue)
	if messageVal.Type == TypeUndefined {
		messageVal = NewStringValue("")
	} else {
		completion = ToString(runtime, messageVal)
		if completion.Type != Normal {
			return completion
		}
		messageVal = completion.Value.(*JavaScriptValue)
	}

	name := nameVal.Value.(*String).Value
	if name == "" {
		return NewNormalCompletion(messageVal)
	}

	message := messageVal.Value.(*String).Value
	if message == "" {
		return NewNormalCompletion(nameVal)
	}

	return NewNormalCompletion(NewStringValue(fmt.Sprintf("%s: %s", name, message)))
}
