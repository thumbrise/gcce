package emit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/op/emit"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

type (
	MockInput  struct{}
	MockOutput struct{}
	MockError  struct{}
)

func (m MockError) Error() string {
	return "mock error"
}

func MockFn(input MockInput) (*MockOutput, *MockError) {
	return &MockOutput{}, nil
}

func TestFunctionToOperation(t *testing.T) {
	expected := schema.Operation{
		ID:      "github.com/thumbrise/gcce/op/emit_test.MockFn",
		Comment: "",
		Input: []schema.Term{
			{
				ID:       "github.com/thumbrise/gcce/op/emit_test.MockInput",
				Comment:  "",
				Required: new(true),
				Kind:     new(schema.KindObject),
				Value:    nil,
				Of:       []schema.Term{},
				Trait:    []schema.Term{},
			},
		},
		Output: []schema.Term{
			{
				ID:       "*github.com/thumbrise/gcce/op/emit_test.MockOutput",
				Comment:  "",
				Required: new(false),
				Kind:     new(schema.KindObject),
				Value:    nil,
				Of:       []schema.Term{},
				Trait:    []schema.Term{},
			},
		},
		Error: []schema.Term{
			{
				ID:       "*github.com/thumbrise/gcce/op/emit_test.MockError",
				Comment:  "",
				Required: new(false),
				Kind:     new(schema.KindObject),
				Value:    nil,
				Of:       []schema.Term{},
				Trait:    []schema.Term{},
			},
		},
		Trait: nil,
	}
	actual, err := emit.FunctionToOperation(MockFn)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
