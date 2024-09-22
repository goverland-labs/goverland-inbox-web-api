package response

import (
	"errors"
	"math"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type parametrizedError interface {
	SetError(key string, code ErrCode, message string)
}

func ResolveError(err error, mappings ...map[error]func(err error) Error) Error {
	for _, m := range mappings {
		for k, v := range m {
			if errors.Is(err, k) {
				return v(err)
			}
		}
	}

	details, ok := status.FromError(err)
	if !ok {
		return NewInternalError()
	}

	switch details.Code() {
	case codes.NotFound:
		return NewNotFoundError()

	case codes.InvalidArgument, codes.OutOfRange:
		ed := NewValidationError()
		enrichErrDetails(ed, details)

		return ed

	case codes.FailedPrecondition, codes.Canceled, codes.AlreadyExists:
		ed := NewNotAcceptableError()
		enrichErrDetails(ed, details)

		return ed

	case codes.PermissionDenied:
		return NewPermissionDeniedError()

	case codes.Unauthenticated:
		return NewUnauthorizedError()

	case codes.Unavailable:
		return NewNotAcceptableError()

	case codes.ResourceExhausted:
		retryAfter := 0
		msg := ""
		for _, d := range details.Details() {
			if info, ok := d.(*errdetails.RetryInfo); ok {
				delay := info.GetRetryDelay().AsDuration()
				retryAfter = int(math.Ceil(delay.Seconds()))
			}

			if info, ok := d.(*errdetails.QuotaFailure); ok {
				for _, v := range info.GetViolations() {
					msg = v.GetDescription()
					break
				}
			}
		}

		return NewRateLimitedError(retryAfter, msg)
	}

	return NewInternalError()
}

func enrichErrDetails(err parametrizedError, st *status.Status) {
	e := st.Details()
	if len(e) == 0 {
		err.SetError(GeneralErrorKey, WrongValue, st.Message())

		return
	}
}

// IsInternalError returns false when the error is caused by invalid request data or by other mismatches caused by an user.
// It can be used to determine whether the error should be logged. If the function returns false, the error shouldn't
// be logged.
func IsInternalError(err error) bool {
	if err == nil {
		return false
	}

	e := ResolveError(err)
	switch e.(type) {
	case *ValidationError:
		return false

	case *RateLimitedError:
		return false
	}

	return true
}
