package runtime

func NewBooleanPrototype(runtime *Runtime) ObjectInterface {
	prototype := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
	prototype.(*Object).BooleanData = NewBooleanValue(false)

	return prototype
}

func DefineBooleanPrototypeProperties(runtime *Runtime, prototype ObjectInterface) {
	// TODO: Define other properties.
}
