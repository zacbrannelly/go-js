package runtime

type Realm struct {
	GlobalEnv *GlobalEnvironment
	// TODO: Other properties.
}

func NewRealm() *Realm {
	// TODO: Initialize the realm according to InitializeHostDefinedRealm.
	return &Realm{
		GlobalEnv: &GlobalEnvironment{},
	}
}
