package runtime

func NewIteratorPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().Intrinsics[IntrinsicObjectPrototype])

	// TODO: Define properties.

	return prototype
}
