package runtime

type ExecutionContext struct {
	Realm    *Realm
	Function *Function
	Script   *Script
	// TODO: Store module record.

	// Points to the environments that can resolve identifier references.
	LexicalEnvironment  Environment
	VariableEnvironment Environment
	PrivateEnvironment  Environment
}
