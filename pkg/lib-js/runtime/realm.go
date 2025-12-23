package runtime

import "math"

type Intrinsic string

const (
	IntrinsicObjectConstructor      Intrinsic = "Object"
	IntrinsicFunctionConstructor    Intrinsic = "Function"
	IntrinsicArrayConstructor       Intrinsic = "Array"
	IntrinsicErrorConstructor       Intrinsic = "Error"
	IntrinsicObjectPrototype        Intrinsic = "Object.prototype"
	IntrinsicArrayPrototype         Intrinsic = "Array.prototype"
	IntrinsicFunctionPrototype      Intrinsic = "Function.prototype"
	IntrinsicIteratorPrototype      Intrinsic = "Iterator.prototype"
	IntrinsicArrayIteratorPrototype Intrinsic = "ArrayIterator.prototype"
	IntrinsicErrorPrototype         Intrinsic = "Error.prototype"
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
	globalObject.DefineOwnProperty(NewStringValue("undefined"), &DataPropertyDescriptor{
		Value:        NewUndefinedValue(),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Object" property.
	globalObject.DefineOwnProperty(NewStringValue("Object"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicObjectConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Function" property.
	globalObject.DefineOwnProperty(NewStringValue("Function"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicFunctionConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Array" property.
	globalObject.DefineOwnProperty(NewStringValue("Array"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicArrayConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Error" property.
	globalObject.DefineOwnProperty(NewStringValue("Error"), &DataPropertyDescriptor{
		Value:        NewJavaScriptValue(TypeObject, realm.GetIntrinsic(IntrinsicErrorConstructor)),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	// "Infinity" property.
	globalObject.DefineOwnProperty(NewStringValue("Infinity"), &DataPropertyDescriptor{
		Value:        NewNumberValue(math.Inf(1), false),
		Writable:     false,
		Configurable: false,
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

	// Intrinsic Constructors.
	r.Intrinsics[IntrinsicObjectConstructor] = NewObjectConstructor(runtime)
	r.Intrinsics[IntrinsicFunctionConstructor] = NewFunctionConstructor(runtime)
	r.Intrinsics[IntrinsicArrayConstructor] = NewArrayConstructor(runtime)
	r.Intrinsics[IntrinsicErrorConstructor] = NewErrorConstructor(runtime)

	// Define properties on the prototypes.
	DefineObjectPrototypeProperties(runtime, r.Intrinsics[IntrinsicObjectPrototype].(*ObjectPrototype))
	DefineArrayPrototypeProperties(runtime, r.Intrinsics[IntrinsicArrayPrototype].(*ArrayObject))
	DefineFunctionPrototypeProperties(runtime, r.Intrinsics[IntrinsicFunctionPrototype].(*FunctionObject))
	DefineIteratorPrototypeProperties(runtime, r.Intrinsics[IntrinsicIteratorPrototype])
	DefineArrayIteratorPrototypeProperties(runtime, r.Intrinsics[IntrinsicArrayIteratorPrototype])
	DefineErrorPrototypeProperties(runtime, r.Intrinsics[IntrinsicErrorPrototype])

	// Set constructors to the prototypes (needs to be done after both the constructors and the prototypes are created).
	SetConstructor(runtime, r.Intrinsics[IntrinsicObjectPrototype], r.Intrinsics[IntrinsicObjectConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicArrayPrototype], r.Intrinsics[IntrinsicArrayConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicFunctionPrototype], r.Intrinsics[IntrinsicFunctionConstructor].(*FunctionObject))
	SetConstructor(runtime, r.Intrinsics[IntrinsicErrorPrototype], r.Intrinsics[IntrinsicErrorConstructor].(*FunctionObject))

	// TODO: Create other intrinsics.
}
