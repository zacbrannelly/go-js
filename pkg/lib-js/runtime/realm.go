package runtime

import "math"

type Intrinsic string

const (
	IntrinsicObjectConstructor         Intrinsic = "Object"
	IntrinsicFunctionConstructor       Intrinsic = "Function"
	IntrinsicArrayConstructor          Intrinsic = "Array"
	IntrinsicErrorConstructor          Intrinsic = "Error"
	IntrinsicEvalErrorConstructor      Intrinsic = "EvalError"
	IntrinsicRangeErrorConstructor     Intrinsic = "RangeError"
	IntrinsicReferenceErrorConstructor Intrinsic = "ReferenceError"
	IntrinsicSyntaxErrorConstructor    Intrinsic = "SyntaxError"
	IntrinsicTypeErrorConstructor      Intrinsic = "TypeError"
	IntrinsicURIErrorConstructor       Intrinsic = "URIError"
	IntrinsicObjectPrototype           Intrinsic = "Object.prototype"
	IntrinsicArrayPrototype            Intrinsic = "Array.prototype"
	IntrinsicFunctionPrototype         Intrinsic = "Function.prototype"
	IntrinsicIteratorPrototype         Intrinsic = "Iterator.prototype"
	IntrinsicArrayIteratorPrototype    Intrinsic = "ArrayIterator.prototype"
	IntrinsicErrorPrototype            Intrinsic = "Error.prototype"
	IntrinsicEvalErrorPrototype        Intrinsic = "EvalError.prototype"
	IntrinsicRangeErrorPrototype       Intrinsic = "RangeError.prototype"
	IntrinsicReferenceErrorPrototype   Intrinsic = "ReferenceError.prototype"
	IntrinsicSyntaxErrorPrototype      Intrinsic = "SyntaxError.prototype"
	IntrinsicTypeErrorPrototype        Intrinsic = "TypeError.prototype"
	IntrinsicURIErrorPrototype         Intrinsic = "URIError.prototype"
)

type Realm struct {
	GlobalEnv    *GlobalEnvironment
	GlobalObject ObjectInterface
	Intrinsics   map[Intrinsic]ObjectInterface
	// TODO: Other properties.
}

func NewRealm(runtime *Runtime) *Realm {
	// TODO: Initialize the realm according to InitializeHostDefinedRealm in the spec.
	var globalObject *Object = NewEmptyObject()

	realm := &Realm{
		GlobalEnv:    NewGlobalEnvironment(globalObject, globalObject),
		GlobalObject: globalObject,
		Intrinsics:   make(map[Intrinsic]ObjectInterface),
	}

	// An execution context with the new realm is required before creating the intrinsics.
	runtime.PushExecutionContext(&ExecutionContext{
		Realm: realm,
	})
	realm.CreateIntrinsics(runtime)

	// "undefined" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("undefined"), &DataPropertyDescriptor{
		Value:        NewUndefinedValue(),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Object" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Object"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicObjectConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Function" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Function"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicFunctionConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicArrayConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Error" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Error"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "EvalError" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("EvalError"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicEvalErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "RangeError" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("RangeError"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicRangeErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "ReferenceError" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("ReferenceError"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicReferenceErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "SyntaxError" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("SyntaxError"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicSyntaxErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "TypeError" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("TypeError"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicTypeErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "URIError" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("URIError"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicURIErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Infinity" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Infinity"), &DataPropertyDescriptor{
		Value:        NewNumberValue(math.Inf(1), false),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "console" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("console"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, NewConsoleObject(runtime)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	return realm
}

func (r *Realm) GetIntrinsic(intrinsic Intrinsic) ObjectInterface {
	if intrinsic, ok := r.Intrinsics[intrinsic]; ok {
		return intrinsic
	}

	panic("Assert failed: Intrinsic not found in realm.")
}

func (r *Realm) CreateIntrinsics(runtime *Runtime) {
	// Intrinsic Prototypes.
	r.Intrinsics[IntrinsicObjectPrototype] = NewObjectPrototype(runtime)
	r.Intrinsics[IntrinsicArrayPrototype] = NewArrayPrototype(runtime)
	r.Intrinsics[IntrinsicFunctionPrototype] = NewFunctionPrototype(runtime)
	r.Intrinsics[IntrinsicIteratorPrototype] = NewIteratorPrototype(runtime)
	r.Intrinsics[IntrinsicArrayIteratorPrototype] = NewArrayIteratorPrototype(runtime)
	r.Intrinsics[IntrinsicErrorPrototype] = NewErrorPrototype(runtime)
	r.Intrinsics[IntrinsicEvalErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicRangeErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicReferenceErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicSyntaxErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicTypeErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicURIErrorPrototype] = NewNativeErrorPrototype(runtime)

	// Intrinsic Constructors.
	r.Intrinsics[IntrinsicObjectConstructor] = NewObjectConstructor(runtime)
	r.Intrinsics[IntrinsicFunctionConstructor] = NewFunctionConstructor(runtime)
	r.Intrinsics[IntrinsicArrayConstructor] = NewArrayConstructor(runtime)
	r.Intrinsics[IntrinsicErrorConstructor] = NewErrorConstructor(runtime)
	r.Intrinsics[IntrinsicEvalErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeEvalError, IntrinsicEvalErrorPrototype)
	r.Intrinsics[IntrinsicRangeErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeRangeError, IntrinsicRangeErrorPrototype)
	r.Intrinsics[IntrinsicReferenceErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeReferenceError, IntrinsicReferenceErrorPrototype)
	r.Intrinsics[IntrinsicSyntaxErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeSyntaxError, IntrinsicSyntaxErrorPrototype)
	r.Intrinsics[IntrinsicTypeErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeTypeError, IntrinsicTypeErrorPrototype)
	r.Intrinsics[IntrinsicURIErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeURIError, IntrinsicURIErrorPrototype)

	// Define properties on the prototypes.
	DefineObjectPrototypeProperties(runtime, r.Intrinsics[IntrinsicObjectPrototype].(*ObjectPrototype))
	DefineArrayPrototypeProperties(runtime, r.Intrinsics[IntrinsicArrayPrototype].(*ArrayObject))
	DefineFunctionPrototypeProperties(runtime, r.Intrinsics[IntrinsicFunctionPrototype].(*FunctionObject))
	DefineIteratorPrototypeProperties(runtime, r.Intrinsics[IntrinsicIteratorPrototype])
	DefineArrayIteratorPrototypeProperties(runtime, r.Intrinsics[IntrinsicArrayIteratorPrototype])
	DefineErrorPrototypeProperties(runtime, r.Intrinsics[IntrinsicErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeSyntaxError, r.Intrinsics[IntrinsicSyntaxErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeTypeError, r.Intrinsics[IntrinsicTypeErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeReferenceError, r.Intrinsics[IntrinsicReferenceErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeRangeError, r.Intrinsics[IntrinsicRangeErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeURIError, r.Intrinsics[IntrinsicURIErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeEvalError, r.Intrinsics[IntrinsicEvalErrorPrototype])

	// Set constructors to the prototypes (needs to be done after both the constructors and the prototypes are created).
	SetConstructor(runtime, r.Intrinsics[IntrinsicObjectPrototype], r.Intrinsics[IntrinsicObjectConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicArrayPrototype], r.Intrinsics[IntrinsicArrayConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicFunctionPrototype], r.Intrinsics[IntrinsicFunctionConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicErrorPrototype], r.Intrinsics[IntrinsicErrorConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicEvalErrorPrototype], r.Intrinsics[IntrinsicEvalErrorConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicRangeErrorPrototype], r.Intrinsics[IntrinsicRangeErrorConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicReferenceErrorPrototype], r.Intrinsics[IntrinsicReferenceErrorConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicSyntaxErrorPrototype], r.Intrinsics[IntrinsicSyntaxErrorConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicTypeErrorPrototype], r.Intrinsics[IntrinsicTypeErrorConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicURIErrorPrototype], r.Intrinsics[IntrinsicURIErrorConstructor].(*FunctionObject))

	// TODO: Create other intrinsics.
}
