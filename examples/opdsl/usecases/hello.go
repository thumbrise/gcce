package usecases

type Hello struct {
	suffix string
}

func NewHello(suffix string) *Hello {
	return &Hello{suffix: suffix}
}

func (h *Hello) Hello(name string) (string, error) {
	return "Hello " + name + h.suffix, nil
}
