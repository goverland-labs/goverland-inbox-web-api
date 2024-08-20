package response

import (
	"fmt"
	"net/http"
)

const (
	GeneralErrorKey = "_"
)

type InternalError struct {
	BaseError
}

func (e *InternalError) PublicMessage() string {
	return "internal error"
}

func (e *InternalError) GetHTTPStatus() int {
	return http.StatusInternalServerError
}

func NewInternalError() *InternalError {
	err := &InternalError{}
	return err
}

type NotFoundError struct {
	BaseError
}

func (e *NotFoundError) PublicMessage() string {
	return "not found"
}

func (e *NotFoundError) GetHTTPStatus() int {
	return http.StatusNotFound
}

func NewNotFoundError() *NotFoundError {
	err := &NotFoundError{}

	return err
}

type PermissionDeniedError struct {
	BaseError
	details map[string]ErrorMessage
}

func (e *PermissionDeniedError) PublicMessage() string {
	return "permission denied"
}

func (e *PermissionDeniedError) GetHTTPStatus() int {
	return http.StatusForbidden
}

func (e *PermissionDeniedError) Errors() map[string]ErrorMessage {
	return e.details
}

func (e *PermissionDeniedError) SetError(key string, code ErrCode, message string) {
	e.details[key] = ErrorMessage{
		Code:    code,
		Message: message,
	}
}

func NewPermissionDeniedError(details ...map[string]ErrorMessage) *PermissionDeniedError {
	err := &PermissionDeniedError{
		details: make(map[string]ErrorMessage),
	}

	if len(details) > 0 {
		for key, description := range details[0] {
			err.SetError(key, description.Code, description.Message)
		}
	}

	return err
}

type UnauthorizedError struct {
	BaseError
}

func (e *UnauthorizedError) PublicMessage() string {
	return "unauthorized"
}

func (e *UnauthorizedError) GetHTTPStatus() int {
	return http.StatusUnauthorized
}

func NewUnauthorizedError() *UnauthorizedError {
	return &UnauthorizedError{}
}

type ErrorMessage struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
}

func (e ErrorMessage) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func WrongValueError(message string) ErrorMessage {
	return ErrorMessage{
		Code:    WrongValue,
		Message: message,
	}
}

func WrongFormatError(message string) ErrorMessage {
	return ErrorMessage{
		Code:    WrongFormat,
		Message: message,
	}
}

func MissedValueError(message string) ErrorMessage {
	return ErrorMessage{
		Code:    MissedValue,
		Message: message,
	}
}

type ValidationError struct {
	BaseError
	details map[string]ErrorMessage
}

func (e *ValidationError) PublicMessage() string {
	return "validation error"
}

func (e *ValidationError) GetHTTPStatus() int {
	return http.StatusBadRequest
}

func (e *ValidationError) Errors() map[string]ErrorMessage {
	return e.details
}

func (e *ValidationError) SetError(key string, code ErrCode, message string) {
	e.details[key] = ErrorMessage{
		Code:    code,
		Message: message,
	}
}

func NewValidationError(details ...map[string]ErrorMessage) *ValidationError {
	err := &ValidationError{
		details: make(map[string]ErrorMessage),
	}

	if len(details) > 0 {
		for key, description := range details[0] {
			err.SetError(key, description.Code, description.Message)
		}
	}

	return err
}

type UnprocessableEntityError struct {
	ValidationError
}

func (e *UnprocessableEntityError) PublicMessage() string {
	return "not processable"
}

func (e *UnprocessableEntityError) GetHTTPStatus() int {
	return http.StatusUnprocessableEntity
}

func NewNotAcceptableError(details ...map[string]ErrorMessage) *UnprocessableEntityError {
	list := make(map[string]ErrorMessage)
	if len(details) > 0 {
		for name, description := range details[0] {
			list[name] = description
		}
	}

	err := &UnprocessableEntityError{ValidationError{details: list}}

	return err
}

type RateLimitedError struct {
	BaseError

	// time in seconds
	RetryAfter int

	// public message to display to the client
	Message string
}

func NewRateLimitedError(retryAfter int, msg string) *RateLimitedError {
	return &RateLimitedError{
		RetryAfter: retryAfter,
		Message:    msg,
	}
}

func (e *RateLimitedError) PublicMessage() string {
	if e.Message != "" {
		return e.Message
	}

	return "too many requests"
}

func (e *RateLimitedError) GetHTTPStatus() int {
	return http.StatusTooManyRequests
}
