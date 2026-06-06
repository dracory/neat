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
	return errors.As(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method returning error.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// StructuredError represents a structured error with type, message, and optional underlying error.
type StructuredError struct {
	Type    string
	Message string
	Err     error
	Module  string
}

// Error returns the formatted error message with type, module prefix, and underlying error.
func (e *StructuredError) Error() string {
	var formattedMessage string
	if e.Err != nil {
		formattedMessage = fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	} else {
		formattedMessage = fmt.Sprintf("%s: %s", e.Type, e.Message)
	}

	if e.Module != "" {
		formattedMessage = fmt.Sprintf("[%s] %s", e.Module, formattedMessage)
	}

	return formattedMessage
}

// Unwrap returns the underlying error for error unwrapping.
func (e *StructuredError) Unwrap() error {
	return e.Err
}

// SetModule sets the module name for the structured error.
func (e *StructuredError) SetModule(module string) *StructuredError {
	e.Module = module
	return e
}

// Standard error variables for common error conditions.
var (
	// Validation errors
	ErrNilDatabase      = &StructuredError{Type: "ValidationError", Message: "database connection cannot be nil"}
	ErrNilQuery         = &StructuredError{Type: "ValidationError", Message: "query cannot be nil"}
	ErrNotInTransaction = &StructuredError{Type: "ValidationError", Message: "operation requires an active transaction"}
	ErrInvalidSavepoint = &StructuredError{Type: "ValidationError", Message: "invalid savepoint name"}

	// Argument errors
	ErrNilModel    = &StructuredError{Type: "ArgumentError", Message: "model cannot be nil"}
	ErrNilRelation = &StructuredError{Type: "ArgumentError", Message: "relation cannot be nil"}

	// Configuration errors
	ErrInvalidDriver     = &StructuredError{Type: "ConfigurationError", Message: "invalid database driver"}
	ErrMissingConnection = &StructuredError{Type: "ConfigurationError", Message: "database connection not established"}
)

// NewValidationError creates a new validation error with the provided message.
func NewValidationError(message string) *StructuredError {
	return &StructuredError{Type: "ValidationError", Message: message}
}

// NewArgumentError creates a new argument error with the provided message.
func NewArgumentError(message string) *StructuredError {
	return &StructuredError{Type: "ArgumentError", Message: message}
}

// NewConfigurationError creates a new configuration error with the provided message.
func NewConfigurationError(message string) *StructuredError {
	return &StructuredError{Type: "ConfigurationError", Message: message}
}
