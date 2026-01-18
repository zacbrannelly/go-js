package runtime

func NewBigIntPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))

	return prototype
}

func DefineBigIntPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// TODO: Define other properties.
}
