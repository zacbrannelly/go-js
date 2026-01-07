package runtime

func NewArrayBufferPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
	return prototype
}

func DefineArrayBufferPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// TODO: Define other properties.
}
