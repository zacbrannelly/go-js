package runtime

import "slices"

type ForInIterator struct {
	Object           ObjectInterface
	ObjectWasVisited bool
	VisitedKeys      []string
	RemainingKeys    []*JavaScriptValue
}

func NewForInIterator(object ObjectInterface) *ForInIterator {
	return &ForInIterator{
		Object:           object,
		ObjectWasVisited: false,
		VisitedKeys:      make([]string, 0),
		RemainingKeys:    make([]*JavaScriptValue, 0),
	}
}

func (iterator *ForInIterator) Next(runtime *Runtime) *Completion {
	for {
		if !iterator.ObjectWasVisited {
			completion := iterator.Object.OwnPropertyKeys()
			if completion.Type != Normal {
				return completion
			}

			for _, key := range completion.Value.([]*JavaScriptValue) {
				if key.Type == TypeString {
					iterator.RemainingKeys = append(iterator.RemainingKeys, key)
				}
			}

			iterator.ObjectWasVisited = true
		}

		for len(iterator.RemainingKeys) > 0 {
			key := iterator.RemainingKeys[0]
			keyString := key.Value.(*String).Value
			iterator.RemainingKeys = iterator.RemainingKeys[1:]

			if slices.Contains(iterator.VisitedKeys, keyString) {
				continue
			}

			completion := iterator.Object.GetOwnProperty(runtime, key)
			if completion.Type != Normal {
				return completion
			}

			if desc, ok := completion.Value.(PropertyDescriptor); ok && desc != nil {
				iterator.VisitedKeys = append(iterator.VisitedKeys, keyString)
				if desc.GetEnumerable() {
					return NewNormalCompletion(CreateIteratorResultObject(runtime, key, false))
				}
			}
		}

		completion := iterator.Object.GetPrototypeOf()
		if completion.Type != Normal {
			return completion
		}

		prototype := completion.Value.(*JavaScriptValue)
		if prototype.Type == TypeNull {
			return NewNormalCompletion(CreateIteratorResultObject(runtime, NewUndefinedValue(), true))
		}

		iterator.Object = prototype.Value.(ObjectInterface)
		iterator.ObjectWasVisited = false
	}
}
