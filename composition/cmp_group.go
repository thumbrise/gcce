package composition

import "reflect"

type groupCompiler struct {
	rt      reflect.Type
	matched []*node
}

func newGroupCompiler(rt reflect.Type, matched []*node) *groupCompiler {
	return &groupCompiler{
		rt:      rt,
		matched: matched,
	}
}

func (c *groupCompiler) deps(s schema) (deps []*node, err error) {
	return c.matched, nil
}
