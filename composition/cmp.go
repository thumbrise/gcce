package composition

type compiler interface {
	deps(s schema) ([]*node, error)
}
