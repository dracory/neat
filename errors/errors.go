package errors

import (
	"errors"
	"fmt"

	contractserrors "github.com/dracory/neat/contracts/errors"
)

type errorString struct {
	text   string
	module string
	args   []any
}

// New creates a new error with the provided text and optional module
func New(text string, module ...string) contractserrors.Error {
	err := &errorString{
		text: text,
	}

	if len(module) > 0 {
		err.module = module[0]
	}

	return err
}

// Args sets the arguments for formatting the error message.
func (e *errorString) Args(args ...any) contractserrors.Error {
	e.args = args
	return e
}

// Error returns the formatted error message with module prefix and arguments.
func (e *errorString) Error() string {
	formattedText := e.text

	if len(e.args) > 0 {
		formattedText = fmt.Sprintf(e.text, e.args...)
	}

	if e.module != "" {
		formattedText = fmt.Sprintf("[%s] %s", e.module, formattedText)
	}

	return formattedText
}

// SetModule sets the module name for the error.
func (e *errorString) SetModule(module string) contractserrors.Error {
	e.module = module
	return e
}

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets target to that error value and returns true.
func As(err error, target any) bool {
	return errors.As(err, &target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method returning error.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
