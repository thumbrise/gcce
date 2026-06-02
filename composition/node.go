package composition

import (
	"fmt"
	"reflect"
)

func newConstructorNode(ctor interface{}) (*node, error) {
	f, valid := inspectFunction(ctor)
	if !valid {
		return nil, fmt.Errorf("invalid constructor signature, got %s", reflect.TypeOf(ctor))
	}
	cmp, ok := newConstructorCompiler(f)
	if !ok {
		if isAnonymous(f) {
			return nil, fmt.Errorf("anonymous constructor: named package-level functions only, got %s", f.Type)
		}
		return nil, fmt.Errorf("invalid constructor signature, got %s", f.Type)
	}
	return &node{
		rt:       f.Out(0),
		compiler: cmp,
	}, nil
}

type node struct {
	compiler
	rt       reflect.Type
	metadata Metadata
}

func (n *node) String() string {
	return n.rt.String()
}
