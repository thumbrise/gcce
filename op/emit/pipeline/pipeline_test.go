package pipeline_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/op/emit"
	"github.com/thumbrise/gcce/op/emit/pipeline"
	"github.com/thumbrise/gcce/op/emit/pipeline/contract"
	"github.com/thumbrise/gcce/op/emit/pipeline/pass"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

type mockOperationPlugin struct {
	name    string
	visitFn func(opPass pass.OperationPass) error
}

func (p *mockOperationPlugin) Name() string                          { return p.name }
func (p *mockOperationPlugin) Visit(opPass pass.OperationPass) error { return p.visitFn(opPass) }

type mockInstructionPlugin struct {
	name    string
	visitFn func(instrPass pass.InstructionPass) error
}

func (p *mockInstructionPlugin) Name() string { return p.name }
func (p *mockInstructionPlugin) Visit(instrPass pass.InstructionPass) error {
	return p.visitFn(instrPass)
}

func addGreeting(a string, b string) string {
	return a + b
}

func TestPipeline_Empty(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{
			ID:      "test-empty",
			Version: "v1.0.0",
			Comment: "no operations",
		},
		nil,
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)
	assert.Equal(t, "test-empty", instr.ID)
	assert.Equal(t, "v1.0.0", instr.Version)
	assert.Equal(t, "no operations", instr.Comment)
	assert.Empty(t, instr.Operations)
}

func TestPipeline_SingleOperation(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{
			ID: "test-single",
		},
		[]contract.OperationRegistration{
			{FN: addGreeting},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)

	require.Len(t, instr.Operations, 1)
	op := instr.Operations[0]
	assert.Contains(t, op.ID, "addGreeting")
	assert.Len(t, op.Input, 2)
	assert.Len(t, op.Output, 1)
}

func multiply(a int, b int) int { return a * b }

func TestPipeline_MultipleOperations(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{ID: "test-multi"},
		[]contract.OperationRegistration{
			{FN: addGreeting},
			{FN: multiply},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)
	require.Len(t, instr.Operations, 2)
}

func TestPipeline_NilFunction(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{ID: "test-nil"},
		[]contract.OperationRegistration{
			{FN: nil},
		},
	)
	_, err := ppln.Compile()
	require.Error(t, err)
	assert.ErrorIs(t, err, pipeline.ErrEmitFailure)
}

func TestPipeline_AnonymousFunction(t *testing.T) {
	fn := func() {}
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{ID: "test-anon"},
		[]contract.OperationRegistration{
			{FN: fn},
		},
	)
	_, err := ppln.Compile()
	require.Error(t, err)
	assert.ErrorIs(t, err, emit.ErrIsAnonymous)
}

func TestPipeline_DropByPlugin(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{ID: "test-drop"},
		[]contract.OperationRegistration{
			{
				FN: addGreeting,
				Plugins: []contract.OperationPlugin{
					&mockOperationPlugin{
						name: "dropper",
						visitFn: func(opPass pass.OperationPass) error {
							opPass.Drop("test drop")

							return nil
						},
					},
				},
			},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)
	assert.Empty(t, instr.Operations)
}

func TestPipeline_PluginModifies(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{ID: "test-modify"},
		[]contract.OperationRegistration{
			{
				FN: addGreeting,
				Plugins: []contract.OperationPlugin{
					&mockOperationPlugin{
						name: "modifier",
						visitFn: func(opPass pass.OperationPass) error {
							opPass.SetID("custom-id", "test plugin")
							opPass.SetComment("modified by plugin", "test plugin")

							return nil
						},
					},
				},
			},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)
	require.Len(t, instr.Operations, 1)
	assert.Equal(t, "custom-id", instr.Operations[0].ID)
	assert.Equal(t, "modified by plugin", instr.Operations[0].Comment)
}

func TestPipeline_PluginError(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{ID: "test-plugin-err"},
		[]contract.OperationRegistration{
			{
				FN: addGreeting,
				Plugins: []contract.OperationPlugin{
					&mockOperationPlugin{
						name: "failing",
						visitFn: func(opPass pass.OperationPass) error {
							return errors.New("something went wrong")
						},
					},
				},
			},
		},
	)
	_, err := ppln.Compile()
	require.Error(t, err)
	assert.ErrorIs(t, err, pipeline.ErrPluginFailure)
}

func TestPipeline_InstructionPlugin(t *testing.T) {
	var visited bool

	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{
			ID: "test-instr-plugin",
			Plugins: []contract.InstructionPlugin{
				&mockInstructionPlugin{
					name: "tracker",
					visitFn: func(instrPass pass.InstructionPass) error {
						visited = true

						instrPass.SetComment("plugin was here", "test")

						return nil
					},
				},
			},
		},
		[]contract.OperationRegistration{
			{FN: addGreeting},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)
	assert.True(t, visited)
	assert.Equal(t, "plugin was here", instr.Comment)
}

func TestPipeline_DefaultMetadata(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{},
		[]contract.OperationRegistration{
			{FN: addGreeting},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)
	assert.Equal(t, "unknown", instr.ID)
	assert.Equal(t, "unknown", instr.Comment)
	assert.Equal(t, "v0.0.0-unknown", instr.Version)
}

func TestPipeline_InstructionDropped(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{
			ID: "test-instr-drop",
			Plugins: []contract.InstructionPlugin{
				&mockInstructionPlugin{
					name: "dropper",
					visitFn: func(instrPass pass.InstructionPass) error {
						instrPass.Drop("drop whole instruction")

						return nil
					},
				},
			},
		},
		[]contract.OperationRegistration{
			{FN: addGreeting},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	assert.Nil(t, instr)
}

func pBool(v bool) *bool               { return &v }
func pKind(k schema.Kind) *schema.Kind { return &k }

func testOperationFields(name string, age int) (string, error) {
	return "", nil
}

func TestPipeline_OperationFields(t *testing.T) {
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{ID: "test-fields"},
		[]contract.OperationRegistration{
			{FN: testOperationFields},
		},
	)
	instr, err := ppln.Compile()
	require.NoError(t, err)
	require.NotNil(t, instr)
	require.Len(t, instr.Operations, 1)

	op := instr.Operations[0]
	assert.Contains(t, op.ID, "testOperationFields")

	require.Len(t, op.Input, 2)
	assert.Equal(t, "input0", op.Input[0].ID)
	assert.Equal(t, pBool(true), op.Input[0].Required)
	assert.Equal(t, pKind(schema.KindString), op.Input[0].Kind)

	assert.Equal(t, "input1", op.Input[1].ID)
	assert.Equal(t, pBool(true), op.Input[1].Required)

	require.Len(t, op.Output, 1)
	assert.Equal(t, "output0", op.Output[0].ID)

	require.Len(t, op.Error, 1)
	assert.Equal(t, "error0", op.Error[0].ID)
}
