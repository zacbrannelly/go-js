package runtime

func NewIteratorPrototype(runtime *Runtime) ObjectInterface {
	return OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
}

func DefineIteratorPrototypeProperties(runtime *Runtime, obj ObjectInterface) {
	// TODO: Define properties.
}
