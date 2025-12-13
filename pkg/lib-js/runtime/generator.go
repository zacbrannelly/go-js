package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func GeneratorStartWithFunction(runtime *Runtime, generator *Object, functionBody ast.Node) {
	closure := func() *Completion {
		return Evaluate(runtime, functionBody)
	}
	GeneratorStartWithClosure(runtime, generator, closure)
}

func GeneratorStartWithClosure(runtime *Runtime, generator *Object, closure IteratorClosure) {
	if generator.GeneratorState != GeneratorStateSuspendedStart {
		panic("Assert failed: Generator is not in the suspended start state.")
	}

	// startClosure := func() *Completion {
	// 	panic("TODO: Implement GeneratorStart closure.")
	// }

	// TODO: Some how get the startClosure into the generator's execution context, so that is called when the generator is started.

	generator.GeneratorContext = runtime.GetRunningExecutionContext()
}

func GeneratorResume(runtime *Runtime, generator *Object, value *JavaScriptValue, generatorBrand string) *Completion {
	panic("TODO: Implement GeneratorResume.")
}
