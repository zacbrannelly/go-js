package runtime

import "fmt"

type SyntaxError struct {
	Message string
}

func NewSyntaxError(message string) *SyntaxError {
	return &SyntaxError{
		Message: message,
	}
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("SyntaxError: %s", e.Message)
}

type TypeError struct {
	Message string
}

func NewTypeError(message string) *TypeError {
	return &TypeError{
		Message: message,
	}
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("TypeError: %s", e.Message)
}

type ReferenceError struct {
	Message string
}

func NewReferenceError(message string) *ReferenceError {
	return &ReferenceError{
		Message: message,
	}
}

func (e *ReferenceError) Error() string {
	return fmt.Sprintf("ReferenceError: %s", e.Message)
}
