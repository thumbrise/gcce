package composition

type ctorType int

const (
	ctorUnknown           ctorType = iota
	ctorValue                      // (deps) (result)
	ctorValueError                 // (deps) (result, error)
	ctorValueCleanup               // (deps) (result, cleanup)
	ctorValueCleanupError          // (deps) (result, cleanup, error)
)

type constructorCompiler struct {
	typ ctorType
	fn  function
}

func newConstructorCompiler(fn function) (*constructorCompiler, bool) {
	if isAnonymous(fn) {
		return nil, false
	}

	ctorType := determineCtorType(fn)
	if ctorType == ctorUnknown {
		return nil, false
	}

	return &constructorCompiler{
		typ: ctorType,
		fn:  fn,
	}, true
}

func (c constructorCompiler) deps(s schema) (deps []*node, err error) {
	for i := 0; i < c.fn.NumIn(); i++ {
		in := c.fn.In(i)

		node, err := s.find(in)
		if err != nil {
			return nil, err
		}

		deps = append(deps, node)
	}

	return deps, nil
}

func determineCtorType(fn function) ctorType {
	switch {
	case fn.NumOut() == 1:
		return ctorValue
	case fn.NumOut() == 2:
		if isError(fn.Out(1)) {
			return ctorValueError
		}

		if isCleanup(fn.Out(1)) {
			return ctorValueCleanup
		}
	case fn.NumOut() == 3 && isCleanup(fn.Out(1)) && isError(fn.Out(2)):
		return ctorValueCleanupError
	}

	return ctorUnknown
}
