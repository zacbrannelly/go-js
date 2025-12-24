package runtime

var (
	zeroString = NewStringValue("0")
	oneString  = NewStringValue("1")
)

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

		next, ok := completion.Value.(*IteratorStepResult)
		if ok && next.Done {
			return NewNormalCompletion(targetValue)
		}

		value, ok := completion.Value.(*JavaScriptValue)
		if !ok {
			panic("Assert failed: AddEntriesFromIterable received an invalid result.")
		}

		if value.Type != TypeObject {
			throwCompletion := NewThrowCompletion(NewTypeError(runtime, "Iterator.next returned a non-object"))
			return IteratorClose(runtime, iterator, throwCompletion)
		}

		nextObj := value.Value.(ObjectInterface)

		completion = nextObj.Get(runtime, zeroString, value)
		IfAbruptCloseIterator(runtime, completion, iterator)

		key := completion.Value.(*JavaScriptValue)

		completion = nextObj.Get(runtime, oneString, value)
		IfAbruptCloseIterator(runtime, completion, iterator)

		value = completion.Value.(*JavaScriptValue)

		completion = adder.Call(runtime, targetValue, []*JavaScriptValue{key, value})
		IfAbruptCloseIterator(runtime, completion, iterator)
	}
}
