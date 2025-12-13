package runtime

type ArrayIteratorKind int

const (
	ArrayIteratorKindKey ArrayIteratorKind = iota
	ArrayIteratorKindValue
	ArrayIteratorKindEntry
)

func CreateArrayIterator(runtime *Runtime, array ObjectInterface, kind ArrayIteratorKind) ObjectInterface {
	closure := func() *Completion {
		panic("TODO: Implement CreateArrayIterator closure.")
	}

	return CreateIteratorFromClosure(
		runtime,
		closure,
		"%ArrayIteratorPrototype%",
		runtime.GetRunningRealm().Intrinsics[IntrinsicArrayIteratorPrototype],
	)
}
