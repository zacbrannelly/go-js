package runtime

type Intrinsic string

const (
	IntrinsicArrayPrototype Intrinsic = "Array.prototype"
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

	// "undefined" property.
	globalObject.DefineOwnProperty(NewStringValue("undefined"), &DataPropertyDescriptor{
		Value:        NewUndefinedValue(),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

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

	return realm
}

func (r *Realm) CreateIntrinsics(runtime *Runtime) {
	r.Intrinsics[IntrinsicArrayPrototype] = NewArrayPrototype(runtime)

	// TODO: Create other intrinsics.
}
