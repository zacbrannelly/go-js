package runtime

import "math"

type Intrinsic string

const (
	IntrinsicObjectConstructor            Intrinsic = "Object"
	IntrinsicFunctionConstructor          Intrinsic = "Function"
	IntrinsicArrayConstructor             Intrinsic = "Array"
	IntrinsicStringConstructor            Intrinsic = "String"
	IntrinsicNumberConstructor            Intrinsic = "Number"
	IntrinsicBooleanConstructor           Intrinsic = "Boolean"
	IntrinsicErrorConstructor             Intrinsic = "Error"
	IntrinsicSymbolConstructor            Intrinsic = "Symbol"
	IntrinsicEvalErrorConstructor         Intrinsic = "EvalError"
	IntrinsicRangeErrorConstructor        Intrinsic = "RangeError"
	IntrinsicReferenceErrorConstructor    Intrinsic = "ReferenceError"
	IntrinsicSyntaxErrorConstructor       Intrinsic = "SyntaxError"
	IntrinsicTypeErrorConstructor         Intrinsic = "TypeError"
	IntrinsicURIErrorConstructor          Intrinsic = "URIError"
	IntrinsicMathObject                   Intrinsic = "Math"
	IntrinsicArrayBufferConstructor       Intrinsic = "ArrayBuffer"
	IntrinsicInt8ArrayConstructor         Intrinsic = "Int8Array"
	IntrinsicUint8ArrayConstructor        Intrinsic = "Uint8Array"
	IntrinsicUint8ClampedArrayConstructor Intrinsic = "Uint8ClampedArray"
	IntrinsicInt16ArrayConstructor        Intrinsic = "Int16Array"
	IntrinsicUint16ArrayConstructor       Intrinsic = "Uint16Array"
	IntrinsicInt32ArrayConstructor        Intrinsic = "Int32Array"
	IntrinsicUint32ArrayConstructor       Intrinsic = "Uint32Array"
	IntrinsicBigInt64ArrayConstructor     Intrinsic = "BigInt64Array"
	IntrinsicBigUint64ArrayConstructor    Intrinsic = "BigUint64Array"
	IntrinsicFloat16ArrayConstructor      Intrinsic = "Float16Array"
	IntrinsicFloat32ArrayConstructor      Intrinsic = "Float32Array"
	IntrinsicFloat64ArrayConstructor      Intrinsic = "Float64Array"
	IntrinsicProxyConstructor             Intrinsic = "Proxy"
	IntrinsicObjectPrototype              Intrinsic = "Object.prototype"
	IntrinsicArrayPrototype               Intrinsic = "Array.prototype"
	IntrinsicFunctionPrototype            Intrinsic = "Function.prototype"
	IntrinsicIteratorPrototype            Intrinsic = "Iterator.prototype"
	IntrinsicArrayIteratorPrototype       Intrinsic = "ArrayIterator.prototype"
	IntrinsicStringPrototype              Intrinsic = "String.prototype"
	IntrinsicNumberPrototype              Intrinsic = "Number.prototype"
	IntrinsicBooleanPrototype             Intrinsic = "Boolean.prototype"
	IntrinsicErrorPrototype               Intrinsic = "Error.prototype"
	IntrinsicEvalErrorPrototype           Intrinsic = "EvalError.prototype"
	IntrinsicRangeErrorPrototype          Intrinsic = "RangeError.prototype"
	IntrinsicReferenceErrorPrototype      Intrinsic = "ReferenceError.prototype"
	IntrinsicSyntaxErrorPrototype         Intrinsic = "SyntaxError.prototype"
	IntrinsicTypeErrorPrototype           Intrinsic = "TypeError.prototype"
	IntrinsicURIErrorPrototype            Intrinsic = "URIError.prototype"
	IntrinsicArrayBufferPrototype         Intrinsic = "ArrayBuffer.prototype"
	IntrinsicTypedArrayPrototype          Intrinsic = "TypedArray.prototype"
	IntrinsicInt8ArrayPrototype           Intrinsic = "Int8Array.prototype"
	IntrinsicUint8ArrayPrototype          Intrinsic = "Uint8Array.prototype"
	IntrinsicUint8ClampedArrayPrototype   Intrinsic = "Uint8ClampedArray.prototype"
	IntrinsicInt16ArrayPrototype          Intrinsic = "Int16Array.prototype"
	IntrinsicUint16ArrayPrototype         Intrinsic = "Uint16Array.prototype"
	IntrinsicInt32ArrayPrototype          Intrinsic = "Int32Array.prototype"
	IntrinsicUint32ArrayPrototype         Intrinsic = "Uint32Array.prototype"
	IntrinsicBigInt64ArrayPrototype       Intrinsic = "BigInt64Array.prototype"
	IntrinsicBigUint64ArrayPrototype      Intrinsic = "BigUint64Array.prototype"
	IntrinsicFloat16ArrayPrototype        Intrinsic = "Float16Array.prototype"
	IntrinsicFloat32ArrayPrototype        Intrinsic = "Float32Array.prototype"
	IntrinsicFloat64ArrayPrototype        Intrinsic = "Float64Array.prototype"
	IntrinsicParseIntFunction             Intrinsic = "parseInt"
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

	// "globalThis" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("globalThis"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, globalObject),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "undefined" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("undefined"), &DataPropertyDescriptor{
		Value:        NewUndefinedValue(),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "NaN" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("NaN"), &DataPropertyDescriptor{
		Value:        NewNumberValue(math.NaN(), true),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "parseInt" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("parseInt"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicParseIntFunction)),
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

	// "String" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("String"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicStringConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Number" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Number"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicNumberConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Boolean" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Boolean"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicBooleanConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Symbol" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Symbol"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicSymbolConstructor)),
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

	// "Math" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Math"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicMathObject)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "ArrayBuffer" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("ArrayBuffer"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicArrayBufferConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Int8Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Int8Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicInt8ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Uint8Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Uint8Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicUint8ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Uint8ClampedArray" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Uint8ClampedArray"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicUint8ClampedArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Int16Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Int16Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicInt16ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Uint16Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Uint16Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicUint16ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Int32Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Int32Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicInt32ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Uint32Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Uint32Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicUint32ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "BigInt64Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("BigInt64Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicBigInt64ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "BigUint64Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("BigUint64Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicBigUint64ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Float16Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Float16Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicFloat16ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Float32Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Float32Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicFloat32ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Float64Array" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Float64Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicFloat64ArrayConstructor)),
		Writable:     true,
		Configurable: true,
		Enumerable:   false,
	})

	// "Proxy" property.
	globalObject.DefineOwnProperty(runtime, NewStringValue("Proxy"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicProxyConstructor)),
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
	r.Intrinsics[IntrinsicStringPrototype] = NewStringPrototype(runtime)
	r.Intrinsics[IntrinsicNumberPrototype] = NewNumberPrototype(runtime)
	r.Intrinsics[IntrinsicBooleanPrototype] = NewBooleanPrototype(runtime)
	r.Intrinsics[IntrinsicArrayIteratorPrototype] = NewArrayIteratorPrototype(runtime)
	r.Intrinsics[IntrinsicErrorPrototype] = NewErrorPrototype(runtime)
	r.Intrinsics[IntrinsicEvalErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicRangeErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicReferenceErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicSyntaxErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicTypeErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicURIErrorPrototype] = NewNativeErrorPrototype(runtime)
	r.Intrinsics[IntrinsicArrayBufferPrototype] = NewArrayBufferPrototype(runtime)
	r.Intrinsics[IntrinsicTypedArrayPrototype] = NewTypedArrayPrototype(runtime)
	r.Intrinsics[IntrinsicInt8ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameInt8)
	r.Intrinsics[IntrinsicUint8ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameUint8)
	r.Intrinsics[IntrinsicUint8ClampedArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameUint8Clamped)
	r.Intrinsics[IntrinsicInt16ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameInt16)
	r.Intrinsics[IntrinsicUint16ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameUint16)
	r.Intrinsics[IntrinsicInt32ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameInt32)
	r.Intrinsics[IntrinsicUint32ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameUint32)
	r.Intrinsics[IntrinsicBigInt64ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameBigInt64)
	r.Intrinsics[IntrinsicBigUint64ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameBigUint64)
	r.Intrinsics[IntrinsicFloat16ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameFloat16)
	r.Intrinsics[IntrinsicFloat32ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameFloat32)
	r.Intrinsics[IntrinsicFloat64ArrayPrototype] = NewConcreteTypedArrayPrototype(runtime, TypedArrayNameFloat64)

	// Intrinsic Constructors.
	r.Intrinsics[IntrinsicObjectConstructor] = NewObjectConstructor(runtime)
	r.Intrinsics[IntrinsicFunctionConstructor] = NewFunctionConstructor(runtime)
	r.Intrinsics[IntrinsicArrayConstructor] = NewArrayConstructor(runtime)
	r.Intrinsics[IntrinsicStringConstructor] = NewStringConstructor(runtime)
	r.Intrinsics[IntrinsicNumberConstructor] = NewNumberConstructor(runtime)
	r.Intrinsics[IntrinsicBooleanConstructor] = NewBooleanConstructor(runtime)
	r.Intrinsics[IntrinsicSymbolConstructor] = NewSymbolConstructor(runtime)
	r.Intrinsics[IntrinsicErrorConstructor] = NewErrorConstructor(runtime)
	r.Intrinsics[IntrinsicEvalErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeEvalError, IntrinsicEvalErrorPrototype)
	r.Intrinsics[IntrinsicRangeErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeRangeError, IntrinsicRangeErrorPrototype)
	r.Intrinsics[IntrinsicReferenceErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeReferenceError, IntrinsicReferenceErrorPrototype)
	r.Intrinsics[IntrinsicSyntaxErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeSyntaxError, IntrinsicSyntaxErrorPrototype)
	r.Intrinsics[IntrinsicTypeErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeTypeError, IntrinsicTypeErrorPrototype)
	r.Intrinsics[IntrinsicURIErrorConstructor] = NewNativeErrorConstructor(runtime, NativeErrorTypeURIError, IntrinsicURIErrorPrototype)
	r.Intrinsics[IntrinsicArrayBufferConstructor] = NewArrayBufferConstructor(runtime)
	r.Intrinsics[IntrinsicInt8ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameInt8, IntrinsicInt8ArrayPrototype)
	r.Intrinsics[IntrinsicUint8ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameUint8, IntrinsicUint8ArrayPrototype)
	r.Intrinsics[IntrinsicUint8ClampedArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameUint8Clamped, IntrinsicUint8ClampedArrayPrototype)
	r.Intrinsics[IntrinsicInt16ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameInt16, IntrinsicInt16ArrayPrototype)
	r.Intrinsics[IntrinsicUint16ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameUint16, IntrinsicUint16ArrayPrototype)
	r.Intrinsics[IntrinsicInt32ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameInt32, IntrinsicInt32ArrayPrototype)
	r.Intrinsics[IntrinsicUint32ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameUint32, IntrinsicUint32ArrayPrototype)
	r.Intrinsics[IntrinsicBigInt64ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameBigInt64, IntrinsicBigInt64ArrayPrototype)
	r.Intrinsics[IntrinsicBigUint64ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameBigUint64, IntrinsicBigUint64ArrayPrototype)
	r.Intrinsics[IntrinsicFloat16ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameFloat16, IntrinsicFloat16ArrayPrototype)
	r.Intrinsics[IntrinsicFloat32ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameFloat32, IntrinsicFloat32ArrayPrototype)
	r.Intrinsics[IntrinsicFloat64ArrayConstructor] = NewTypedArrayConstructor(runtime, TypedArrayNameFloat64, IntrinsicFloat64ArrayPrototype)
	r.Intrinsics[IntrinsicProxyConstructor] = NewProxyObjectConstructor(runtime)

	// Intrinsic Objects.
	r.Intrinsics[IntrinsicMathObject] = NewMathObject(runtime)
	r.Intrinsics[IntrinsicParseIntFunction] = NewParseIntFunction(runtime)

	// Define properties on the prototypes.
	DefineObjectPrototypeProperties(runtime, r.Intrinsics[IntrinsicObjectPrototype].(*ObjectPrototype))
	DefineArrayPrototypeProperties(runtime, r.Intrinsics[IntrinsicArrayPrototype].(*ArrayObject))
	DefineFunctionPrototypeProperties(runtime, r.Intrinsics[IntrinsicFunctionPrototype])
	DefineIteratorPrototypeProperties(runtime, r.Intrinsics[IntrinsicIteratorPrototype])
	DefineArrayIteratorPrototypeProperties(runtime, r.Intrinsics[IntrinsicArrayIteratorPrototype])
	DefineStringPrototypeProperties(runtime, r.Intrinsics[IntrinsicStringPrototype])
	DefineNumberPrototypeProperties(runtime, r.Intrinsics[IntrinsicNumberPrototype])
	DefineBooleanPrototypeProperties(runtime, r.Intrinsics[IntrinsicBooleanPrototype])
	DefineErrorPrototypeProperties(runtime, r.Intrinsics[IntrinsicErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeSyntaxError, r.Intrinsics[IntrinsicSyntaxErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeTypeError, r.Intrinsics[IntrinsicTypeErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeReferenceError, r.Intrinsics[IntrinsicReferenceErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeRangeError, r.Intrinsics[IntrinsicRangeErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeURIError, r.Intrinsics[IntrinsicURIErrorPrototype])
	DefineNativeErrorPrototypeProperties(runtime, NativeErrorTypeEvalError, r.Intrinsics[IntrinsicEvalErrorPrototype])
	DefineNumberConstructorProperties(runtime, r.Intrinsics[IntrinsicNumberConstructor])
	DefineArrayBufferPrototypeProperties(runtime, r.Intrinsics[IntrinsicArrayBufferPrototype])
	DefineTypedArrayPrototypeProperties(runtime, r.Intrinsics[IntrinsicTypedArrayPrototype])

	// Set constructors to the prototypes (needs to be done after both the constructors and the prototypes are created).
	SetConstructor(runtime, r.Intrinsics[IntrinsicObjectPrototype], r.Intrinsics[IntrinsicObjectConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicArrayPrototype], r.Intrinsics[IntrinsicArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicFunctionPrototype], r.Intrinsics[IntrinsicFunctionConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicStringPrototype], r.Intrinsics[IntrinsicStringConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicNumberPrototype], r.Intrinsics[IntrinsicNumberConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicBooleanPrototype], r.Intrinsics[IntrinsicBooleanConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicErrorPrototype], r.Intrinsics[IntrinsicErrorConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicEvalErrorPrototype], r.Intrinsics[IntrinsicEvalErrorConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicRangeErrorPrototype], r.Intrinsics[IntrinsicRangeErrorConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicReferenceErrorPrototype], r.Intrinsics[IntrinsicReferenceErrorConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicSyntaxErrorPrototype], r.Intrinsics[IntrinsicSyntaxErrorConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicTypeErrorPrototype], r.Intrinsics[IntrinsicTypeErrorConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicURIErrorPrototype], r.Intrinsics[IntrinsicURIErrorConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicArrayBufferPrototype], r.Intrinsics[IntrinsicArrayBufferConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicInt8ArrayPrototype], r.Intrinsics[IntrinsicInt8ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicUint8ArrayPrototype], r.Intrinsics[IntrinsicUint8ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicUint8ClampedArrayPrototype], r.Intrinsics[IntrinsicUint8ClampedArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicInt16ArrayPrototype], r.Intrinsics[IntrinsicInt16ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicUint16ArrayPrototype], r.Intrinsics[IntrinsicUint16ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicInt32ArrayPrototype], r.Intrinsics[IntrinsicInt32ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicUint32ArrayPrototype], r.Intrinsics[IntrinsicUint32ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicBigInt64ArrayPrototype], r.Intrinsics[IntrinsicBigInt64ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicBigUint64ArrayPrototype], r.Intrinsics[IntrinsicBigUint64ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicFloat16ArrayPrototype], r.Intrinsics[IntrinsicFloat16ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicFloat32ArrayPrototype], r.Intrinsics[IntrinsicFloat32ArrayConstructor].(FunctionInterface))
	SetConstructor(runtime, r.Intrinsics[IntrinsicFloat64ArrayPrototype], r.Intrinsics[IntrinsicFloat64ArrayConstructor].(FunctionInterface))

	// TODO: Create other intrinsics.
}
