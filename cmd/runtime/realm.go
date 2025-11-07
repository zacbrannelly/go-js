package runtime

type Realm struct {
	GlobalEnv    *GlobalEnvironment
	GlobalObject ObjectInterface
	// TODO: Other properties.
}

func NewRealm() *Realm {
	// TODO: Initialize the realm according to InitializeHostDefinedRealm in the spec.
	var globalObject *Object = NewEmptyObject()

	// "undefined" property.
	globalObject.DefineOwnProperty(NewStringValue("undefined"), &DataPropertyDescriptor{
		Value:        NewUndefinedValue(),
		Writable:     false,
		Configurable: false,
		Enumerable:   false,
	})

	return &Realm{
		GlobalEnv:    NewGlobalEnvironment(globalObject, globalObject),
		GlobalObject: globalObject,
	}
}
