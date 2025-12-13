package runtime

var zeroString = NewStringValue("0")
var oneString = NewStringValue("1")

func AddEntriesFromIterable(
	runtime *Runtime,
	target ObjectInterface,
	iterable *JavaScriptValue,
	adder *FunctionObject,
) *Completion {
	targetValue := NewJavaScriptValue(TypeObject, target)
	completion := GetIterator(runtime, iterable, IteratorKindSync)
	if completion.Type != Normal {
		return completion
	}

	iterator := completion.Value.(*Iterator)

	for {
		completion := IteratorStepValue(runtime, iterator)
		if completion.Type != Normal {
			return completion
		}

		next := completion.Value.(*IteratorStepResult)
		if next.Done {
			return NewNormalCompletion(targetValue)
		}

		nextObj := next.Value.Value.(ObjectInterface)

		if next.Value.Type != TypeObject {
			throwCompletion := NewThrowCompletion(NewTypeError("Iterator.next returned a non-object"))
			return IteratorClose(runtime, iterator, throwCompletion)
		}

		completion = nextObj.Get(runtime, zeroString, next.Value)
		IfAbruptCloseIterator(runtime, completion, iterator)

		key := completion.Value.(*JavaScriptValue)

		completion = nextObj.Get(runtime, oneString, next.Value)
		IfAbruptCloseIterator(runtime, completion, iterator)

		value := next.Value.Value.(*JavaScriptValue)

		completion = adder.Call(runtime, targetValue, []*JavaScriptValue{key, value})
		IfAbruptCloseIterator(runtime, completion, iterator)
	}
}
