package runtime

type Intrinsic string

const (
	IntrinsicObjectConstructor Intrinsic = "Object"
	IntrinsicObjectPrototype   Intrinsic = "Object.prototype"
	IntrinsicArrayPrototype    Intrinsic = "Array.prototype"
	IntrinsicFunctionPrototype Intrinsic = "Function.prototype"
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
		Value:        NewJavaScriptValue(TypeObject, realm.Intrinsics[IntrinsicObjectConstructor]),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	return realm
}

func (r *Realm) CreateIntrinsics(runtime *Runtime) {
	// Intrinsic Prototypes.
	r.Intrinsics[IntrinsicObjectPrototype] = NewObjectPrototype(runtime)
	r.Intrinsics[IntrinsicArrayPrototype] = NewArrayPrototype(runtime)
	r.Intrinsics[IntrinsicFunctionPrototype] = NewFunctionPrototype(runtime)

	// Intrinsic Constructors.
	r.Intrinsics[IntrinsicObjectConstructor] = NewObjectConstructor(runtime)

	// TODO: Create other intrinsics.
}
