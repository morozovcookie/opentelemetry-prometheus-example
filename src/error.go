package otelexample

import (
	"errors"
	"fmt"
)

var _ fmt.Stringer = (*ErrorCode)(nil)

// ErrorCode represents a type of the error.
type ErrorCode string

const (
	ErrorCodeOK       = ErrorCode("ok")
	ErrorCodeInvalid  = ErrorCode("invalid")
	ErrorCodeInternal = ErrorCode("internal")
)

// The String method is used to print values passed as an operand
// to any format that accepts a string or to an unformatted printer
// such as Print.
func (ec ErrorCode) String() string {
	return string(ec)
}

var _ error = (*Error)(nil)

// Error represents an error.
type Error struct {
	// Code is the machine readable code, for reference purpose.
	Code ErrorCode

	// Message is the human readable message for end user.
	Message string

	// Err is the embed error.
	Err error
}

func (e *Error) Error() string {
	return ""
}

func ErrorCodeFromError(err error) ErrorCode {
	if err == nil {
		return ErrorCodeOK
	}

	var customErr *Error
	ok := errors.As(err, &customErr)
	if ok && customErr.Code != "" {
		return customErr.Code
	}

	if ok && customErr.Err != nil {
		return ErrorCodeFromError(customErr.Err)
	}

	return ErrorCodeInternal
}

func ErrorMessageFromError(err error) string {
	if err == nil {
		return ""
	}

	var customErr *Error
	ok := errors.As(err, &customErr)
	if ok && customErr.Message != "" {
		return customErr.Message
	}

	if ok && customErr.Err != nil {
		return ErrorMessageFromError(customErr.Err)
	}

	return "an internal error has occurred"
}
