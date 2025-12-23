package runtime

type ArrayIteratorKind int

const (
	ArrayIteratorKindKey ArrayIteratorKind = iota
	ArrayIteratorKindValue
	ArrayIteratorKindEntry
)

func CreateArrayIterator(runtime *Runtime, array ObjectInterface, kind ArrayIteratorKind) ObjectInterface {
	closure := []Instruction{
		EmitEvaluateNativeCallback(func(runtime *Runtime, vm *ExecutionVM) *Completion {
			vm.ScratchSpace["index"] = int(0)
			return nil
		}),
		// Loop starts here.
		EmitEvaluateNativeCallback(func(runtime *Runtime, vm *ExecutionVM) *Completion {
			completion := LengthOfArrayLike(runtime, array)
			if completion.Type != Normal {
				return completion
			}

			len := int(completion.Value.(*JavaScriptValue).Value.(*Number).Value)

			// Break the loop if no more elements are left.
			index := vm.ScratchSpace["index"].(int)
			if index >= len {
				return NewNormalCompletion(NewBooleanValue(true))
			}

			indexNumber := NewNumberValue(float64(index), false)

			var result *JavaScriptValue
			if kind == ArrayIteratorKindKey {
				result = indexNumber
			} else {
				completion = ToString(indexNumber)
				if completion.Type != Normal {
					return completion
				}

				elementKey := completion.Value.(*JavaScriptValue)

				completion = array.Get(runtime, elementKey, NewJavaScriptValue(TypeObject, array))
				if completion.Type != Normal {
					return completion
				}

				elementValue := completion.Value.(*JavaScriptValue)

				if kind == ArrayIteratorKindValue {
					result = elementValue
				} else {
					resultArray := CreateArrayFromList(runtime, []*JavaScriptValue{indexNumber, elementValue})
					result = NewJavaScriptValue(TypeObject, resultArray)
				}
			}

			vm.ScratchSpace["result"] = CreateIteratorResultObject(runtime, result, false)
			return NewNormalCompletion(NewBooleanValue(false))
		}),
		// If the previous completion was true, break the loop and complete the closure.
		EmitJumpIfTrue(3),
		// Yield the result.
		EmitYield(func(runtime *Runtime, vm *ExecutionVM) *JavaScriptValue {
			result := vm.ScratchSpace["result"].(*JavaScriptValue)
			vm.ScratchSpace["result"] = nil
			return result
		}),
		// Increment the index.
		EmitEvaluateNativeCallback(func(runtime *Runtime, vm *ExecutionVM) *Completion {
			vm.ScratchSpace["index"] = vm.ScratchSpace["index"].(int) + 1
			return nil
		}),
	}

	// Loop back to just after the initial setup.
	closure = append(closure, EmitJump(-len(closure)))

	// Clean up the scratch space and return undefined.
	cleanup := EmitEvaluateNativeCallback(func(runtime *Runtime, vm *ExecutionVM) *Completion {
		delete(vm.ScratchSpace, "index")
		delete(vm.ScratchSpace, "result")
		return NewNormalCompletion(NewUndefinedValue())
	})
	closure = append(closure, cleanup)

	return CreateIteratorFromClosure(
		runtime,
		closure,
		"%ArrayIteratorPrototype%",
		runtime.GetRunningRealm().GetIntrinsic(IntrinsicArrayIteratorPrototype),
	)
}
