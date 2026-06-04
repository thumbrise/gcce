package trait

import "github.com/thumbrise/op-universal-schema-go/schema"

const (
	GroupID      = BaseId + "/group"
	GroupComment = "Group the operation belongs to"
)

func NewGroup(groupName string) schema.Term {
	return schema.Term{
		ID:      GroupID,
		Comment: GroupComment,
		Value:   groupName,
	}
}
