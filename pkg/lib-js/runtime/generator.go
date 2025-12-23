package runtime

import "zbrannelly.dev/go-js/pkg/lib-js/parser/ast"

func GeneratorStartWithFunction(runtime *Runtime, generator *Object, functionBody ast.Node) {
	instructions := Compile(runtime, functionBody)
	GeneratorStartWithClosure(runtime, generator, instructions)
}

func GeneratorStartWithClosure(runtime *Runtime, generator *Object, closureInstructions []Instruction) {
	if generator.GeneratorState != GeneratorStateSuspendedStart {
		panic("Assert failed: Generator is not in the suspended start state.")
	}

	executionContext := runtime.GetRunningExecutionContext()
	executionContext.Generator = generator

	// Epilogue of the closure. Called after the closure returns, throws an exception, or completes normally.
	closureEpilogue := EmitEvaluateNativeCallback(func(runtime *Runtime, vm *ExecutionVM) *Completion {
		// TODO: Assert that the closure threw an exception or had an explicit/implicit return.

		// Pop the result of the closure off the stack.
		result := vm.PopEvaluationStack()

		// Pop the generator context off the execution context stack.
		generatorContext := runtime.PopExecutionContext()

		// Mark the generator as completed.
		generatorContext.Generator.GeneratorState = GeneratorStateCompleted

		var resultValue *JavaScriptValue
		switch result.Type {
		case Normal:
			resultValue = NewUndefinedValue()
		case Return:
			resultValue = result.Value.(*JavaScriptValue)
		case Throw:
			return result
		default:
			panic("Assert failed: Invalid result type in GeneratorStart closure.")
		}

		resultObj := CreateIteratorResultObject(runtime, resultValue, true)
		return NewNormalCompletion(resultObj)
	})
	executionContext.VM.Instructions = append(closureInstructions, closureEpilogue)

	generator.GeneratorContext = executionContext
}

func GeneratorResume(runtime *Runtime, generator *Object, value *JavaScriptValue, generatorBrand string) *Completion {
	completion := GeneratorValidate(generator, generatorBrand)
	if completion.Type != Normal {
		return completion
	}

	state := completion.Value.(GeneratorState)

	if state == GeneratorStateCompleted {
		return NewNormalCompletion(CreateIteratorResultObject(runtime, NewUndefinedValue(), true))
	}

	if state != GeneratorStateSuspendedStart && state != GeneratorStateSuspendedYield {
		panic("Assert failed: Generator is not in suspended state.")
	}

	methodContext := runtime.GetRunningExecutionContext()

	// Resume the generator.
	generator.GeneratorState = GeneratorStateExecuting
	runtime.PushExecutionContext(generator.GeneratorContext)

	// TODO: Push `value` onto the generator's evaluation stack?

	// Run the generator VM (until it suspends or completes).
	completion = ExecuteVM(runtime, generator.GeneratorContext.VM)

	if methodContext != runtime.GetRunningExecutionContext() {
		panic("Assert failed: GeneratorResume returned to the wrong execution context.")
	}

	return completion
}

func GeneratorValidate(generator *Object, generatorBrand string) *Completion {
	if generator.GeneratorBrand != generatorBrand {
		return NewThrowCompletion(NewTypeError("Generator brand mismatch."))
	}

	if generator.GeneratorState == GeneratorStateExecuting {
		return NewThrowCompletion(NewTypeError("Generator is already executing."))
	}

	return NewNormalCompletion(generator.GeneratorState)
}

func CreateIteratorResultObject(runtime *Runtime, value *JavaScriptValue, done bool) *JavaScriptValue {
	obj := OrdinaryObjectCreate(runtime.GetRunningRealm().GetIntrinsic(IntrinsicObjectPrototype))
	CreateDataProperty(obj, NewStringValue("value"), value)
	CreateDataProperty(obj, NewStringValue("done"), NewBooleanValue(done))
	return NewJavaScriptValue(TypeObject, obj)
}
