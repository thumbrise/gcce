package emit_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/op/emit"
	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

type (
	MockInput  struct{}
	MockOutput struct{}
	MockError  struct{}

	MockInputSimple  struct{ Name string }
	MockOutputSimple struct{ Result int }
)

func (m MockError) Error() string { return "mock error" }

func MockFn(input MockInput) (*MockOutput, *MockError) {
	return &MockOutput{}, nil
}

func TestFunctionToOperation(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFn",
		Input: []schema.Term{{
			ID:       "input0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait: []schema.Term{
				trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockInput"),
			},
		}},
		Output: []schema.Term{{
			ID:       "output0",
			Required: ptr(false),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait: []schema.Term{
				trait.NewFQN("*github.com/thumbrise/gcce/op/emit_test.MockOutput"),
			},
		}},
		Error: []schema.Term{{
			ID:       "error0",
			Required: ptr(false),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait: []schema.Term{
				trait.NewFQN("*github.com/thumbrise/gcce/op/emit_test.MockError"),
			},
		}},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFn)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func MockFnNoError(in MockInputSimple) MockOutputSimple {
	return MockOutputSimple{}
}

func TestNoErrorFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnNoError",
		Input: []schema.Term{{
			ID:       "input0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait: []schema.Term{
				trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockInputSimple"),
			},
		}},
		Output: []schema.Term{{
			ID:       "output0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait: []schema.Term{
				trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockOutputSimple"),
			},
		}},
		Error: []schema.Term{},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnNoError)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func MockFnMultiInput(a string, b int, c bool) *MockOutputSimple {
	return nil
}

func TestMultiInputFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnMultiInput",
		Input: []schema.Term{
			{
				ID:       "input0",
				Required: ptr(true),
				Kind:     kind(schema.KindString),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("string")},
			},
			{
				ID:       "input1",
				Required: ptr(true),
				Kind:     kind(schema.KindInteger),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("int")},
			},
			{
				ID:       "input2",
				Required: ptr(true),
				Kind:     kind(schema.KindBoolean),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("bool")},
			},
		},
		Output: []schema.Term{{
			ID:       "output0",
			Required: ptr(false),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("*github.com/thumbrise/gcce/op/emit_test.MockOutputSimple")},
		}},
		Error: []schema.Term{},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnMultiInput)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func MockFnMultiOutput(in MockInputSimple) (MockOutputSimple, error) {
	return MockOutputSimple{}, nil
}

func TestMultiOutputFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnMultiOutput",
		Input: []schema.Term{{
			ID:       "input0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockInputSimple")},
		}},
		Output: []schema.Term{{
			ID:       "output0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockOutputSimple")},
		}},
		Error: []schema.Term{{
			ID:       "error0",
			Required: ptr(true),
			Kind:     nil,
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("error")},
		}},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnMultiOutput)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func MockFnOnlyError(in MockInputSimple) error {
	return nil
}

func TestOnlyErrorFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnOnlyError",
		Input: []schema.Term{{
			ID:       "input0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockInputSimple")},
		}},
		Output: []schema.Term{},
		Error: []schema.Term{{
			ID:       "error0",
			Required: ptr(true),
			Kind:     nil,
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("error")},
		}},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnOnlyError)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func MockFnOnlyOutput(in MockInputSimple) MockOutputSimple {
	return MockOutputSimple{}
}

func TestOnlyOutputFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnOnlyOutput",
		Input: []schema.Term{{
			ID:       "input0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockInputSimple")},
		}},
		Output: []schema.Term{{
			ID:       "output0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockOutputSimple")},
		}},
		Error: []schema.Term{},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnOnlyOutput)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func MockFnVariadic(prefix string, nums ...int) MockOutputSimple {
	return MockOutputSimple{}
}

func TestVariadicFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnVariadic",
		Input: []schema.Term{
			{
				ID:       "input0",
				Required: ptr(true),
				Kind:     kind(schema.KindString),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("string")},
			},
			{
				ID:       "input1",
				Required: ptr(true),
				Kind:     kind(schema.KindArray),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("[]int")},
			},
		},
		Output: []schema.Term{{
			ID:       "output0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockOutputSimple")},
		}},
		Error: []schema.Term{},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnVariadic)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func MockFnTimeBytes(t time.Time, data []byte) {}
func TestTimeBytesFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnTimeBytes",
		Input: []schema.Term{
			{
				ID:       "input0",
				Required: ptr(true),
				Kind:     kind(schema.KindDatetime),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("time.Time")},
			},
			{
				ID:       "input1",
				Required: ptr(true),
				Kind:     kind(schema.KindBinary),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("[]uint8")},
			},
		},
		Output: []schema.Term{},
		Error:  []schema.Term{},
		Trait:  []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnTimeBytes)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func MockFnMapSlice(m map[string]int, s []string) {}
func TestMapSliceFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnMapSlice",
		Input: []schema.Term{
			{
				ID:       "input0",
				Required: ptr(true),
				Kind:     kind(schema.KindObject),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("map[string]int")},
			},
			{
				ID:       "input1",
				Required: ptr(true),
				Kind:     kind(schema.KindArray),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("[]string")},
			},
		},
		Output: []schema.Term{},
		Error:  []schema.Term{},
		Trait:  []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnMapSlice)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func MockFnPointerInput(in *MockInputSimple) {}
func TestPointerInputFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnPointerInput",
		Input: []schema.Term{{
			ID:       "input0",
			Required: ptr(false),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("*github.com/thumbrise/gcce/op/emit_test.MockInputSimple")},
		}},
		Output: []schema.Term{},
		Error:  []schema.Term{},
		Trait:  []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnPointerInput)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func MockFnNamedError(in MockInputSimple) *MockError {
	return nil
}

func TestNamedErrorFn(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnNamedError",
		Input: []schema.Term{{
			ID:       "input0",
			Required: ptr(true),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("github.com/thumbrise/gcce/op/emit_test.MockInputSimple")},
		}},
		Output: []schema.Term{},
		Error: []schema.Term{{
			ID:       "error0",
			Required: ptr(false),
			Kind:     kind(schema.KindObject),
			Of:       []schema.Term{},
			Trait:    []schema.Term{trait.NewFQN("*github.com/thumbrise/gcce/op/emit_test.MockError")},
		}},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnNamedError)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestAnonymousFunction(t *testing.T) {
	fn := func() {}
	_, err := emit.FunctionToOperation(fn)
	require.ErrorIs(t, err, emit.ErrIsAnonymous)
}

func TestNonFunction(t *testing.T) {
	_, err := emit.FunctionToOperation("not a func")
	require.ErrorIs(t, err, emit.ErrIsNotFunction)
}

func MockFnContextInput(ctx context.Context, b int, c bool) *MockOutputSimple {
	return nil
}

func TestContextInputAsIt(t *testing.T) {
	expected := schema.Operation{
		ID: "github.com/thumbrise/gcce/op/emit_test.MockFnContextInput",
		Input: []schema.Term{
			{
				ID:       "input0",
				Required: ptr(true),
				Kind:     nil,
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("context.Context")},
			},
			{
				ID:       "input1",
				Required: ptr(true),
				Kind:     kind(schema.KindInteger),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("int")},
			},
			{
				ID:       "input2",
				Required: ptr(true),
				Kind:     kind(schema.KindBoolean),
				Of:       []schema.Term{},
				Trait:    []schema.Term{trait.NewFQN("bool")},
			},
		},
		Output: []schema.Term{
			{
				ID:       "output0",
				Comment:  "",
				Required: ptr(false),
				Kind:     kind(schema.KindObject),
				Value:    nil,
				Of:       []schema.Term{},
				Trait: []schema.Term{
					trait.NewFQN("*github.com/thumbrise/gcce/op/emit_test.MockOutputSimple"),
				},
			},
		},
		Error: []schema.Term{},
		Trait: []schema.Term{},
	}
	actual, err := emit.FunctionToOperation(MockFnContextInput)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func ptr(b bool) *bool                { return &b }
func kind(k schema.Kind) *schema.Kind { return &k }
