package runtime

type Realm struct {
	GlobalEnv    *GlobalEnvironment
	GlobalObject *Object
	// TODO: Other properties.
}

func NewRealm() *Realm {
	// TODO: Initialize the realm according to InitializeHostDefinedRealm in the spec.
	var globalObject *Object = NewEmptyObject()

	// "undefined" property.
	globalObject.Set(
		NewStringValue("undefined"),
		NewUndefinedValue(),
		NewJavaScriptValue(TypeObject, globalObject),
	)

	return &Realm{
		GlobalEnv:    NewGlobalEnvironment(globalObject, globalObject),
		GlobalObject: globalObject,
	}
}
