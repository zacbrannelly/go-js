package runtime

type CompletionType int

const (
	Normal CompletionType = iota
	Break
	Continue
	Return
	Throw
	Unused
)

type Completion struct {
	Type   CompletionType
	Value  any
	Target string
}

func NewNormalCompletion(value any) *Completion {
	return &Completion{
		Type:  Normal,
		Value: value,
	}
}

func NewThrowCompletion(value any) *Completion {
	return &Completion{
		Type:  Throw,
		Value: value,
	}
}

func NewUnusedCompletion() *Completion {
	return &Completion{
		Type: Unused,
	}
}
