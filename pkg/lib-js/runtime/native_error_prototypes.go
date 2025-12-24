package runtime

type NativeErrorType string

const (
	NativeErrorTypeSyntaxError    NativeErrorType = "SyntaxError"
	NativeErrorTypeTypeError      NativeErrorType = "TypeError"
	NativeErrorTypeReferenceError NativeErrorType = "ReferenceError"
	NativeErrorTypeRangeError     NativeErrorType = "RangeError"
	NativeErrorTypeURIError       NativeErrorType = "URIError"
	NativeErrorTypeEvalError      NativeErrorType = "EvalError"
)

func NewNativeErrorPrototype(runtime *Runtime) ObjectInterface {
	return OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicErrorPrototype))
}

func DefineNativeErrorPrototypeProperties(runtime *Runtime, errorType NativeErrorType, errorProto ObjectInterface) {
	// Error.prototype.message
	errorProto.DefineOwnProperty(runtime, NewStringValue("message"), &DataPropertyDescriptor{
		Value:        NewStringValue(""),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})

	// Error.prototype.name
	errorProto.DefineOwnProperty(runtime, NewStringValue("name"), &DataPropertyDescriptor{
		Value:        NewStringValue(string(errorType)),
		Writable:     true,
		Enumerable:   false,
		Configurable: true,
	})
}
