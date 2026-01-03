package runtime

// TODO: Support "labelSet" parameter.
func LabelledEvaluation(runtime *Runtime, loopCompletion *Completion) *Completion {
	if loopCompletion.Type == Break {
		if loopCompletion.Target == "" {
			if loopCompletion.Value == nil {
				loopCompletion = NewNormalCompletion(NewUndefinedValue())
			} else {
				loopCompletion = NewNormalCompletion(loopCompletion.Value)
			}
		}
	}

	return loopCompletion
}
