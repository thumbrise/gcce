package errs

import "errors"

var (
	ErrNotImplement    = errors.New("does not implement")
	ErrNilInterface    = errors.New("nil: not a pointer to interface")
	ErrNotInterfacePtr = errors.New("not a pointer to interface")
)
