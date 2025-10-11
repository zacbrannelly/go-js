package runtime

type Realm struct {
	GlobalEnv *GlobalEnvironment
	// TODO: Other properties.
}

func NewRealm() *Realm {
	// TODO: Initialize the realm according to InitializeHostDefinedRealm in the spec.
	var globalObject *Object = NewEmptyObject()
	return &Realm{
		GlobalEnv: NewGlobalEnvironment(globalObject, globalObject),
	}
}
