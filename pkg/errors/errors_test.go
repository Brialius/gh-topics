package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name   string
		appErr *AppError
		want   string
	}{
		{
			name: "error with wrapped error",
			appErr: &AppError{
				Type:    APIError,
				Message: "test message",
				Err:     errors.New("wrapped error"),
			},
			want: "API_ERROR: test message: wrapped error",
		},
		{
			name: "error without wrapped error",
			appErr: &AppError{
				Type:    InputError,
				Message: "test message",
				Err:     nil,
			},
			want: "INPUT_ERROR: test message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	appErr := &AppError{
		Type:    APIError,
		Message: "test message",
		Err:     wrappedErr,
	}

	if got := appErr.Unwrap(); got != wrappedErr {
		t.Errorf("AppError.Unwrap() = %v, want %v", got, wrappedErr)
	}
}

func TestUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "API error",
			err: &AppError{
				Type:    APIError,
				Message: "test message",
				Err:     errors.New("wrapped error"),
			},
			want: "GitHub API error: test message: wrapped error",
		},
		{
			name: "Config error",
			err: &AppError{
				Type:    ConfigError,
				Message: "test message",
				Err:     nil,
			},
			want: "Configuration error: test message. Please check your configuration settings.",
		},
		{
			name: "Input error",
			err: &AppError{
				Type:    InputError,
				Message: "test message",
				Err:     nil,
			},
			want: "Input error: test message. Please check your command arguments.",
		},
		{
			name: "Internal error",
			err: &AppError{
				Type:    InternalError,
				Message: "test message",
				Err:     nil,
			},
			want: "Internal error: test message. This is likely a bug in the application.",
		},
		{
			name: "Standard error",
			err:  errors.New("standard error"),
			want: "Error: standard error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserFriendlyMessage(tt.err); got != tt.want {
				t.Errorf("UserFriendlyMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	message := "test message"
	wrappedErr := errors.New("wrapped error")

	appErr := NewAPIError(message, wrappedErr)

	if appErr.Type != APIError {
		t.Errorf("NewAPIError().Type = %v, want %v", appErr.Type, APIError)
	}
	if appErr.Message != message {
		t.Errorf("NewAPIError().Message = %v, want %v", appErr.Message, message)
	}
	if appErr.Err != wrappedErr {
		t.Errorf("NewAPIError().Err = %v, want %v", appErr.Err, wrappedErr)
	}
}

func TestUserFriendlyMessagePreservesCauses(t *testing.T) {
	for _, kind := range []ErrorType{APIError, ConfigError, InputError, InternalError} {
		for _, cause := range []string{"HTTP 401: Bad credentials", "HTTP 404: Not Found", "HTTP 429: rate limit", "config.yml: yaml: line 3: invalid syntax"} {
			err := &AppError{Type: kind, Message: "operation failed", Err: errors.New(cause)}
			got := UserFriendlyMessage(err)
			if !strings.Contains(got, cause) || !strings.Contains(got, "operation failed") {
				t.Errorf("cause lost for %s: %s", kind, got)
			}
		}
	}
}
