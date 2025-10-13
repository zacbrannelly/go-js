package runtime

func NegateBooleanValue(result *Completion) *Completion {
	if result.Type == Throw {
		return result
	}

	resultValue := result.Value.(*JavaScriptValue)
	if resultValue.Type == TypeBoolean {
		return NewNormalCompletion(NewBooleanValue(!resultValue.Value.(*Boolean).Value))
	}

	panic("Assert failed: Result is not a boolean.")
}
