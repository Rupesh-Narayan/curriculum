package noonerror

import "errors"

type NoonError struct {
	Err     error
	Message string
}

func (n *NoonError) Error() string {
	return n.Err.Error()
}

func New(e error, message string) error {
	return &NoonError{e, message}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

var (
	// ErrParamMissing required field missing error
	ErrParamMissing = errors.New("paramMissing")

	// ErrParamMissing required field missing error
	ErrBadRequest = errors.New("badRequest")

	// ErrInternalServer sent in case of server error
	ErrInternalServer = errors.New("somethingWentWrong")

	// ErrUserNotFound user vertex not available in db
	ErrUserNotFound = errors.New("userNotFound")

	// ErrRelationExists requested relation exists between users
	ErrRelationExists = errors.New("relationExists")

	// ErrInvalidRequest for invalid request
	ErrInvalidRequest = errors.New("invalidRequest")
)
