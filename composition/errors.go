package composition

import "errors"

var (
	ErrTypeNotExists              = errors.New("not exists in the container")
	errInvalidInvocationSignature = errors.New("invalid invocation signature")
	errCycleDetected              = errors.New("cycle detected")
	errMultipleDefinitions        = errors.New("multiple definitions")
)
