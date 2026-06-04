package filter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/op/emit/filter"
	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

func TestContext_RemovesAndRenumbers(t *testing.T) {
	ops := []schema.Operation{
		{
			ID: "test",
			Input: []schema.Term{
				{ID: "input0", Trait: []schema.Term{trait.NewFQN("context.Context")}},
				{ID: "input1", Trait: []schema.Term{trait.NewFQN("string")}},
				{ID: "input2", Trait: []schema.Term{trait.NewFQN("context.Context")}},
				{ID: "input3", Trait: []schema.Term{trait.NewFQN("int")}},
			},
		},
	}

	filtered := filter.Context(ops)
	require.Len(t, filtered, 1)
	op := filtered[0]
	assert.Len(t, op.Input, 2)
	assert.Equal(t, "input0", op.Input[0].ID)
	assert.Equal(t, "string", op.Input[0].Trait[0].Value)
	assert.Equal(t, "input1", op.Input[1].ID)
	assert.Equal(t, "int", op.Input[1].Trait[0].Value)
}

func TestContext_NoContextDoesNothing(t *testing.T) {
	ops := []schema.Operation{
		{
			ID: "test",
			Input: []schema.Term{
				{ID: "input0", Trait: []schema.Term{trait.NewFQN("string")}},
			},
		},
	}
	filtered := filter.Context(ops)
	assert.Len(t, filtered[0].Input, 1)
	assert.Equal(t, "input0", filtered[0].Input[0].ID)
}

func TestContext_EmptyInput(t *testing.T) {
	ops := []schema.Operation{{ID: "test"}}
	filtered := filter.Context(ops)
	assert.Empty(t, filtered[0].Input)
}
