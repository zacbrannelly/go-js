package runtime

type SyntaxError struct {
	Message string
}

func NewSyntaxError(message string) *SyntaxError {
	return &SyntaxError{
		Message: message,
	}
}

func (e *SyntaxError) Error() string {
	return e.Message
}

type TypeError struct {
	Message string
}

func NewTypeError(message string) *TypeError {
	return &TypeError{
		Message: message,
	}
}
