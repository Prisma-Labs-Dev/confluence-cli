package cli

import "fmt"

type ValidationError struct {
	Message string
	Hint    string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func validationError(message, hint string) error {
	return &ValidationError{Message: message, Hint: hint}
}

func validationErrorf(hint, format string, args ...any) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...), Hint: hint}
}

func validateRange(name string, value, min, max int, hint string) error {
	if value < min || value > max {
		return validationErrorf(hint, "%s must be between %d and %d", name, min, max)
	}
	return nil
}
