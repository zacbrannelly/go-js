package runtime

import "slices"

func LoopContinues(runtime *Runtime, completion *Completion) bool {
	if completion.Type == Normal {
		return true
	}

	if completion.Type != Continue {
		return false
	}

	if completion.Target == "" {
		return true
	}

	if slices.Contains(runtime.GetRunningLabels(), completion.Target) {
		return true
	}

	return false
}
