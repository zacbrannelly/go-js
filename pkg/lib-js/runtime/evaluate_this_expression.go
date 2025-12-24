package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateThisExpression(runtime *Runtime, thisExpression *ast.BasicNode) *Completion {
	return ResolveThisBinding(runtime)
}

func ResolveThisBinding(runtime *Runtime) *Completion {
	env := GetThisEnvironment(runtime)
	return env.GetThisBinding(runtime)
}

func GetThisEnvironment(runtime *Runtime) Environment {
	env := runtime.GetRunningExecutionContext().LexicalEnvironment

	for env != nil {
		if env.HasThisBinding() {
			return env
		}

		env = env.GetOuterEnvironment()
	}

	panic("Assert failed: This should never happen.")
}
