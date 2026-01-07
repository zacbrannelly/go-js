package runtime

type Iterator struct {
	Iterator *JavaScriptValue
	Next     *JavaScriptValue
	Done     bool
}

type IteratorClosure func() *Completion

type IteratorKind int

const (
	IteratorKindSync IteratorKind = iota
	IteratorKindAsync
)

var (
	nextString   = NewStringValue("next")
	valueString  = NewStringValue("value")
	doneString   = NewStringValue("done")
	returnString = NewStringValue("return")
)

type IteratorStepResult struct {
	Value *JavaScriptValue
	Done  bool
}

func CreateIteratorFromClosure(
	runtime *Runtime,
	closure []Instruction,
	generatorBrand string,
	generatorPrototype ObjectInterface,
) ObjectInterface {
	generator := OrdinaryObjectCreate(generatorPrototype)
	generatorObj := generator.(*Object)

	generatorObj.IsGenerator = true
	generatorObj.GeneratorBrand = generatorBrand
	generatorObj.GeneratorState = GeneratorStateSuspendedStart

	callerContext := runtime.GetRunningExecutionContext()

	calleeContext := &ExecutionContext{
		Function:  nil,
		Realm:     runtime.GetRunningRealm(),
		Script:    callerContext.Script,
		Generator: generatorObj,
		VM:        NewExecutionVM(),
	}
	runtime.PushExecutionContext(calleeContext)
	defer runtime.PopExecutionContext()

	GeneratorStartWithClosure(runtime, generatorObj, closure)

	return generator
}

func IteratorStepValue(runtime *Runtime, iterator *Iterator) *Completion {
	completion := IteratorStep(runtime, iterator)
	if completion.Type != Normal {
		return completion
	}

	result, ok := completion.Value.(*IteratorStepResult)
	if ok && result.Done {
		return completion
	}

	value, ok := completion.Value.(*JavaScriptValue)
	if !ok {
		panic("Assert failed: IteratorStepValue received an invalid result.")
	}

	completion = IteratorValue(runtime, value)
	if completion.Type == Throw {
		iterator.Done = true
	}

	return completion
}

func IteratorStep(runtime *Runtime, iterator *Iterator) *Completion {
	completion := IteratorNext(runtime, iterator, nil)
	if completion.Type != Normal {
		return completion
	}

	result := completion.Value.(*JavaScriptValue)

	completion = IteratorComplete(runtime, result)
	if completion.Type == Throw {
		iterator.Done = true
		return completion
	}

	done := completion.Value.(*JavaScriptValue).Value.(*Boolean).Value

	if done {
		iterator.Done = true
		return NewNormalCompletion(&IteratorStepResult{
			Done: true,
		})
	}

	return NewNormalCompletion(result)
}

func IteratorValue(runtime *Runtime, iteratorResult *JavaScriptValue) *Completion {
	object := iteratorResult.Value.(ObjectInterface)
	return object.Get(runtime, valueString, iteratorResult)
}

func IteratorNext(runtime *Runtime, iterator *Iterator, value *JavaScriptValue) *Completion {
	if _, callable := iterator.Next.Value.(FunctionInterface); !callable {
		return NewThrowCompletion(NewTypeError(runtime, "Iterator.next is not a function"))
	}

	args := []*JavaScriptValue{}
	if value != nil {
		args = append(args, value)
	}

	function := iterator.Next.Value.(FunctionInterface)
	completion := function.Call(runtime, iterator.Iterator, args)
	if completion.Type != Normal {
		return completion
	}

	if completion.Type == Throw {
		iterator.Done = true
		return completion
	}

	result := completion.Value.(*JavaScriptValue)
	if result.Type != TypeObject {
		iterator.Done = true
		return NewThrowCompletion(NewTypeError(runtime, "Iterator.next returned a non-object"))
	}

	return NewNormalCompletion(result)
}

func IteratorComplete(runtime *Runtime, iteratorResult *JavaScriptValue) *Completion {
	object := iteratorResult.Value.(ObjectInterface)
	completion := object.Get(runtime, doneString, iteratorResult)
	if completion.Type != Normal {
		return completion
	}

	done := completion.Value.(*JavaScriptValue)
	return ToBoolean(done)
}

func IteratorClose(runtime *Runtime, iterator *Iterator, providedCompletion *Completion) *Completion {
	completion := ToObject(runtime, iterator.Iterator)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue)
	completion = GetMethod(runtime, object, returnString)

	if completion.Type == Normal {
		result := completion.Value.(*JavaScriptValue)
		if result.Type == TypeUndefined {
			return providedCompletion
		}

		if _, callable := result.Value.(FunctionInterface); !callable {
			completion = NewThrowCompletion(NewTypeError(runtime, "Iterator.return is not a function"))
		} else {
			function := result.Value.(FunctionInterface)
			completion = function.Call(runtime, object, []*JavaScriptValue{})
		}
	}

	if providedCompletion.Type == Throw {
		return providedCompletion
	}

	if completion.Type == Throw {
		return completion
	}

	value := completion.Value.(*JavaScriptValue)
	if value.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Iterator.return returned a non-object"))
	}

	return providedCompletion
}

func IteratorToList(runtime *Runtime, iterator *Iterator) *Completion {
	values := make([]*JavaScriptValue, 0)
	for {
		completion := IteratorStepValue(runtime, iterator)
		if completion.Type != Normal {
			return completion
		}

		if iterator.Done {
			break
		}

		values = append(values, completion.Value.(*JavaScriptValue))
	}

	return NewNormalCompletion(values)
}

func GetIteratorDirect(runtime *Runtime, obj *JavaScriptValue) *Completion {
	object := obj.Value.(ObjectInterface)

	completion := object.Get(runtime, nextString, obj)
	if completion.Type != Normal {
		return completion
	}
	next := completion.Value.(*JavaScriptValue)

	return NewNormalCompletion(&Iterator{
		Iterator: obj,
		Next:     next,
		Done:     false,
	})
}

func GetIterator(runtime *Runtime, obj *JavaScriptValue, kind IteratorKind) *Completion {
	var method *JavaScriptValue
	if kind == IteratorKindAsync {
		panic("TODO: Implement GetIterator for async iterators.")
	} else {
		completion := GetMethod(runtime, obj, runtime.SymbolIterator)
		if completion.Type != Normal {
			return completion
		}

		method = completion.Value.(*JavaScriptValue)
	}

	return GetIteratorFromMethod(runtime, method, obj)
}

func GetIteratorFromMethod(runtime *Runtime, method *JavaScriptValue, object *JavaScriptValue) *Completion {
	// Semantics of Call operation in the spec.
	if _, callable := method.Value.(FunctionInterface); !callable {
		return NewThrowCompletion(NewTypeError(runtime, "Method provided is not a function"))
	}

	function := method.Value.(FunctionInterface)
	completion := function.Call(runtime, object, []*JavaScriptValue{})
	if completion.Type != Normal {
		return completion
	}

	iterator := completion.Value.(*JavaScriptValue)
	if iterator.Type != TypeObject {
		return NewThrowCompletion(NewTypeError(runtime, "Iterator is not an object"))
	}

	return GetIteratorDirect(runtime, iterator)
}

func IfAbruptCloseIterator(runtime *Runtime, value *Completion, iterator *Iterator) *Completion {
	if value.Type != Normal {
		return IteratorClose(runtime, iterator, value)
	}

	return value
}

func GetMethod(runtime *Runtime, obj *JavaScriptValue, key *JavaScriptValue) *Completion {
	completion := ToObject(runtime, obj)
	if completion.Type != Normal {
		return completion
	}

	object := completion.Value.(*JavaScriptValue).Value.(ObjectInterface)
	completion = object.Get(runtime, key, obj)
	if completion.Type != Normal {
		return completion
	}

	method := completion.Value.(*JavaScriptValue)
	if method.Type == TypeUndefined || method.Type == TypeNull {
		return NewNormalCompletion(NewUndefinedValue())
	}

	if _, callable := method.Value.(FunctionInterface); !callable {
		return NewThrowCompletion(NewTypeError(runtime, "Method is not a function"))
	}

	return NewNormalCompletion(method)
}
