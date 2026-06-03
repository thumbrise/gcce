package trait

import "github.com/thumbrise/op-universal-schema-go/schema"

const (
	FQNID      = BaseId + "/fqn"
	FQNComment = "Full qualified name of entity."
)

func NewFQN(value string) schema.Term {
	t := schema.Term{
		ID:      FQNID,
		Comment: FQNComment,
		Value:   value,
	}

	return t
}
