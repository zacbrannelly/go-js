package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func EvaluateInitializer(runtime *Runtime, initializer *ast.BasicNode) *Completion {
	if len(initializer.GetChildren()) == 0 {
		panic("Assert failed: Initializer node has no children.")
	}

	return Evaluate(runtime, initializer.GetChildren()[0])
}
