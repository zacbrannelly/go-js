package runtime

type CompletionType int

const (
	Normal CompletionType = iota
	Break
	Continue
	Return
	Throw
)

type Completion struct {
	Type   CompletionType
	Value  any
	Target string
	Unused bool
}

func NewNormalCompletion(value any) *Completion {
	return &Completion{
		Type:   Normal,
		Value:  value,
		Unused: false,
	}
}

func NewThrowCompletion(value any) *Completion {
	return &Completion{
		Type:   Throw,
		Value:  value,
		Unused: false,
	}
}

func NewUnusedCompletion() *Completion {
	return &Completion{
		Type:   Normal,
		Unused: true,
	}
}
