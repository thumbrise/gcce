package observable_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thumbrise/gcce/op/emit/pipeline"
	"github.com/thumbrise/gcce/op/emit/pipeline/observable"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

var logger = slog.New(slog.DiscardHandler)

func TestTermPass_Required_NilPointer(t *testing.T) {
	term := schema.Term{ID: "test", Required: nil}
	pass := observable.NewTermPass("input0", logger, nil, term)
	assert.False(t, pass.Required())
}

func TestTermPass_Required_TrueValue(t *testing.T) {
	term := schema.Term{ID: "test", Required: ptr(true)}
	pass := observable.NewTermPass("input0", logger, nil, term)
	assert.True(t, pass.Required())
}

func TestTermPass_Drop_MarksForRemoval(t *testing.T) {
	term := schema.Term{ID: "input0", Required: ptr(true)}
	pass := observable.NewTermPass("input0", logger, nil, term)
	pass.Drop("it's a context")
	assert.True(t, pass.DropMe())
}

func TestTermPass_Rename_ChangesID(t *testing.T) {
	term := schema.Term{ID: "old_name", Required: ptr(true)}
	pass := observable.NewTermPass("input0", logger, nil, term)
	pass.Rename("new_name", "cosmetic")
	assert.Equal(t, "new_name", pass.ID())
}

func TestTermPass_AppendTrait_AddsToSlice(t *testing.T) {
	term := schema.Term{ID: "test", Required: ptr(true)}
	pass := observable.NewTermPass("input0", logger, nil, term)
	trait := schema.Term{ID: "http.method", Value: "POST"}
	pass.AppendTrait(trait, "needed for http")
	assert.Len(t, pass.Traits(), 1)
	assert.Equal(t, "http.method", pass.Traits()[0].ID)
}

func TestTermPass_MapTraits_ProvidesNilReflectType(t *testing.T) {
	term := schema.Term{
		ID: "test",
		Trait: []schema.Term{
			{ID: "http.method", Value: "GET"},
		},
	}
	pass := observable.NewTermPass("input0", logger, nil, term)
	err := pass.MapTraits(func(idx int, traitPass pipeline.TermPass) error {
		assert.Nil(t, traitPass.ReflectType())

		return nil
	})
	assert.NoError(t, err)
}

func ptr[T any](v T) *T {
	return &v
}
