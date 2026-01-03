package runtime

func NewNumberPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
	prototype.(*Object).NumberData = NewNumberValue(0, false)

	return prototype
}

func DefineNumberPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// TODO: Define other properties.
}
