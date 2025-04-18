// Package errors provides custom error types and error handling utilities for gh-topics
package errors

import (
	"fmt"
)

// ErrorType represents the type of an error
type ErrorType string

const (
	// APIError represents errors related to API calls
	APIError ErrorType = "API_ERROR"
	// ConfigError represents errors related to configuration
	ConfigError ErrorType = "CONFIG_ERROR"
	// InputError represents errors related to user input
	InputError ErrorType = "INPUT_ERROR"
	// InternalError represents internal application errors
	InternalError ErrorType = "INTERNAL_ERROR"
)

// AppError is a custom error type that includes context about the error
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new API error
func NewAPIError(message string, err error) *AppError {
	return &AppError{
		Type:    APIError,
		Message: message,
		Err:     err,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(message string, err error) *AppError {
	return &AppError{
		Type:    ConfigError,
		Message: message,
		Err:     err,
	}
}

// NewInputError creates a new input error
func NewInputError(message string, err error) *AppError {
	return &AppError{
		Type:    InputError,
		Message: message,
		Err:     err,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    InternalError,
		Message: message,
		Err:     err,
	}
}

// UserFriendlyMessage returns a user-friendly error message
func UserFriendlyMessage(err error) string {
	if appErr, ok := err.(*AppError); ok {
		switch appErr.Type {
		case APIError:
			return fmt.Sprintf("GitHub API error: %s. Please check your network connection and try again.", appErr.Message)
		case ConfigError:
			return fmt.Sprintf("Configuration error: %s. Please check your configuration settings.", appErr.Message)
		case InputError:
			return fmt.Sprintf("Input error: %s. Please check your command arguments.", appErr.Message)
		case InternalError:
			return fmt.Sprintf("Internal error: %s. This is likely a bug in the application.", appErr.Message)
		}
	}
	return fmt.Sprintf("Error: %v", err)
}
