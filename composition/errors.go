package composition

import "errors"

var (
	ErrTypeNotExists              = errors.New("not exists in the container")
	ErrInvalidInvocationSignature = errors.New("invalid invocation signature")
	ErrCycleDetected              = errors.New("cycle detected")
	ErrMultipleDefinitions        = errors.New("multiple definitions")
	ErrNilConstructor             = errors.New("invalid constructor, got nil")
	ErrNotImplement               = errors.New("does not implement")
	ErrNilInterface               = errors.New("nil: not a pointer to interface")
	ErrNotInterfacePtr            = errors.New("not a pointer to interface")
	ErrInvalidCtorSignature       = errors.New("invalid constructor signature")
	ErrAnonymousCtor              = errors.New("anonymous constructor: named package-level functions only")
)
