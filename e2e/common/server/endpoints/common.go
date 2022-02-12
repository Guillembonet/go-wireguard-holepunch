package endpoints

import "net/http"

// Error type.
type Error struct {
	Code    int     `json:"code"`
	Message string  `json:"error"`
	Reason  *string `json:"reason"`
}

// NewErrorValidation creates API validation error with custom message.
func NewErrorValidation(message string, reasonErr error) *Error {
	reason := reasonErr.Error()
	return &Error{
		Code:    http.StatusBadRequest,
		Message: message,
		Reason:  &reason,
	}
}

// NewGenericError creates generic error with custom message.
func NewGenericError(message string, reasonErr error) *Error {
	reason := reasonErr.Error()
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: message,
		Reason:  &reason,
	}
}
